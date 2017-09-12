package rdb

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strconv"
)

const (
	zipListInt16 = 0xC0
	zipListInt32 = 0xD0
	zipListInt64 = 0xE0
	zipListInt24 = 0xF0
	zipListInt8  = 0xFE
	zipListInt4  = 15
)

type entriesReader struct {
	*bufio.Reader
}

// NewEntriesReader возвращает новый EntriesReader
func NewEntriesReader(r io.Reader) EntriesReader {
	return &entriesReader{
		Reader: bufio.NewReaderSize(r, DefaultReaderSize),
	}
}

// NewEntriesStringReader возвращает новый EntriesReader
func NewEntriesStringReader(st string) EntriesReader {
	r := bytes.NewBufferString(st)
	return &entriesReader{Reader: bufio.NewReaderSize(r, DefaultReaderSize)}
}

// SafeRead безопасно читает N байт
func (r *entriesReader) SafeRead(n uint32) ([]byte, error) {
	result := make([]byte, n)
	_, err := io.ReadFull(r.Reader, result)
	return result, err
}

// ReadEntryLength читает размер элементов
func (r *entriesReader) ReadEntryLength() (uint32, error) {
	bytesCount, err := r.SafeRead(4)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(bytesCount), nil
}

// ReadTail читает смещение последнего элемента
func (r *entriesReader) ReadTail() (uint32, error) {
	tail, err := r.SafeRead(4)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(tail), nil
}

// ReadEntriesCount читает количество элементов
func (r *entriesReader) ReadEntriesCount() (uint16, error) {
	lenBytes, err := r.SafeRead(2)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint16(lenBytes), nil
}

// ReadEntryLength читает размер предыдущего элемента
// если первый байт меньше 254 то в нём размер,
// если 254 то размер в следующих 4 байтах
func (r *entriesReader) ReadPrevEntryLength() (uint32, error) {
	prevEntryLength, err := r.ReadByte()
	if err != nil {
		return 0, err
	}
	switch {
	case prevEntryLength < 254:
		return uint32(prevEntryLength), nil
	case prevEntryLength == 254:
		lenBytes, err := r.SafeRead(4)
		if err != nil {
			return 0, err
		}
		return binary.LittleEndian.Uint32(lenBytes), nil
	}
	return 0, fmt.Errorf("unexpected prevEntryLength %#v", prevEntryLength)
}

// ReadEntry читает элемент
// nolint:gocyclo
func (r *entriesReader) ReadEntry() ([]byte, error) {
	_, err := r.ReadPrevEntryLength()
	if err != nil {
		return nil, err
	}

	header, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	switch {
	case header>>6 == len6Bit:
		return r.SafeRead(uint32(header & 0x3f))
	case header>>6 == len14Bit:
		b, err := r.ReadByte()
		if err != nil {
			return nil, err
		}
		return r.SafeRead((uint32(header&0x3f) << 8) | uint32(b))
	case header>>6 == len32Bit:
		lenBytes, err := r.SafeRead(4)
		if err != nil {
			return nil, err
		}
		return r.SafeRead(binary.BigEndian.Uint32(lenBytes))
	case header == zipListInt16:
		intBytes, err := r.SafeRead(2)
		if err != nil {
			return nil, err
		}
		return []byte(
			strconv.FormatInt(
				int64(binary.LittleEndian.Uint16(intBytes)),
				10,
			),
		), nil
	case header == zipListInt32:
		intBytes, err := r.SafeRead(4)
		if err != nil {
			return nil, err
		}
		return []byte(
			strconv.FormatInt(
				int64(binary.LittleEndian.Uint32(intBytes)),
				10,
			),
		), nil
	case header == zipListInt64:
		intBytes, err := r.SafeRead(8)
		if err != nil {
			return nil, err
		}
		return []byte(
			strconv.FormatInt(
				int64(binary.LittleEndian.Uint64(intBytes)),
				10,
			),
		), nil
	case header == zipListInt24:
		intBytes, err := r.SafeRead(3)
		if err != nil {
			return nil, err
		}
		intBytes = append([]byte{0x00}, intBytes...)
		return []byte(
			strconv.FormatInt(
				int64(binary.LittleEndian.Uint32(intBytes)>>8),
				10,
			),
		), nil
	case header == zipListInt8:
		b, err := r.ReadByte()
		return []byte(strconv.FormatInt(int64(b), 10)), err
	case header>>4 == zipListInt4:
		return []byte(strconv.FormatInt(int64(header&0x0f)-1, 10)), nil
	}
	return nil, fmt.Errorf("rdb: unknown unsorted set header byte: %#v", header)
}

// ReadEnd читает закрывающий байт,
// возвращает ошибку если он не найден или в нём неожиданные данные
func (r *entriesReader) ReadEnd() error {
	end, err := r.ReadByte()
	if err != nil {
		return err
	}
	if end != 255 {
		return fmt.Errorf("expected end byte 255 but actual %d", end)
	}
	return nil
}
