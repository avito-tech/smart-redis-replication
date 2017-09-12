package resp

import (
	"fmt"
	"io"
	"strings"
)

// Array это массив разноплановых данных
type Array []interface{}

// writer это реализация Writer
type writer struct {
	io.Writer
}

// NewWriter возвращает новый Writer
func NewWriter(w io.Writer) Writer {
	return &writer{
		Writer: w,
	}
}

// WriteByte записывает один байт
func (w *writer) WriteByte(b byte) error {
	_, err := w.Write([]byte{b})
	if err != nil {
		return fmt.Errorf("error write byte: %v", err)
	}
	return nil
}

// WriteOpcode записывает opcode
func (w *writer) WriteOpcode(opcode byte) error {
	_, err := w.Write([]byte{opcode})
	if err != nil {
		return fmt.Errorf("error write opcode: %v", err)
	}
	return nil
}

// WriteSimpleString записывает простую строку вместе с opcode
func (w *writer) WriteSimpleString(st string) error {
	err := w.WriteOpcode(SimpleStringOpcode)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(fmt.Sprintf("%s\r\n", strings.TrimSpace(st))))
	return err
}

// WriteError записывает строку с ошибкой
func (w *writer) WriteError(errData error) error {
	if errData == nil {
		return fmt.Errorf("expected error but actual nil")
	}
	err := w.WriteOpcode(ErrorOpcode)
	if err != nil {
		return err
	}
	_, err = w.Write(
		[]byte(fmt.Sprintf(
			"%s\r\n",
			strings.TrimSpace(errData.Error()),
		)),
	)
	return err
}

// WriteInteger записывает целое число
func (w *writer) WriteInteger(n int64) error {
	err := w.WriteOpcode(IntegerOpcode)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(fmt.Sprintf("%d\r\n", n)))
	return err
}

// WriteBulkString записывает бинарно-безопасную строку
func (w *writer) WriteBulkString(data []byte) error {
	err := w.WriteOpcode(BulkStringOpcode)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(fmt.Sprintf("%d\r\n", len(data))))
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

// WriterArray записывает массив бинарно-безопасных строк
// nolint:gocyclo
func (w *writer) WriteArray(data []interface{}) error {
	err := w.WriteOpcode(ArrayOpcode)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(fmt.Sprintf("%d\r\n", len(data))))
	if err != nil {
		return err
	}

	for _, item := range data {
		switch op := item.(type) {
		case string:
			err = w.WriteSimpleString(op)
		case int64:
			err = w.WriteInteger(op)
		case Array:
			err = w.WriteArray(op)
		case []byte:
			err = w.WriteBulkString(op)
		case error:
			err = w.WriteError(op)
		default:
			return fmt.Errorf("unexpected type %#v", op)
		}
		if err != nil {
			return err
		}
	}
	return nil
}
