package rdb

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

type intSetReader struct {
	*bufio.Reader
}

// NewIntSetReader возвращает новый IntSetReader
func NewIntSetReader(r io.Reader) IntSetReader {
	return &intSetReader{
		Reader: bufio.NewReaderSize(r, DefaultReaderSize),
	}
}

// NewIntSetStringReader возвращает новый IntSetReader
func NewIntSetStringReader(st string) IntSetReader {
	r := bytes.NewBufferString(st)
	return &intSetReader{Reader: bufio.NewReaderSize(r, DefaultReaderSize)}
}

// SafeRead безопасно читает N байт
func (r *intSetReader) SafeRead(n uint32) ([]byte, error) {
	result := make([]byte, n)
	_, err := io.ReadFull(r.Reader, result)
	return result, err
}

// ReadEntryLength читает размер элементов
func (r *intSetReader) ReadEntryLength() (uint32, error) {
	bytesCount, err := r.SafeRead(4)
	if err != nil {
		return 0, err
	}
	length := binary.LittleEndian.Uint32(bytesCount)

	switch length {
	case 2, 4, 8:
		return length, nil
	}
	return 0, fmt.Errorf("unexpected intset encoding: %d", length)
}

// ReadEntriesCount читает количество элементов
func (r *intSetReader) ReadEntriesCount() (uint32, error) {
	count, err := r.SafeRead(4)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(count), nil
}

// ReadEntry64 читает элемент длиной в 8 байт
func (r *intSetReader) ReadEntry64() (uint64, error) {
	entry, err := r.SafeRead(8)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint64(entry), nil
}

// ReadEntry32 читает элемент длиной в 4 байта
func (r *intSetReader) ReadEntry32() (uint32, error) {
	entry, err := r.SafeRead(4)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(entry), nil
}

// ReadEntry32 читает элемент длиной в 2 байта
func (r *intSetReader) ReadEntry16() (uint16, error) {
	entry, err := r.SafeRead(2)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint16(entry), nil
}
