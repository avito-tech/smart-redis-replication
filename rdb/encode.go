package rdb

import (
	"encoding/binary"
	"errors"
	"math"
	"strconv"
)

const (
	encInt16 = 1
	encInt32 = 2
)

var (
	// ErrNotNumber это ошибка возникает в функции EncodeStringInt если строка
	// не содержит 32битное 10тиричное число
	ErrNotNumber = errors.New("string not number")
)

// EncodeStringInt кодирует текст в число и
// проверяет что при обратной конвертации будет тот же текст
func EncodeStringInt(number string) (int, error) {
	i, err := strconv.ParseInt(number, 10, 32)
	if err != nil {
		return 0, ErrNotNumber
	}
	if number != strconv.FormatInt(i, 10) {
		return 0, ErrNotNumber
	}
	return int(i), nil
}

// EncodeInt кодирует 32битное число в байты
func EncodeInt(i int) []byte {
	switch {
	case i >= math.MinInt8 && i <= math.MaxInt8:
		return []byte{lenEnc << 6, byte(int8(i))}
	case i >= math.MinInt16 && i <= math.MaxInt16:
		b := make([]byte, 3)
		b[0] = lenEnc<<6 | encInt16
		binary.LittleEndian.PutUint16(b[1:], uint16(int16(i)))
		return b
	case i >= math.MinInt32 && i <= math.MaxInt32:
		b := make([]byte, 5)
		b[0] = lenEnc<<6 | encInt32
		binary.LittleEndian.PutUint32(b[1:], uint32(int32(i)))
		return b
	}
	return []byte{}
}

// EncodeString кодирует строку в байты
func EncodeString(st string) []byte {
	i, err := EncodeStringInt(st)
	if err == nil {
		return EncodeInt(i)
	}

	length := EncodeLength(uint32(len(st)))
	return append(length, []byte(st)...)
}

// EncodeLength кодирует длину в байты
func EncodeLength(l uint32) []byte {
	switch {
	case l < 1<<6:
		return []byte{byte(l)}
	case l < 1<<14:
		return []byte{byte(l>>8) | len14Bit<<6, byte(l)}
	}

	b := make([]byte, 5)
	b[0] = len32Bit << 6
	binary.BigEndian.PutUint32(b[1:], l)
	return b
}

// EncodeFloat кодирует float64 в байты
func EncodeFloat(f float64) []byte {
	switch {
	case math.IsNaN(f):
		return []byte{253}
	case math.IsInf(f, 1):
		return []byte{254}
	case math.IsInf(f, -1):
		return []byte{255}
	}
	b := []byte(strconv.FormatFloat(f, 'g', 17, 64))
	return append([]byte{byte(len(b))}, b...)
}
