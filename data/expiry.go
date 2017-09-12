package data

import (
	"time"
)

// Expiry это время жизни объекта, например ключа
type Expiry struct {
	t time.Duration
}

// NewExpiry возвращает новый Expiry
func NewExpiry(milliseconds uint64) Expiry {
	return Expiry{
		t: time.Duration(milliseconds*1000) * time.Microsecond,
	}
}

// Milliseconds возвращает время жизни в миллисекундах
func (e Expiry) Milliseconds() uint64 {
	return uint64(e.t / (1000 * time.Microsecond))
}

// Seconds возвращает время жизни в секундах
func (e Expiry) Seconds() float64 {
	return e.t.Seconds()
}
