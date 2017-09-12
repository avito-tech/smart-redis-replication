package resp

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// reader это реализация Reader
type reader struct {
	*bufio.Reader

	debug bool
	dump  struct {
		name  string
		start bool
		buf   *bytes.Buffer
		dir   string
	}
}

// NewReader возвращает новый Reader
func NewReader(r *bufio.Reader) Reader {
	return newReader(r)
}

// newReader возвращает новый reader
func newReader(r *bufio.Reader) *reader {
	return &reader{
		Reader: r,
	}
}

// NewStringReader возвращает новый Reader на основе строки
func NewStringReader(s string) Reader {
	return newStringReader(s)
}

// newStringReader возвращает новый reader на основе строки
func newStringReader(s string) *reader {
	r := bufio.NewReader(strings.NewReader(s))
	return newReader(r)
}

// NewFileReader возвращает новый Reader на основе файла
func NewFileReader(filename string) (Reader, *os.File, error) {
	return newFileReader(filename)
}

// newFileReader возвращает новый reader на основе файла
func newFileReader(filename string) (*reader, *os.File, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, nil, err
	}
	r := newReader(bufio.NewReader(file))
	return r, file, err
}

// SafeRead безопасно читает N байт
func (r *reader) SafeRead(n uint32) (result []byte, err error) {
	result = make([]byte, n)
	_, err = io.ReadFull(r, result)
	return result, err
}

// ReadOpcode читает код команды
func (r *reader) ReadOpcode() (byte, error) {
	return r.ReadByte()
}

// ReadSimpleString читает простую строку до появления символов \r\n
func (r *reader) ReadSimpleString() (string, error) {
	st, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(st), nil
}

// ReadError читает строку с ошибкой
func (r *reader) ReadError() (string, error) {
	st, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(st), nil
}

// ReadInteger читает целое число
func (r *reader) ReadInteger() (int64, error) {
	st, err := r.ReadString('\n')
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(strings.TrimSpace(st), 10, 64)
}

// ReadBulkString читает бинарно-безопасную строку
func (r *reader) ReadBulkString() (string, error) {
	length, err := r.ReadInteger()
	if err != nil {
		return "", err
	}
	data, err := r.SafeRead(uint32(length))
	if err != nil {
		return "", err
	}
	err = r.ReadCRLF()
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ReadArray читает массив данных, в том числе вложенные
// BulkString, SimpleString, Array, Integer, Error
// nolint:gocyclo
func (r *reader) ReadArray() ([]string, error) {
	length, err := r.ReadInteger()
	if err != nil {
		return nil, err
	}
	if length == 0 {
		return []string{}, nil
	}
	if length == -1 {
		return nil, nil
	}
	var result []string
	for length > 0 {
		length--
		opcode, err := r.ReadOpcode()
		if err != nil {
			return nil, err
		}
		switch opcode {
		case ArrayOpcode:
			data, err := r.ReadArray()
			if err != nil {
				return nil, err
			}
			result = append(result, data...)
		case BulkStringOpcode:
			data, err := r.ReadBulkString()
			if err != nil {
				return nil, err
			}
			result = append(result, data)
		case IntegerOpcode:
			data, err := r.ReadInteger()
			if err != nil {
				return nil, err
			}
			result = append(result, fmt.Sprintf("%d", data))
		case SimpleStringOpcode:
			data, err := r.ReadSimpleString()
			if err != nil {
				return nil, err
			}
			result = append(result, data)
		case ErrorOpcode:
			data, err := r.ReadError()
			if err != nil {
				return nil, err
			}
			result = append(result, data)
		default:
			return nil, fmt.Errorf("unexpected opcode %#v %#v", opcode, ArrayOpcode)
		}
	}
	return result, nil
}

// ReadCRLF читает два байта и проверяет что в них "\r\n"
func (r *reader) ReadCRLF() error {
	crlf, err := r.SafeRead(2)
	if err != nil {
		return err
	}
	if crlf[0] != CR || crlf[1] != LF {
		return fmt.Errorf("expected CRLF but actual %#v", crlf)
	}
	return nil
}
