package main

import (
	"fmt"
	"log"
	"net"
	"time"

	srr "go.avito.ru/gl/smart-redis-replication"
	"go.avito.ru/gl/smart-redis-replication/status"
)

// Migration это обёртка для запуска репликации
type Migration struct {
	Config   Config
	Consumer *Consumer
}

// NewMigration возвращает новый Migration
func NewMigration(
	conf Config,
) (
	m *Migration,
	err error,
) {
	m = new(Migration)
	m.Config = conf
	m.Consumer, err = NewConsumer(*conf.service.address)
	if err != nil {
		return nil, err
	}
	m.Config = conf
	m.Consumer.Config = conf
	return m, nil
}

// ReloadConfig могла бы перечитать конфиг,
// но по факту нам нужно включить чтение RDB
func (m *Migration) ReloadConfig() {
	m.Config.replica.ReadRDB = true
}

// Start запускает миграцию, перезапускает в случае потери соединения
func (m *Migration) Start() error {
	for {
		m.Connect() // nolint:errcheck
		m.ReloadConfig()
	}
}

// SendStatus отправляет статус миграции в лог
func (m *Migration) SendStatus(status status.Status, err error) {
	_ = err
	err = m.Consumer.ReplicaStatus(status)
	_ = err
}

// Connect подключается к редису и запускает миграцию
func (m *Migration) Connect() error {
	conn, err := srr.NewConnect(*m.Config.redis.host, *m.Config.redis.port, -1)
	if err != nil {
		return fmt.Errorf("connect error: %v", err)
	}
	repl, err := conn.NewReplica(m.Config.replica)
	if err != nil {
		return fmt.Errorf("create replica error: %v", err)
	}

	go m.Statistics(repl)

	m.SendStatus(status.Connect, nil)
	err = repl.Do(m.Consumer)
	m.SendStatus(status.Reconnect, err)
	return err
}

// Statistics отправляет размер журнала отставания репликации в мониторинг
func (m *Migration) Statistics(repl srr.Replica) {
	backlog := repl.Backlog()
	if backlog == nil {
		log.Println("statistics off")
		return
	}
	log.Println("statistics on")
	for {
		select {
		case <-repl.Done():
			return
		case <-time.After(1 * time.Minute):
		}
		go m.backlogSize(*m.Config.redis.port, backlog.Count())
	}
}

// backlogSize отправляет размер очереди в мониторинг
func (m *Migration) backlogSize(port int, count int) {
	t := time.Now().Local().Unix()
	data := fmt.Sprintf("migration.backlogsize.%d %d %d\n", port, count, t)

	for {
		conn, err := net.Dial("tcp", "graphite")
		if err == nil {
			_, err := conn.Write([]byte(data))
			_ = conn.Close()
			if err == nil {
				break
			}
		}
		time.Sleep(1 * time.Second)
	}
}
