package rdb

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"

	"github.com/avito-tech/smart-redis-replication/data"
)

const (
	// DBSelectorOpcode это индикатор номера базы данных
	DBSelectorOpcode = 0xFE

	// ResizeDBOpcode это индикатор изменения размеров базы (rdb version 7)
	ResizeDBOpcode = 0xFB

	// AuxFieldOpcode это индикатор дополнительного поля (rdb version 7)
	AuxFieldOpcode = 0xFA

	// ExpirySecondsOpcode это индикатор времени жизни ключа,
	// находится перед ключём к которому относится
	ExpirySecondsOpcode = 0xFD

	// ExpiryMillisecondsOpcode это индикатор времени жизни ключа,
	// находится перед ключём к которому относится
	ExpiryMillisecondsOpcode = 0xFC

	// EOFOpcode это индикатор конца файла, с версии 5 после него идёт 8 byte
	// контрольной суммы,
	// если эта опция на сервере отключена то 8 byte заполнены нулями
	EOFOpcode = 0xFF

	//StringValueOpcode это индикатор строки
	StringValueOpcode = 0x00

	// ListOpcode ...
	ListOpcode = 0x01

	// SetOpcode это индикатор
	SetOpcode = 0x02

	// SortedSetOpcode это индикатор SortedSet закодированного через List
	SortedSetOpcode = 0x03

	// ListHashMapOpcode это индикатор HashMap закодированного через List
	ListHashMapOpcode = 0x04

	// ZipMapHashMapOpcode это индикатор HashMap (собственный формат)
	ZipMapHashMapOpcode = 0x09

	// ZipListOpcode это индикатор сжатого неупорядоченного набора уникальных
	// значений
	ZipListOpcode = 0x0a

	// IntSetOpcode это индикатор двоичного дерева поиска целых чисел.
	// IntSet используется, когда все элементы набора являются целыми числами.
	// Номера в наборе всегда отсортированы
	IntSetOpcode = 0x0b

	// ZipListSortedSetOpcode это индикатор SortedSet закодированного через ZipList
	ZipListSortedSetOpcode = 0x0c

	// ZipListHashMapOpcode это индикатор HashMap закодированного черезе ZipList
	ZipListHashMapOpcode = 0x0d

	// QuickListOpcode это индикатор List закодированного через QuickList
	QuickListOpcode = 0x0e

	len6Bit  = 0x0
	len14Bit = 0x1
	len32Bit = 0x2
	lenEnc   = 0x3
)

var (
	// DefaultReaderSize это размер буфера чтения по умолчанию
	DefaultReaderSize = 16384

	// rdbSignature это обязательная часть файла []byte("REDIS")
	rdbSignature = []byte{0x52, 0x45, 0x44, 0x49, 0x53}

	// ErrValuesIsNotUsed эта ошибка означает что значение использовать не нужно
	ErrValuesIsNotUsed = errors.New("value is not used")
)

// reader реализует интерфейс Reader
type reader struct {
	*bufio.Reader
}

// NewReader возвращает новый Reader
func NewReader(r io.Reader) Reader {
	return &reader{
		Reader: bufio.NewReaderSize(r, DefaultReaderSize),
	}
}

// NewStringReader возвращает новый Reader
func NewStringReader(st string) Reader {
	r := bytes.NewBufferString(st)
	return &reader{Reader: bufio.NewReaderSize(r, DefaultReaderSize)}
}

// SafeRead безопасно читает N байт
func (r *reader) SafeRead(n uint32) ([]byte, error) {
	result := make([]byte, n)
	_, err := io.ReadFull(r.Reader, result)
	return result, err
}

// ReadOpcode читает код команды
func (r *reader) ReadOpcode() (byte, error) {
	return r.Reader.ReadByte()
}

