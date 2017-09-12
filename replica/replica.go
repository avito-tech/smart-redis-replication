package replica

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"

	"go.avito.ru/gl/smart-redis-replication/backlog"
	"go.avito.ru/gl/smart-redis-replication/command"
	"go.avito.ru/gl/smart-redis-replication/resp"
	"go.avito.ru/gl/smart-redis-replication/status"
)

// replica это реализация Replica
type replica struct {
	ctx    context.Context
	cancel context.CancelFunc
	err    error

	conn resp.Conn

	config Config
	status struct {
		// startDecodeRDB означает что декодирование RDB уже началось
		startDecodeRDB bool
	}

	// decoder это объект в котором происходит чтение rdb и backlog
	decoder Decoder

	// backlog это журнал отставания репликации
	backlog *backlog.Backlog

	// consumer это получатель информации из репликации
	consumer Consumer
}

// NewReplica возвращает новую Replica
func NewReplica(
	r io.ReadWriteCloser,
	config Config,
) Replica {
	replica := &replica{
		conn:    resp.New(r),
		config:  config,
		backlog: backlog.New(config.BacklogSize),
	}
	if config.Debug {
		_ = replica.conn.EnableDebug(config.DebugDumpDir)
	}
	replica.ctx, replica.cancel = context.WithCancel(context.Background())
	return replica
}

// Backlog возвращает журнал отставания репликации
func (r *replica) Backlog() *backlog.Backlog {
	return r.backlog
}

// SetCacheRDB устанавливает статус кеширования RDB на диск
// status = true - кешировать
func (r *replica) SetCacheRDB(status bool) {
	r.config.CacheRDB = status
}

// SendStatus отправляет статус репликации
func (r *replica) SendStatus(status status.Status) {
	if r.consumer != nil {
		r.consumer.ReplicaStatus(status) // nolint:errcheck
	}
}

// sendSync отправляет сообщение SYNC необходимое для запуска репликации
func (r *replica) sendSync() error {
	return r.sendRaw("SYNC")
}

// sendRaw отправляет сообщение в сокет и не читает ответ
func (r *replica) sendRaw(data string) error {
	_, err := r.conn.Write([]byte(fmt.Sprintf("%s\r\n", data)))
	return err
}

// Do запускает процесс репликации,
// возвращает ошибку в случае разрыва соединения с сервером,
// метод синхронный
func (r *replica) Do(consumer Consumer) (err error) {
	defer r.Cancel(&err)
	if consumer == nil {
		return errors.New("expected consumer but actual nil")
	}
	r.consumer = consumer

	r.decoder, err = NewDecoder(
		r.backlog,
		r.config,
	)
	if err != nil {
		return err
	}

	r.SendStatus(status.StartSync)
	err = r.sendSync()
	if err != nil {
		return err
	}
	err = r.decode()
	return err
}

// Done возвращает канал для ожидания завершения репликации
func (r *replica) Done() <-chan struct{} {
	return r.ctx.Done()
}

// Status возвращает true если репликация работает и false если нет
func (r *replica) Status() bool {
	select {
	case <-r.ctx.Done():
		return false
	default:
		return true
	}
}

// Err возвращает ошибку декодирования
func (r *replica) Err() error {
	return r.err
}

// Cancel прекращает процесс репликации
func (r *replica) Cancel(err *error) {
	r.err = *err
	if r.decoder != nil {
		r.decoder.Cancel(err)
	}
	if r.consumer != nil {
		r.consumer.Cancel(err)
	}
	r.cancel()
	r.SendStatus(status.StopReplication)
}

func (r *replica) decode() (err error) {
	var cmd command.Command
	for {
		if !r.decoder.Status() {
			return r.decoder.Err()
		}
		cmd, err = r.conn.Command()

		if err != nil {
			return err
		}
		switch cmd.Type() {
		case command.Empty:
			continue
		case command.RDB:
			r.consumer.ReplicaStatus(status.StartCacheRDB)

			if r.status.startDecodeRDB {
				return errors.New("unexpected RDB command")
			}
			r.status.startDecodeRDB = true

			err = r.cacheRDB(cmd)
			r.consumer.ReplicaStatus(status.StopCacheRDB)
			if err == nil {
				r.decoder.SetFileRDB(r.config.CacheRDBFile) //nolint:errcheck
				r.startDecoder()
			}
		default:
			if r.consumer.CheckCommand(cmd) {
				err = r.backlog.Add(cmd)
			}
		}
		if err != nil {
			return err
		}
	}
}

// createRDBDir создаёт директорию для хренения RDB кеша
func (r *replica) createRDBDir() error {
	dir := path.Dir(r.config.CacheRDBFile)
	if !resp.IsDir(dir) {
		err := os.MkdirAll(dir, os.ModeDir)
		if err != nil {
			return err
		}
	}
	return nil
}

// cacheRDB сохраняет RDB в файл, предварительно удаляя старый кеш
func (r *replica) cacheRDB(cmd command.Command) error {
	size, err := cmd.ConvertToRDB()
	if err != nil {
		return err
	}
	err = r.createRDBDir()
	if err != nil {
		return err
	}
	err = r.deleteCacheRDB()
	if err != nil {
		return err
	}
	file, err := os.Create(r.config.CacheRDBFile)
	if err != nil {
		return err
	}
	defer file.Close() // nolint:errcheck

	n, err := io.CopyN(file, r.conn, size)
	if n != size {
		return fmt.Errorf(
			"save rdb cache error: expected %d byte but actual %d byte save",
			size,
			n,
		)
	}
	return err
}

// deleteCacheRDB удаляет кеш RDB
func (r *replica) deleteCacheRDB() error {
	if _, err := os.Stat(r.config.CacheRDBFile); err == nil {
		return os.Remove(r.config.CacheRDBFile)
	}
	return nil
}

// startDecoder запускает обработку RDB и Backlog
func (r *replica) startDecoder() {
	go r.decoder.Decode(r.consumer) // nolint:errcheck
}
