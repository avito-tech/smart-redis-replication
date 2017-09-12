package resp

import (
	"bufio"
	"io"
)

// conn это коннект к redis по протоколу RESP
type conn struct {
	close io.Closer
	Reader
	Writer
}

// New возвращает новый RESP коннект
func New(r io.ReadWriteCloser) Conn {
	return &conn{
		close:  r,
		Reader: NewReader(bufio.NewReader(r)),
		Writer: NewWriter(r),
	}
}

// Close закрывает соединение
func (c *conn) Close() error {
	return c.close.Close()
}
