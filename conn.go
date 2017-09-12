package replica

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"

	"go.avito.ru/gl/smart-redis-replication/replica"
)

// Conn это постоянное соединение с redis сервером
type Conn struct {
	sync.Mutex
	replication bool

	conn io.ReadWriteCloser
}

// NewConnect возвращает новый Conn
func NewConnect(host string, port int, db int) (*Conn, error) {
	if host == "" {
		return nil, fmt.Errorf("expected host")
	}
	if port <= 0 {
		return nil, fmt.Errorf("expected port > 0")
	}
	if db < -1 {
		return nil, fmt.Errorf("expected db > -2")
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return nil, err
	}
	redisConn, err := NewConn(conn)
	if err != nil {
		return nil, err
	}
	if db > -1 {
		err = redisConn.Send("SELECT", db)
		if err != nil {
			err = redisConn.Close()
			return nil, err
		}
	}
	return redisConn, nil
}

// NewConn возвращает новый Conn
func NewConn(conn io.ReadWriteCloser) (*Conn, error) {
	if conn == nil {
		return nil, fmt.Errorf("expected conn io.ReadWriteCloser")
	}
	return &Conn{
		conn: conn,
	}, nil
}

func (c *Conn) replicationLock() {
	c.Lock()
	defer c.Unlock()
	c.replication = true
}

// NewReplica возвращает новый Replica, не переводит коннект в режим репликации
func (c *Conn) NewReplica(config replica.Config) (replica.Replica, error) {
	c.replicationLock()
	return replica.NewReplica(c.conn, config), nil
}

// Send отправляет комманду и не читает ответ
func (c *Conn) Send(commandName string, args ...interface{}) error {
	c.Lock()
	defer c.Unlock()
	if c.replication {
		return fmt.Errorf("error send: replication mode is enabled")
	}

	return c.send(commandName, args...)
}

// Close закрывает соединение
func (c *Conn) Close() error {
	err := c.conn.Close()
	return err
}

// send отправляет комманду с аргументами в сокет
func (c *Conn) send(commandName string, args ...interface{}) error {
	return c.writeCommand(commandName, args...)
}

// writeCommand формирует и записывает комманду непосредственно в сокет
// nolint:gocyclo
func (c *Conn) writeCommand(
	commandName string,
	args ...interface{},
) (
	err error,
) {
	err = c.writeLen('*', len(args)+1)
	if err != nil {
		return err
	}
	_, err = c.conn.Write([]byte(commandName))
	if err != nil {
		return err
	}

	for _, arg := range args {
		if err != nil {
			break
		}
		switch arg := arg.(type) {
		case string:
			err = c.writeString(arg)
		case []byte:
			err = c.writeBytes(arg)
		case int:
			err = c.writeInt64(int64(arg))
		case int64:
			err = c.writeInt64(arg)
		case float64:
			err = c.writeFloat64(arg)
		case bool:
			if arg {
				err = c.writeString("1")
			} else {
				err = c.writeString("0")
			}
		case nil:
			err = c.writeString("")
		default:
			var buf bytes.Buffer
			fmt.Fprint(&buf, arg)
			err = c.writeBytes(buf.Bytes())
		}
	}
	return err
}

func (c *Conn) writeLen(prefix byte, n int) error {
	_, err := c.conn.Write([]byte(fmt.Sprintf("%s%d\r\n", string(prefix), n)))
	return err
}

func (c *Conn) writeString(s string) error {
	err := c.writeLen('$', len(s))
	if err != nil {
		return err
	}
	_, err = c.conn.Write([]byte(fmt.Sprintf("%s\r\n", s)))
	return err
}

func (c *Conn) writeInt64(n int64) error {
	return c.writeBytes(strconv.AppendInt([]byte{}, n, 10))
}

func (c *Conn) writeFloat64(n float64) error {
	return c.writeBytes(strconv.AppendFloat([]byte{}, n, 'g', -1, 64))
}

func (c *Conn) writeBytes(p []byte) error {
	err := c.writeLen('$', len(p))
	if err != nil {
		return err
	}
	_, err = c.conn.Write(p)
	if err != nil {
		return err
	}
	_, err = c.conn.Write([]byte("\r\n"))
	return err
}
