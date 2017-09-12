package rdb

import (
	"io"

	"github.com/avito-tech/smart-redis-replication/data"
)

// KeyConsumer это потребитель ключей с данными прочитанных из RDB файла
type KeyConsumer interface {
	Key(data.Key) error
}

// Consumer это интерфейс потребителя всех данных прочитанных из RDB файла
type Consumer interface {
	// SetMagic устанавливает Magic строку
	SetMagic(Magic) error

	// SetAuxField устанавливает дополнительный параметр
	SetAuxField(AuxField) error

	// SetResizeDB устанавливает размеры базы данных
	SetResizeDB(db uint32, resizeDB ResizeDB) error

	// SetEOF устанавливает конец передачи данных
	SetEOF(EOF) error

	// Key принимает ключ с данными
	Key(data.Key) error
}

// Decoder это интерфейс для декодирования RDB файла
type Decoder interface {
	DecodeKeys(KeyConsumer) error
	Decode(Consumer) error
	Next() (interface{}, error)
}

// Reader это интерфейс для чтения специфичных для RDB форматов
type Reader interface {
	io.Reader

	// ReadByte безопасно читает один байт
	ReadByte() (byte, error)

	// ReadOpcode читает код команды
	ReadOpcode() (byte, error)

	// ReadString читает закодированную строку
	ReadString() (line string, err error)

	// SafeRead безопасно читает N байт
	SafeRead(n uint32) ([]byte, error)

	// ReadLength читает длину или кодировку
	ReadLength() (length uint32, encoding int8, err error)

	// ReadUint8 читает закодированный uint8
	ReadUint8() (uint8, error)

	// ReadFloat64 читает закодированный float64
	ReadFloat64() (float64, error)

	ReadMagic() (Magic, error)
	ReadAuxField() (AuxField, error)
	ReadDBSelector() (DBSelector, error)
	ReadResizeDB() (ResizeDB, error)
	ReadEOF() (EOF, error)
	ReadExpiry(opcode byte) (data.Expiry, error)

	ReadSet(expiry data.Expiry) (data.SetKey, error)

	// SortedSet
	ReadSortedSet(expiry data.Expiry) (data.SortedSetKey, error)
	ReadZipListSortedSet(expiry data.Expiry) (data.SortedSetKey, error)

	// HashMap
	ReadZipListHashMap(expiry data.Expiry) (data.MapKey, error)
	ReadZipMapHashMap(expiry data.Expiry) (data.MapKey, error)
	ReadListHashMap(expiry data.Expiry) (data.MapKey, error)

	// List
	ReadZipList(expiry data.Expiry) (data.ListKey, error)
	ReadQuickList(expiry data.Expiry) (data.ListKey, error)
	ReadList(expiry data.Expiry) (data.ListKey, error)

	ReadIntSet(expiry data.Expiry) (data.IntegerSetKey, error)
	ReadStringValue(expiry data.Expiry) (data.StringKey, error)
}

// EntriesReader это интерфейс для чтения ZipList структур
type EntriesReader interface {
	io.Reader

	// ReadByte безопасно читает один байт
	ReadByte() (byte, error)

	// ReadEntryLength читает размер элементов
	ReadEntryLength() (uint32, error)

	// ReadTail читает смещение последнего элемента
	ReadTail() (uint32, error)

	// ReadEntriesCount читает количество элементов
	ReadEntriesCount() (uint16, error)

	// ReadEntry читает элемент
	ReadEntry() ([]byte, error)
}

// IntSetReader это интерфейс для чтения бинарного дерева поиска целых чисел
// Отличается от EntriesReader тем что все элементы это целые числа
// одного размера: uint64, uint32, uint16
type IntSetReader interface {
	// SafeRead безопасно читает N-байт
	SafeRead(n uint32) ([]byte, error)

	// ReadEntryLength читает размерность элементов (example: 2, 4, 8 bytes)
	ReadEntryLength() (uint32, error)

	// ReadEntriesCount читает количество элементов
	ReadEntriesCount() (uint32, error)

	// ReadEntry64 читает entry в формате uint64
	ReadEntry64() (uint64, error)

	// ReadEntry32 читает entry в формате uint32
	ReadEntry32() (uint32, error)

	// ReadEntry16 читает entry в формате uint16
	ReadEntry16() (uint16, error)
}

// ZipMapReader это интерфейс для чтения ZipMap структур
type ZipMapReader interface {
	// SafeRead безопасно читает N-байт
	SafeRead(n uint32) ([]byte, error)

	// ReadCount читает количество элементов
	ReadCount() (uint8, error)

	// ReadLength читает длину строки
	ReadLength() (uint32, error)

	// ReadKey читает ключ
	ReadKey() ([]byte, error)

	// ReadValue читает значение ключа
	ReadValue() ([]byte, error)
}