// ReadString читает строку RDB файла, поддерживается только несжатая версия
// nolint:gocyclo
func (r *reader) ReadString() (string, error) {
	length, encoding, err := r.ReadLength()
	if err != nil {
		return "", err
	}

	switch encoding {
	// length-prefixed string
	case -1:
		data, err := r.SafeRead(length)
		if err != nil {
			return "", err
		}
		return string(data), nil

	// integer as string
	case 0, 1, 2:
		data, err := r.SafeRead(1 << uint8(encoding))
		if err != nil {
			return "", err
		}
		var num uint32

		if encoding == 0 {
			num = uint32(data[0])
		} else if encoding == 1 {
			num = uint32(data[0]) | (uint32(data[1]) << 8)
		} else if encoding == 2 {
			num = uint32(data[0]) | (uint32(data[1]) << 8) | (uint32(data[2]) << 16) | (uint32(data[3]) << 24) //nolint:lll
		}
		return fmt.Sprintf("%d", num), nil

	// compressed string
	case 3:
		clength, _, err := r.ReadLength()
		if err != nil {
			return "", err
		}
		length, _, err := r.ReadLength()
		if err != nil {
			return "", err
		}
		data, err := r.SafeRead(clength)
		if err != nil {
			return "", err
		}
		result := string(lzfDecompress(data, length))
		if len(result) != int(length) {
			return "", fmt.Errorf(
				"expected decompressed string length %d but actual %d",
				length,
				len(result),
			)
		}
		return result, nil
	}
	return "", fmt.Errorf("unsupported string encoding")
}

// ReadLength читает префикс зашифрованной длины
func (r *reader) ReadLength() (length uint32, encoding int8, err error) {
	prefix, err := r.ReadByte()
	if err != nil {
		return 0, 0, err
	}
	kind := (prefix & 0xC0) >> 6

	switch kind {
	case len6Bit:
		length = uint32(prefix & 0x3F)
		return length, -1, nil
	case len14Bit:
		data, err := r.ReadByte()
		if err != nil {
			return 0, 0, err
		}
		length = (uint32(prefix&0x3F) << 8) | uint32(data)
		return length, -1, nil
	case len32Bit:
		data, err := r.SafeRead(4)
		if err != nil {
			return 0, 0, err
		}
		length = binary.BigEndian.Uint32(data)
		return length, -1, nil
	case lenEnc:
		encoding = int8(prefix & 0x3F)
		return 0, encoding, nil
	}
	return 0, 0, fmt.Errorf("Undefined length")
}

// ReadUint8 читает один байт
func (r *reader) ReadUint8() (uint8, error) {
	d, err := r.ReadByte()
	if err != nil {
		return 0, err
	}
	return d, nil
}

// ReadFloat64 читает значение с двойной точностью,
// в том числе бесконечно положительное и бесконечно отрицательное число
func (r *reader) ReadFloat64() (float64, error) {
	length, err := r.ReadUint8()
	if err != nil {
		return 0, err
	}
	switch length {
	case 253:
		return math.NaN(), nil
	case 254:
		return math.Inf(0), nil
	case 255:
		return math.Inf(-1), nil
	default:
		floatBytes, err := r.SafeRead(uint32(length))
		if err != nil {
			return 0, err
		}
		f, err := strconv.ParseFloat(string(floatBytes), 64)
		if err != nil {
			return 0, err
		}
		return f, nil
	}
}

// Magic это сигнатура RDB файла
type Magic struct {
	rdbVersion uint32
}

// NewMagic возвращает новый Magic
func NewMagic(rdbVersion uint32) Magic {
	return Magic{rdbVersion: rdbVersion}
}

// GetRDBVersion возвращает номер rdb версии
func (m Magic) GetRDBVersion() uint32 {
	return m.rdbVersion
}

// Bytes возвращает бинарное представление
func (m Magic) Bytes() []byte {
	version := []byte(fmt.Sprintf("%04d", m.rdbVersion))
	return append(rdbSignature, version...)
}

// ReadMagic читает Magic сигнатуру файла
func (r *reader) ReadMagic() (Magic, error) {
	signature, err := r.SafeRead(5)
	if err != nil {
		return Magic{}, err
	}
	if !bytes.Equal(signature, rdbSignature) {
		return Magic{}, fmt.Errorf(
			"expected rdb signature %#v but actual %#v (%q)",
			rdbSignature,
			signature,
			string(signature),
		)
	}
	version, err := r.SafeRead(4)
	if err != nil {
		return Magic{}, fmt.Errorf("error rdb version: %q", err)
	}
	return Magic{
		rdbVersion: binary.LittleEndian.Uint32(version),
	}, nil
}

