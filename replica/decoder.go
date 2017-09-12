package replica

import (
	"context"
	"errors"
	"io"
	"os"

	"github.com/avito-tech/smart-redis-replication/backlog"
	"github.com/avito-tech/smart-redis-replication/command"
	"github.com/avito-tech/smart-redis-replication/data"
	"github.com/avito-tech/smart-redis-replication/rdb"
	"github.com/avito-tech/smart-redis-replication/status"
)

type decoder struct {
	ctx      context.Context
	cancel   context.CancelFunc
	err      error
	rdb      rdb.Decoder
	file     *os.File
	backlog  *backlog.Backlog
	consumer Consumer
	config   Config
}

// NewDecoder возвращает новый Decoder
func NewDecoder(
	backlog *backlog.Backlog,
	config Config,
) (
	Decoder,
	error,
) {
	if backlog == nil {
		return nil, errors.New("expected backlog")
	}
	d := &decoder{
		backlog: backlog,
		config: config,
	}
	d.ctx, d.cancel = context.WithCancel(context.Background())
	return d, nil
}

// Backlog возвращает журнал отставания репликации
func (d *decoder) Backlog() *backlog.Backlog {
	return d.backlog
}

// SetRDB устанавливает источник RDB
func (d *decoder) SetRDB(r io.Reader) error {
	if r == nil {
		return errors.New("expected io.Reader but actual nil")
	}
	d.rdb = rdb.NewDecoder(r)
	return nil
}

// SetFileRDB открывает файл в качестве источника RDB
func (d *decoder) SetFileRDB(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return nil
	}
	d.file = file
	return d.SetRDB(file)
}

// SetRDBDecoder устанавливает RDB decoder
func (d *decoder) SetRDBDecoder(dec rdb.Decoder) error {
	if dec == nil {
		return errors.New("expected rdb.Decoder but actual nil")
	}
	d.rdb = dec
	return nil
}

// Done возвращает канал для ожидания завершения декодера
func (d *decoder) Done() <-chan struct{} {
	return d.ctx.Done()
}

// Status возвращает true если декодирование выполняется и false если нет
func (d *decoder) Status() bool {
	select {
	case <-d.ctx.Done():
		return false
	default:
		return true
	}
}

// Err возвращает ошибку декодирования
func (d *decoder) Err() error {
	return d.err
}

// Cancel останавливает декодирование
func (d *decoder) Cancel(err *error) {
	d.err = *err
	if d.consumer != nil {
		d.consumer.Cancel(err)
	}
	d.cancel()
	if d.consumer != nil {
		d.consumer.ReplicaStatus(status.StopDecoder) // nolint:errcheck
	}
}

// Do запускает декодирование RDB и Backlog
func (d *decoder) Decode(consumer Consumer) (err error) {
	defer d.Cancel(&err)
	if consumer == nil {
		return errors.New("expected consumer but actual nil")
	}
	d.consumer = consumer
	if d.rdb == nil {
		return errors.New("empty rdb.Decoder")
	}
	if d.backlog == nil {
		return errors.New("empty backlog")
	}
	
	if d.config.ReadRDB {
		consumer.ReplicaStatus(status.StartReadRDB) // nolint:errcheck
		err = d.decodeRDB(consumer)
		consumer.ReplicaStatus(status.StopReadRDB) // nolint:errcheck
		if err != nil {
			return err
		}
	} else {
		consumer.ReplicaStatus(status.SkipReadRDB) // nolint:errcheck
	}
	consumer.ReplicaStatus(status.StartReadBacklog) // nolint:errcheck
	err = d.decodeBacklog(consumer)
	return err
}

// decodeRDB декодирует RDB
func (d *decoder) decodeRDB(consumer Consumer) error {
	if d.file != nil {
		defer d.file.Close() // nolint:errcheck
	}
	if consumer == nil {
		return errors.New("empty consumer")
	}
	// поидеи сюда тоже надо прокинуть контекст
	return d.rdb.DecodeKeys(consumer)
}

// decodeBacklog декодирует Backlog
// nolint:gocyclo
func (d *decoder) decodeBacklog(consumer Consumer) error {
	if d.rdb == nil {
		return errors.New("empty rdb.Decoder")
	}
	if d.backlog == nil {
		return errors.New("empty backlog")
	}
	if consumer == nil {
		return errors.New("empty consumer")
	}
	var db int
	var key data.Key
	var err error
	for {
		if !d.Status() {
			return d.Err()
		}
		cmd := d.backlog.Get()
		switch cmd.Type() {
		case command.Select:
			db, err = cmd.ConvertToSelectDB()
		case command.Zadd:
			key, err = cmd.ConvertToSortedSetKey(db)
			if err != nil {
				return err
			}
			err = consumer.Key(key)
		case command.Sadd:
			key, err = cmd.ConvertToSetKey(db)
			if err != nil {
				return err
			}
			err = consumer.Key(key)
		default:
			err = consumer.Command(cmd)
		}
		if err != nil {
			return err
		}
	}
}
