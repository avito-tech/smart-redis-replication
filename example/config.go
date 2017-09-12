package main

import (
	"errors"

	"github.com/avito-tech/smart-redis-replication/replica"
)

// Config это минимальный набор данных для запуска мигратора
type Config struct {
	useStderr  *bool
	useMetrics *bool
	redis      struct {
		host *string
		port *int
	}
	service struct {
		address *string
	}
	replica replica.Config
}

// Check проверяет что в конфиге все необходимые данные
func (c Config) Check() error {
	switch {
	case *c.redis.host == "":
		return errors.New("empty redis host")
	case *c.redis.port < 1:
		return errors.New("expected redis port > 0")
	case *c.service.address == "":
		return errors.New("empty service address")
	}
	return nil
}
