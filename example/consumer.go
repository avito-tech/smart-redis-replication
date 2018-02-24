package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/avito-tech/smart-redis-replication/command"
	"github.com/avito-tech/smart-redis-replication/data"
	"github.com/avito-tech/smart-redis-replication/status"
)

// Consumer это получатель информации из репликации
type Consumer struct {
	ctx     context.Context
	cancel  context.CancelFunc
	err     error
	db      int
	service struct {
		address string
		client  *http.Client
	}
	regexp struct {
		keyName *regexp.Regexp
		repl    string
	}
	Config Config
}

// NewConsumer возвращает новый Consumer
func NewConsumer(address string) (c *Consumer, err error) {
	c = new(Consumer)
	c.regexp.repl = "$1"
	c.regexp.keyName, err = regexp.Compile("prefix:(.*)")
	if err != nil {
		return nil, err
	}
	c.service.address = address
	c.service.client = &http.Client{
		Timeout: 6,
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 1000,
		},
	}
	c.ctx, c.cancel = context.WithCancel(context.Background())
	return c, nil
}

// Cancel останавливает обработку данных
func (c *Consumer) Cancel(err *error) {
	c.err = *err
	c.cancel()
}

// Reset сбрасывает контекст
func (c *Consumer) Reset() {
	c.ctx, c.cancel = context.WithCancel(context.Background())
}

// Status возвращает true если репликация не прекратилась, иначе false
func (c *Consumer) Status() bool {
	select {
	case <-c.ctx.Done():
		return false
	default:
		return true
	}
}

// Err возвращает ошибку репликации
func (c *Consumer) Err() error {
	return c.err
}

// CheckKeyName проверяет принадлежность ключа к регулярному выражению
func (c *Consumer) CheckKeyName(keyName string) bool {
	return c.regexp.keyName.MatchString(keyName)
}

// ClearKeyName убирает префикс ключа
func (c *Consumer) ClearKeyName(keyName string) string {
	return string(c.regexp.keyName.ReplaceAll(
		[]byte(keyName),
		[]byte(c.regexp.repl),
	))
}

// Key принимает ключи с данными
func (c *Consumer) Key(key data.Key) error {
	// проверяем префикс ключа
	if !c.CheckKeyName(key.Name()) {
		return nil
	}

	// проверяем в том ли формате мы ожидаем данные
	sortedSet, ok := key.(data.SortedSetKey)
	if !ok {
		return nil
	}

	item := struct {
		keyName string
	}{
		keyName: sortedSet.Name(),
	}

	return c.Send("_PATH_", item)
}

// CheckCommand возвращает true если команда интересна получателю
func (c *Consumer) CheckCommand(cmd command.Command) bool {
	switch cmd.Type() {
	case command.Select:
		return true
	case command.Delete, command.Zrem, command.Sadd, command.Zadd:
		keyName, err := cmd.KeyName()
		if err != nil {
			return false
		}
		// проверяем префикс ключа
		if !c.CheckKeyName(keyName) {
			return false
		}
		return true
	default:
		return false
	}
}

// Command принимает команды управления
func (c *Consumer) Command(cmd command.Command) (err error) {
	switch cmd.Type() {
	case command.Select:
		c.db, err = cmd.ConvertToSelectDB()
	case command.Delete:
		err = c.DeleteKey(cmd)
	case command.Zrem:
		err = c.DeleteItemIDs(cmd)
	default:
		return nil
	}
	return err
}

// DeleteKey обрабатывает удаление ключа
func (c *Consumer) DeleteKey(cmd command.Command) error {
	//	c.Send()
	return nil
}

// DeleteItemIDs удаляет значения из SortedSet
func (c *Consumer) DeleteItemIDs(cmd command.Command) error {
	//	c.Send()
	return nil
}

// Send отправляет данные в сервис
func (c *Consumer) Send(url string, data interface{}) error {
	body, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("post body marshal error: %v", err)
	}
	for {
		select {
		case <-c.ctx.Done():
			return c.ctx.Err()
		default:
		}

		r, err := http.NewRequest(
			"POST",
			c.service.address+url,
			bytes.NewBufferString(string(body)),
		)
		if err != nil {
			return fmt.Errorf("create request error: %q", err)
		}
		r.Header.Set("HTTP_CONNECTION", "keep-alive")
		r.Header.Set("Content-Length", strconv.Itoa(len(string(body))))

		var resp *http.Response
		resp, err = c.service.client.Do(r)
		if resp != nil && resp.Body != nil {
			resp.Body.Close() //nolint:errcheck
		}
		if err == nil && resp.StatusCode == 200 {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// status отправляет отсечку на графики
func (c *Consumer) status(port int, status string) {
	t := time.Now().Local().Unix()
	n := (port - 6300) * 10
	buf := fmt.Sprintf("migration.status.%d.%s %d %d\n", port, status, n, t)

	for {
		conn, err := net.Dial("tcp", "graphite")
		if err == nil {
			_, err := conn.Write([]byte(buf))
			_ = conn.Close()
			if err == nil {
				break
			}
		}
		time.Sleep(1 * time.Second)
	}
}

// ReplicaStatus принимает статус репликации и отправляет его в лог
func (c *Consumer) ReplicaStatus(status status.Status) error {
	go c.status(*c.Config.redis.port, string(status))
	return nil
}