// AuxField это дополнительное поле относящееся к всему файлу
type AuxField struct {
	key   string
	value string
}

// NewAuxField возвращает новый AuxField
func NewAuxField(key, value string) AuxField {
	return AuxField{
		key:   key,
		value: value,
	}
}

// GetKey возвращает название ключа
func (a AuxField) GetKey() string {
	return a.key
}

// GetValue возвращает название ключа
func (a AuxField) GetValue() string {
	return a.value
}

// Bytes возвращает бинарное представление
// nolint:dupl
func (a AuxField) Bytes() []byte {
	key := EncodeString(a.key)
	value := EncodeString(a.value)

	data := make([]byte, 0, len(key)+len(value)+1)

	buffer := bytes.NewBuffer(data)
	buffer.WriteByte(AuxFieldOpcode)
	buffer.Write(key)
	buffer.Write(value)

	return buffer.Bytes()
}

// ReadAuxField читает дополнительные поля относящиеся к файлу:
// версия сервера, время создания и прочее
func (r *reader) ReadAuxField() (AuxField, error) {
	key, err := r.ReadString()
	if err != nil {
		return AuxField{}, err
	}
	value, err := r.ReadString()
	if err != nil {
		return AuxField{}, err
	}
	return AuxField{
		key:   key,
		value: value,
	}, nil
}

// DBSelector это селектор номера базы данных,
// после него идут данные предназначенные для базы данных с таким номером
type DBSelector struct {
	dbNumber uint32
}

// NewDBSelector возвращает новый DBSelector
func NewDBSelector(dbNumber uint32) DBSelector {
	return DBSelector{
		dbNumber: dbNumber,
	}
}

// GetDBNumber возвращает номер базы данных
func (db DBSelector) GetDBNumber() uint32 {
	return db.dbNumber
}

// Bytes возвращает бинарное представление
func (db DBSelector) Bytes() []byte {
	number := EncodeLength(db.dbNumber)
	return append([]byte{DBSelectorOpcode}, number...)
}

// ReadDBSelector читает номер базы данных
func (r *reader) ReadDBSelector() (DBSelector, error) {
	db, _, err := r.ReadLength()
	if err != nil {
		return DBSelector{}, err
	}
	return DBSelector{
		dbNumber: db,
	}, nil
}

// ResizeDB это данные о размере базы данных
type ResizeDB struct {
	hashTableSize       uint32
	expiryHashTableSize uint32
}

// NewResizeDB возвращает новый ResizeDB
func NewResizeDB(hashTableSize, expiryHashTableSize uint32) ResizeDB {
	return ResizeDB{
		hashTableSize:       hashTableSize,
		expiryHashTableSize: expiryHashTableSize,
	}
}

// GetHashTableSize возвращает размер HashTable
func (r ResizeDB) GetHashTableSize() uint32 {
	return r.hashTableSize
}

// GetExpiryHashTableSize возвращает размер HashTable,
// данные в котором устаревают
func (r ResizeDB) GetExpiryHashTableSize() uint32 {
	return r.expiryHashTableSize
}

// Bytes возвращает бинарное представление
// nolint:dupl
func (r ResizeDB) Bytes() []byte {
	hashTableSize := EncodeLength(r.hashTableSize)
	expiryHashTableSize := EncodeLength(r.expiryHashTableSize)

	data := make([]byte, 0, len(hashTableSize)+len(expiryHashTableSize)+1)

	buffer := bytes.NewBuffer(data)
	buffer.WriteByte(ResizeDBOpcode)
	buffer.Write(hashTableSize)
	buffer.Write(expiryHashTableSize)

	return buffer.Bytes()
}

// parseResizeDB читает размер базы данных
func (r *reader) ReadResizeDB() (ResizeDB, error) {
	hashTableSize, _, err := r.ReadLength()
	if err != nil {
		return ResizeDB{}, err
	}
	expiryHashTableSize, _, err := r.ReadLength()
	if err != nil {
		return ResizeDB{}, err
	}

	return ResizeDB{
		hashTableSize:       hashTableSize,
		expiryHashTableSize: expiryHashTableSize,
	}, nil
}

// EOF означает конец файла, доступен с RDB версии 5
type EOF struct {
	checksum uint64
}

