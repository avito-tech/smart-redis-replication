package data

import (
	"regexp"
)

// Key это ключ с минимальным набором методов
type Key interface {
	DB() int
	SetDB(int) error
	Name() string
	Expiry() Expiry
	SetName(string) error
	SetExpiry(Expiry) error
	ReplaceName(*regexp.Regexp, string) error
}

// SortedSetKey это интерфейс упорядоченного набора данных
type SortedSetKey interface {
	Key
	Set(weight float64, value string) error
	SetData(data map[string]float64) error
	Values() map[string]float64
	Weight(value string) (float64, bool)
}

// IntegerSetKey это интерфейс отсортированного набора целых чисел
// В RDB обозначается как IntSet представляет собой дерево бинарного поиска
type IntegerSetKey interface {
	Key
	Set(value uint64) error
	SetData(data map[uint64]struct{}) error
	Values() map[uint64]struct{}
	Is(value uint64) bool
}

// SetKey это интерфейс не упорядоченного набора данных
type SetKey interface {
	Key
	Set(value string) error
	SetData(data map[string]struct{}) error
	Values() map[string]struct{}
	Is(value string) (ok bool)
}

// MapKey это интерфейс классического hash map
type MapKey interface {
	Key
	SetData(data map[string]string) error
	Set(key, value string) error
	Values() map[string]string
	Value(key string) (value string, ok bool)
	Is(key string) (ok bool)
}

// ListKey это интерфейс массива данных
type ListKey interface {
	Key
	SetData(data []string) error
	Rpush(values ...string) error
	Lpush(values ...string) error
	Values() []string
}

// StringKey это интерфейс строкового значения
type StringKey interface {
	Key
	Set(value string) error
	Value() string
}
