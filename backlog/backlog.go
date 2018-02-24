package backlog

import (
	"errors"
	"sync"

	"github.com/avito-tech/smart-redis-replication/command"
)

const (
	// DefaultBacklogSize это размер backlog по умолчанию
	DefaultBacklogSize = 50000000
)

// Backlog это журнал отставания репликации
type Backlog struct {
	sync.RWMutex

	// data это канал с командами
	data chan command.Command

	// size это максимальное количество элементов которые может хранить backlog
	size int
}

// New возвращает новый Backlog
func New(size int) *Backlog {
	b := new(Backlog)
	b.data = make(chan command.Command, size)
	b.size = size
	return b
}

// Add добавляет в backlog команду
func (b *Backlog) Add(command command.Command) error {
	b.Lock()
	defer b.Unlock()

	if len(b.data) == b.size {
		return errors.New("queue size exceeded")
	}

	b.data <- command
	return nil
}

// Get возвращает команду из backlog
func (b *Backlog) Get() command.Command {
	return <-b.data
}

// Count возвращает количество команд в backlog
func (b *Backlog) Count() int {
	return len(b.data)
}
