package command

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/avito-tech/smart-redis-replication/data"
)

// Command это структура содержащая команду в исходном виде
type Command struct {
	data []string
}

// New возвращает новую команду
func New(args []string) Command {
	return Command{
		data: args,
	}
}

// Type возвращает тип команды
func (c Command) Type() Type {
	if len(c.data) == 0 {
		return Empty
	}
	command := Type(strings.ToLower(strings.TrimSpace(c.data[0])))
	if command == "" {
		return Empty
	}
	switch command {
	case Ping, Select, Zadd, Sadd, Zrem, Delete, RDB:
		return command
	}
	return Undefined
}

// KeyName возвращает название ключа если оно предусмотрено командой
func (c Command) KeyName() (string, error) {
	if len(c.data) < 2 {
		return "", fmt.Errorf("expected count args >= 2 but actual %d", len(c.data))
	}
	switch c.Type() {
	case Delete, Zrem, Zadd, Sadd:
		return c.data[1], nil
	}
	return "", fmt.Errorf("unexpected type %q", c.Type())
}

// Values возвращает список значений если они предусмотрены командой
func (c Command) Values() ([]string, error) {
	if len(c.data) < 3 {
		return []string{}, fmt.Errorf(
			"expected count args >= 2 but actual %d",
			len(c.data),
		)
	}
	switch c.Type() {
	case Zrem:
		return c.data[2:], nil
	}
	return []string{}, fmt.Errorf("unexpected type %q", c.Type())
}

// ConvertToSelectDB конвертирует команду в номер базы данных
func (c Command) ConvertToSelectDB() (db int, err error) {
	if len(c.data) < 2 {
		return 0, fmt.Errorf("expected count args >=2 but actual %d", len(c.data))
	}
	commandType := c.Type()
	if commandType != Select {
		return 0, fmt.Errorf("expected Select command but actual %s", commandType)
	}
	return strconv.Atoi(strings.TrimSpace(c.data[1]))
}

// ConvertToRDB конвертирует команду в размер RDB
func (c Command) ConvertToRDB() (size int64, err error) {
	if len(c.data) < 2 {
		return 0, fmt.Errorf("expected count args >=2 but actual %d", len(c.data))
	}
	commandType := c.Type()
	if commandType != RDB {
		return 0, fmt.Errorf("expected RDB command but actual %s", commandType)
	}
	return strconv.ParseInt(strings.TrimSpace(c.data[1]), 10, 64)
}

// ConvertToSortedSetKey конвертирует команду Zadd в ключ SortedSetKey
func (c Command) ConvertToSortedSetKey(db int) (data.SortedSetKey, error) {
	if len(c.data) < 4 {
		return nil, fmt.Errorf("expected count args >=3 but actual %d", len(c.data))
	}
	if len(c.data)%2 != 0 {
		return nil, errors.New("expected even count args but actual odd")
	}
	commandType := c.Type()
	if commandType != Zadd {
		return nil, fmt.Errorf("expected Zadd command but actual %s", commandType)
	}
	keyName := c.data[1]
	key := data.NewSortedSet(keyName)
	err := key.SetDB(db)
	if err != nil {
		return nil, err
	}
	count := len(c.data)
	for i := 2; i < count-1; i += 2 {
		score, err := strconv.ParseFloat(c.data[i], 64)
		if err != nil {
			return nil, err
		}
		value := c.data[i+1]
		err = key.Set(score, value)
		if err != nil {
			return nil, err
		}
	}
	return key, nil
}

// ConvertToSetKey конвертирует команду Sadd в ключ SetKey
func (c Command) ConvertToSetKey(db int) (data.SetKey, error) {
	if len(c.data) < 3 {
		return nil, fmt.Errorf("expected count args >=4 but actual %d", len(c.data))
	}
	commandType := c.Type()
	if commandType != Sadd {
		return nil, fmt.Errorf("expected Sadd command but actual %s", commandType)
	}
	keyName := c.data[1]
	key := data.NewSet(keyName)
	err := key.SetDB(db)
	if err != nil {
		return nil, err
	}
	count := len(c.data)
	for i := 2; i < count; i++ {
		err = key.Set(c.data[i])
		if err != nil {
			return nil, err
		}
	}
	return key, nil
}