// NewEOF возвращает новый EOF
func NewEOF() EOF {
	return EOF{}
}

// Bytes возвращает бинарное представление
func (e EOF) Bytes() []byte {
	data := make([]byte, 9)
	data[0] = EOFOpcode
	binary.LittleEndian.PutUint64(data[1:], e.checksum)
	return data
}

// ReadEOF читает EOF
// EOF доступен с RDB версии 5
func (r *reader) ReadEOF() (EOF, error) {
	checksum, err := r.SafeRead(8)
	if err != nil {
		return EOF{}, fmt.Errorf("error checksum: %q", err)
	}
	return EOF{
		checksum: binary.LittleEndian.Uint64(checksum),
	}, io.EOF
}

// ReadExpiry читает время жизни ключа
func (r *reader) ReadExpiry(opcode byte) (data.Expiry, error) {
	var expiry uint64

	switch opcode {
	case ExpirySecondsOpcode:
		body, err := r.SafeRead(4)
		if err != nil {
			return data.NewExpiry(0), err
		}
		expiry = uint64(binary.LittleEndian.Uint32(body)) * 1000
	case ExpiryMillisecondsOpcode:
		body, err := r.SafeRead(8)
		if err != nil {
			return data.NewExpiry(0), err
		}
		expiry = binary.LittleEndian.Uint64(body)
	default:
		return data.NewExpiry(0), fmt.Errorf("unexpected expiry opcode %#v", opcode)
	}

	return data.NewExpiry(expiry), nil
}

