package replica

import (
	"io"

	"go.avito.ru/gl/smart-redis-replication/backlog"
	"go.avito.ru/gl/smart-redis-replication/command"
	"go.avito.ru/gl/smart-redis-replication/data"
	"go.avito.ru/gl/smart-redis-replication/rdb"
	"go.avito.ru/gl/smart-redis-replication/status"
)

// Consumer это получатель информации из репликации
type Consumer interface {
	// Key принимает ключи с данными
	Key(data.Key) error

	// Command принимает команды управления
	Command(command command.Command) error

	// CheckCommand возвращает true если команда интересна получателю
	CheckCommand(command command.Command) bool

	// Status принимает статус репликации
	ReplicaStatus(status.Status) error

	// Cancel останавливает обработку данных
	Cancel(err *error)
}

// Replica это интерфейс репликации
type Replica interface {
	// Done возвращает канал для ожидания завершения репликации
	Done() <-chan struct{}

	// Status возвращает true если репликация выполняется и false если нет
	Status() bool

	// Err возвращает ошибку репликации
	Err() error

	// Возвращает журнал отставания репликации
	Backlog() *backlog.Backlog

	// Do запускает процесс репликации
	Do(consumer Consumer) error
}

// Decoder это интерфейс декодера который наполняет Consumer
type Decoder interface {
	SetRDB(r io.Reader) error

	SetFileRDB(filename string) error

	SetRDBDecoder(rdb.Decoder) error

	// Возвращает журнал отставания репликации
	Backlog() *backlog.Backlog

	// Done возвращает канал для ожидания завершения декодера
	Done() <-chan struct{}

	// Status возвращает true если декодирование выполняется и false если нет
	Status() bool

	// Err возвращает ошибку выполнения
	Err() error

	// Decode запускает процес наполнения Consumer данными
	Decode(consumer Consumer) error

	// Cancel останавливает обработку данных
	Cancel(err *error)
}
