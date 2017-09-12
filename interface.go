package replica

import (
	"go.avito.ru/gl/smart-redis-replication/replica"
)

// Client это интерфейс соединения с redis сервером
type Client interface {
	// NewReplica возвращает клиент для чтения логической репликации
	NewReplica() (replica.Replica, error)

	// Send отправляет простые команды на сервер
	Send(commandName string, args ...interface{}) error

	// Close закрывает сетевое соединение
	Close() error
}

// Replica это проброс интерфейса Replica из подпакета на уровень выше
type Replica interface {
	replica.Replica
}