// ReadSet читает неупорядоченный набор значений
// nolint:dupl
func (r *reader) ReadSet(expiry data.Expiry) (data.SetKey, error) {
	keyName, err := r.ReadString()
	if err != nil {
		return nil, err
	}

	key := data.NewSet(keyName)
	err = key.SetExpiry(expiry)
	if err != nil {
		return nil, err
	}
	err = r.DecodeSetList(key)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// DecodeSetList декодирует Set закодированный через List
// nolint:dupl
func (r *reader) DecodeSetList(key data.SetKey) error {
	count, _, err := r.ReadLength()
	if err != nil {
		return err
	}
	for count > 0 {
		count--
		value, err := r.ReadString()
		if err != nil {
			return err
		}
		err = key.Set(value)
		if err != nil {
			return err
		}
	}
	return nil
}

// ReadZipListSortedSet читает SortedSet закодированный с помощью ZipList
// nolint:dupl
func (r *reader) ReadZipListSortedSet(
	expiry data.Expiry,
) (
	data.SortedSetKey,
	error,
) {
	keyName, err := r.ReadString()
	if err != nil {
		return nil, err
	}

	body, err := r.ReadString()
	if err != nil {
		return nil, err
	}

	key := data.NewSortedSet(keyName)
	err = key.SetExpiry(expiry)
	if err != nil {
		return nil, err
	}

	err = r.DecodeSortedSetZipList(key, body)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// DecodeSortedSetZipList декодирует значения SortedSet закодированные через
// ZipList
// nolint:dupl
func (r *reader) DecodeSortedSetZipList(
	key data.SortedSetKey,
	body string,
) error {
	dec := NewEntriesStringReader(body)

	_, err := dec.ReadEntryLength()
	if err != nil {
		return err
	}

	_, err = dec.ReadTail()
	if err != nil {
		return err
	}
	count, err := dec.ReadEntriesCount()
	if err != nil {
		return err
	}
	if count%2 == 1 {
		return fmt.Errorf(
			`expected "even" entries count bud actual "odd", count %d`,
			count,
		)
	}

	for count > 1 {
		count -= 2
		value, err := dec.ReadEntry()
		if err != nil {
			return err
		}

		scoreBytes, err := dec.ReadEntry()
		if err != nil {
			return err
		}
		score, err := strconv.ParseFloat(string(scoreBytes), 64)
		if err != nil {
			return err
		}
		err = key.Set(score, string(value))
		if err != nil {
			return err
		}
	}
	return nil
}

// ReadSortedSet читает SortedSet закодированный с помощью List
// nolint:dupl
func (r *reader) ReadSortedSet(expiry data.Expiry) (data.SortedSetKey, error) {
	keyName, err := r.ReadString()
	if err != nil {
		return nil, err
	}

	key := data.NewSortedSet(keyName)
	err = key.SetExpiry(expiry)
	if err != nil {
		return nil, err
	}

	err = r.DecodeSortedSetList(key)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// DecodeSortedSetList декодирует значения SortedSet закодированные через List
// nolint:dupl
func (r *reader) DecodeSortedSetList(key data.SortedSetKey) error {
	count, _, err := r.ReadLength()
	if err != nil {
		return err
	}
	for count > 0 {
		count--

		value, err := r.ReadString()
		if err != nil {
			return err
		}
		score, err := r.ReadFloat64()
		if err != nil {
			return err
		}
		err = key.Set(score, value)
		if err != nil {
			return err
		}
	}
	return nil
}

// ZipMapReader это Reader для чтения ZipMap формата
type zipMapReader struct {
	*bufio.Reader
}

// NewZipMapReader возвращает новый ZipMapReader
func NewZipMapReader(data string) ZipMapReader {
	return &zipMapReader{
		Reader: bufio.NewReaderSize(bytes.NewBufferString(data), DefaultReaderSize),
	}
}

// SafeRead безопасно читает N байт
func (r *zipMapReader) SafeRead(n uint32) ([]byte, error) {
	result := make([]byte, n)
	_, err := io.ReadFull(r.Reader, result)
	return result, err
}

// ReadCount читает количество элементов
func (r *zipMapReader) ReadCount() (uint8, error) {
	count, err := r.ReadByte()
	if err != nil {
		return 0, err
	}
	if count < 254 {
		return count, nil
	}
	return 0, ErrValuesIsNotUsed
}

// ReadLength читает длину строки
func (r *zipMapReader) ReadLength() (uint32, error) {
	lenByte, err := r.ReadByte()
	if err != nil {
		return 0, err
	}
	if lenByte < 253 {
		return uint32(lenByte), nil
	} else if lenByte == 253 {
		data, err := r.SafeRead(4)
		if err != nil {
			return 0, fmt.Errorf("error read length: %q", err)
		}
		return binary.LittleEndian.Uint32(data), nil
	}
	return 0, fmt.Errorf("unexpected length byte %#v", lenByte)
}

// ReadKey читает название ключа в zipMap hashMap
func (r *zipMapReader) ReadKey() ([]byte, error) {
	length, err := r.ReadLength()
	if err != nil {
		return nil, err
	}
	data, err := r.SafeRead(length)
	if err != nil {
		return nil, fmt.Errorf("error read key: %q", err)
	}
	return data, nil
}

// ReadValue читает значение ключа в zipMap hashMap
func (r *zipMapReader) ReadValue() ([]byte, error) {
	length, err := r.ReadLength()
	if err != nil {
		return nil, err
	}
	free, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	data, err := r.SafeRead(length)
	if err != nil {
		return nil, err
	}
	if free > 0 {
		if int(free) > len(data) {
			return nil, fmt.Errorf(
				"expected free < length but actual free:%d and length:%d",
				uint32(free),
				len(data),
			)
		}
		return data[0 : len(data)-int(free)], nil
	}
	return data, nil
}

// ReadZipListHashMap читает HashMap закодированный с помощью ZipList
// nolint:dupl
func (r *reader) ReadZipListHashMap(expiry data.Expiry) (data.MapKey, error) {
	keyName, err := r.ReadString()
	if err != nil {
		return nil, err
	}

	body, err := r.ReadString()
	if err != nil {
		return nil, err
	}

	key := data.NewMap(keyName)
	err = key.SetExpiry(expiry)
	if err != nil {
		return nil, err
	}
	err = r.DecodeZipListHashMap(key, body)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// DecodeZipList декодирует ZipList реализацию
// nolint:dupl
func (r *reader) DecodeZipListHashMap(key data.MapKey, body string) error {
	dec := NewEntriesStringReader(body)

	_, err := dec.ReadEntryLength()
	if err != nil {
		return err
	}

	_, err = dec.ReadTail()
	if err != nil {
		return err
	}
	count, err := dec.ReadEntriesCount()
	if err != nil {
		return err
	}
	if count%2 == 1 {
		return fmt.Errorf(
			`expected "even" entries count bud actual "odd", count %d`,
			count,
		)
	}

	for count > 1 {
		count -= 2

		keyName, err := dec.ReadEntry()
		if err != nil {
			return err
		}

		keyValue, err := dec.ReadEntry()
		if err != nil {
			return err
		}

		err = key.Set(string(keyName), string(keyValue))
		if err != nil {
			return err
		}
	}
	return nil
}

// ReadHashMap читает HashMap закодированный с помощью List
// nolint:dupl
func (r *reader) ReadListHashMap(expiry data.Expiry) (data.MapKey, error) {
	keyName, err := r.ReadString()
	if err != nil {
		return nil, err
	}

	key := data.NewMap(keyName)
	err = key.SetExpiry(expiry)
	if err != nil {
		return nil, err
	}
	err = r.DecodeHashMapList(key)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// DecodeList декодирует List реализацию
// nolint:dupl
func (r *reader) DecodeHashMapList(key data.MapKey) error {
	count, _, err := r.ReadLength()
	if err != nil {
		return err
	}
	for count > 0 {
		count--

		keyName, err := r.ReadString()
		if err != nil {
			return err
		}
		keyValue, err := r.ReadString()
		if err != nil {
			return err
		}
		err = key.Set(keyName, keyValue)
		if err != nil {
			return err
		}
	}
	return nil
}

// ReadZipMapHashMap читает HashMap закодированный в строку
// nolint:dupl
func (r *reader) ReadZipMapHashMap(expiry data.Expiry) (data.MapKey, error) {
	keyName, err := r.ReadString()
	if err != nil {
		return nil, err
	}
	body, err := r.ReadString()
	if err != nil {
		return nil, err
	}

	key := data.NewMap(keyName)
	err = key.SetExpiry(expiry)
	if err != nil {
		return nil, err
	}
	err = r.DecodeZipMapHashMap(key, body)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// DecodeZipMap декодирует ZipMap реалиазацию
//   Структура строки:
//      <zmlen>
//            <len> "foo" <len> <free> "bar"
//            <len> "hello" <len> <free> "world"
//      <zmend>
//      zmlen - 1 байт, который содержит размер zip-карты
//         - если значение больше или равно 254 то
//            такое значение не используется
//           (в этом случае длина вычисляется чтением всей карты)
//      len - длина строки, отличается от ReadLength()
//      free - 1 байт, хранит количество свободных байт в конце строки
//      Zmend - 1 байт, всегда 255, указывает на конец файла
// nolint:dupl,gocyclo
func (r *reader) DecodeZipMapHashMap(key data.MapKey, body string) error {
	dec := NewZipMapReader(body)

	count, err := dec.ReadCount()
	if err == ErrValuesIsNotUsed {
		for {
			keyName, err := dec.ReadKey()
			if err == io.EOF {
				return nil
			}
			if err != nil {
				return err
			}
			keyValue, err := dec.ReadValue()
			if err != nil {
				return err
			}
			err = key.Set(string(keyName), string(keyValue))
			if err != nil {
				return err
			}
		}
	}
	for count > 0 {
		count--
		keyName, err := dec.ReadKey()
		if err != nil {
			return err
		}
		keyValue, err := dec.ReadValue()
		if err != nil {
			return err
		}
		err = key.Set(string(keyName), string(keyValue))
		if err != nil {
			return err
		}
	}
	return nil
}

// ReadZipList читает List закодированный как ZipList
// nolint:dupl
func (r *reader) ReadZipList(expiry data.Expiry) (data.ListKey, error) {
	keyName, err := r.ReadString()
	if err != nil {
		return nil, err
	}

	body, err := r.ReadString()
	if err != nil {
		return nil, err
	}

	key := data.NewList(keyName)
	err = key.SetExpiry(expiry)
	if err != nil {
		return nil, err
	}
	err = r.DecodeZipList(key, body)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// DecodeZipList декодирует ZipList реализацию
// nolint:dupl
func (r *reader) DecodeZipList(key data.ListKey, body string) error {
	dec := NewEntriesStringReader(body)

	_, err := dec.ReadEntryLength()
	if err != nil {
		return err
	}

	_, err = dec.ReadTail()
	if err != nil {
		return err
	}
	count, err := dec.ReadEntriesCount()
	if err != nil {
		return err
	}

	for count > 0 {
		count--
		value, err := dec.ReadEntry()
		if err != nil {
			return err
		}

		err = key.Rpush(string(value))
		if err != nil {
			return err
		}
	}
	return nil
}

// ReadQuickList читает List закодированный как несколько ZipList,
// объединяя их в один List
// nolint:dupl
func (r *reader) ReadQuickList(expiry data.Expiry) (data.ListKey, error) {
	keyName, err := r.ReadString()
	if err != nil {
		return nil, err
	}
	key := data.NewList(keyName)
	err = key.SetExpiry(expiry)
	if err != nil {
		return nil, err
	}
	err = r.DecodeQuickList(key)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// DecodeQuickList декодирует QuickList реализацию,
// QuickList это список состоящий из ZipList
// nolint:dupl
func (r *reader) DecodeQuickList(key data.ListKey) error {
	count, _, err := r.ReadLength()
	if err != nil {
		return err
	}

	for count > 0 {
		count--
		body, err := r.ReadString()
		if err != nil {
			return err
		}
		err = r.DecodeZipList(key, body)
		if err != nil {
			return err
		}
	}
	return nil
}

// ReadList читает List закодированный без сжатия
// nolint:dupl
func (r *reader) ReadList(expiry data.Expiry) (data.ListKey, error) {
	keyName, err := r.ReadString()
	if err != nil {
		return nil, err
	}
	count, _, err := r.ReadLength()
	if err != nil {
		return nil, err
	}

	key := data.NewList(keyName)
	err = key.SetExpiry(expiry)
	if err != nil {
		return nil, err
	}
	for count > 0 {
		count--
		value, err := r.ReadString()
		if err != nil {
			return nil, err
		}
		err = key.Rpush(value)
		if err != nil {
			return nil, err
		}
	}
	return key, nil
}

// ReadIntSet читает InegerSet
// nolint:dupl
func (r *reader) ReadIntSet(expiry data.Expiry) (data.IntegerSetKey, error) {
	keyName, err := r.ReadString()
	if err != nil {
		return nil, err
	}

	body, err := r.ReadString()
	if err != nil {
		return nil, err
	}

	key := data.NewIntegerSet(keyName)
	err = key.SetExpiry(expiry)
	if err != nil {
		return nil, err
	}
	err = r.DecodeIntegerSet(key, body)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// Decode декодирует IntSet из строки
//   Структура строки: <encoding><length-of-contents><contents>
//     encoding - тип чисел, 2, 4, 8 байтовые (uint8, uint32, uint64)
//     length-of-contents - количество элементов
//     contents - перечень элементов кратные encoding
// nolint:gocyclo
func (r *reader) DecodeIntegerSet(key data.IntegerSetKey, body string) error {
	dec := NewIntSetStringReader(body)
	length, err := dec.ReadEntryLength()
	if err != nil {
		return err
	}
	count, err := dec.ReadEntriesCount()
	if err != nil {
		return err
	}
	switch length {
	case 2:
		for count > 0 {
			count--
			value, err := dec.ReadEntry16()
			if err != nil {
				return err
			}
			err = key.Set(uint64(value))
			if err != nil {
				return err
			}
		}
	case 4:
		for count > 0 {
			count--
			value, err := dec.ReadEntry32()
			if err != nil {
				return err
			}
			err = key.Set(uint64(value))
			if err != nil {
				return err
			}
		}
	case 8:
		for count > 0 {
			count--
			value, err := dec.ReadEntry64()
			if err != nil {
				return err
			}
			err = key.Set(value)
			if err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("expected entry length 2, 4, 8 bytes but actual %d", length)
	}
	return nil
}

// ReadStringValue читает ключ-значение
// nolint:dupl
func (r *reader) ReadStringValue(expiry data.Expiry) (data.StringKey, error) {
	keyName, err := r.ReadString()
	if err != nil {
		return nil, err
	}

	value, err := r.ReadString()
	if err != nil {
		return nil, err
	}

	key := data.NewString(keyName, value)
	err = key.SetExpiry(expiry)
	if err != nil {
		return nil, err
	}
	return key, nil
}
