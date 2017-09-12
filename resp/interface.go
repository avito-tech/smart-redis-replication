package resp

import (
	"io"

	"github.com/avito-tech/smart-redis-replication/command"
)

const (
	// SimpleStringOpcode это первый байт простого однострочного ответа ("+")
	SimpleStringOpcode = 0x2b

	// ErrorOpcode это первый байт ответа ошибки ("-")
	ErrorOpcode = 0x2d

	// IntegerOpcode это первый байт целого числа (":")
	IntegerOpcode = 0x3a

	// BulkStringOpcode это первый байт бинарных данных ("$")
	BulkStringOpcode = 0x24

	// ArrayOpcode это первый байт массива ("*")
	ArrayOpcode = 0x2a

	// CR это символ \r
	CR = 0xd

	// LF это символ \n
	LF = 0xa
)

// Reader это интерфейс для чтения комманд из потока
type Reader interface {
	io.Reader
	Command() (command.Command, error)

	//	ReadByte() (byte, error)
	//	ReadOpcode() (byte, error)
	ReadString(delim byte) (string, error)

	//	ReadSimpleString() (string, error)
	//	ReadError() (string, error)
	//	ReadInteger() (int64, error)

	//	ReadBulkString() (string, error)
	//	ReadArray() ([]string, error)

	// EnableDebug включает отладку
	// В dir сохраняются команды в бинарном виде по одному файлу на команду
	EnableDebug(dir string) error

	// DisableDebug выключает отладку
	DisableDebug()
}

// Writer это интерфейс для записи комманд в поток
type Writer interface {
	io.Writer
	WriteByte(byte) error
	WriteOpcode(byte) error
	WriteSimpleString(string) error
	WriteError(error) error
	WriteInteger(int64) error
	WriteBulkString([]byte) error
	WriteArray([]interface{}) error
}

// Closer это интерфейс для закрытия соединения
type Closer interface {
	io.Closer
}

// Conn это интерфейс для чтения/записи в поток
type Conn interface {
	Reader
	Writer
	Closer
}
