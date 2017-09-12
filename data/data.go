package data

import (
	"errors"
)

// SortedSet это упорядоченный набор данных
type SortedSet struct {
	key
	data map[string]float64
}

// IntegerSet это отсортированный набор целых чисел
type IntegerSet struct {
	key
	data map[uint64]struct{}
}

// Set это не упорядоченный набор данных
type Set struct {
	key
	data map[string]struct{}
}

// Map это hash map
type Map struct {
	key
	data map[string]string
}

// List это массив данных
type List struct {
	key
	data []string
}

// String это строка данных
type String struct {
	key
	value string
}

// NewSortedSet возвращает новый SortedSet
func NewSortedSet(name string) *SortedSet {
	s := new(SortedSet)
	s.name = name
	s.data = make(map[string]float64)
	return s
}

// Set устанавливает значение с весом
func (s *SortedSet) Set(weight float64, value string) error {
	s.data[value] = weight
	return nil
}

// SetData полностью меняет набор данных
func (s *SortedSet) SetData(data map[string]float64) error {
	if data == nil {
		return errors.New("expected data")
	}
	s.data = data
	return nil
}

// Weight возвращает вес значения и был ли запрос успешным
func (s *SortedSet) Weight(value string) (weight float64, ok bool) {
	weight, ok = s.data[value]
	return weight, ok
}

// Values возвращает набор данных
func (s *SortedSet) Values() map[string]float64 {
	return s.data
}

// NewIntegerSet возвращает новый IntegerSet
func NewIntegerSet(name string) *IntegerSet {
	s := new(IntegerSet)
	s.name = name
	s.data = make(map[uint64]struct{})
	return s
}

// Set устанавливает значение в набор
func (s *IntegerSet) Set(value uint64) error {
	s.data[value] = struct{}{}
	return nil
}

// SetData полностью меняет набор данных
func (s *IntegerSet) SetData(data map[uint64]struct{}) error {
	if data == nil {
		return errors.New("expected data")
	}
	s.data = data
	return nil
}

// Values возвращает набор данных
func (s *IntegerSet) Values() map[uint64]struct{} {
	return s.data
}

// Is возвращает true если значение есть, false если значения нет
func (s *IntegerSet) Is(value uint64) (ok bool) {
	_, ok = s.data[value]
	return ok
}

// NewSet возвращает новый Set
func NewSet(name string) *Set {
	s := new(Set)
	s.name = name
	s.data = make(map[string]struct{})
	return s
}

// Set устанавливает значение
func (s *Set) Set(value string) error {
	s.data[value] = struct{}{}
	return nil
}

// SetData полностью меняет набор данных
func (s *Set) SetData(data map[string]struct{}) error {
	if data == nil {
		return errors.New("expected data")
	}
	s.data = data
	return nil
}

// Values возвращает набор данных
func (s *Set) Values() map[string]struct{} {
	return s.data
}

// Is возвращает true если значение есть, false если значения нет
func (s *Set) Is(value string) (ok bool) {
	_, ok = s.data[value]
	return ok
}

// NewMap возвращает новый Map
func NewMap(name string) *Map {
	m := new(Map)
	m.name = name
	m.data = make(map[string]string)
	return m
}

// SetData полностью меняет набор данных
func (m *Map) SetData(data map[string]string) error {
	if data == nil {
		return errors.New("expected data")
	}
	m.data = data
	return nil
}

// Set устанавливает ключ-значение
func (m *Map) Set(key, value string) error {
	m.data[key] = value
	return nil
}

// Values возвращает набор данных
func (m *Map) Values() map[string]string {
	return m.data
}

// Value возвращает значение и был ли запрос успешным
func (m *Map) Value(key string) (value string, ok bool) {
	value, ok = m.data[key]
	return value, ok
}

// Is возвращает true если значение есть, false если значения нет
func (m *Map) Is(key string) (ok bool) {
	_, ok = m.data[key]
	return ok
}

// NewList возвращает новый List
func NewList(name string) *List {
	l := new(List)
	l.name = name
	return l
}

// SetData полностью меняет набор данных
func (l *List) SetData(data []string) error {
	if data == nil {
		return errors.New("expected data")
	}
	l.data = data
	return nil
}

// Rpush добавляет значения в конец списка
func (l *List) Rpush(values ...string) error {
	l.data = append(l.data, values...)
	return nil
}

// Lpush добавляет значения в начало списка
func (l *List) Lpush(values ...string) error {
	l.data = append(values, l.data...)
	return nil
}

// Values возвращает список
func (l *List) Values() []string {
	return l.data
}

// NewString возвращает новый String
func NewString(name, value string) *String {
	s := new(String)
	s.name = name
	s.value = value
	return s
}

// Set устанавливает значение
func (s *String) Set(value string) error {
	s.value = value
	return nil
}

// Value возвращает значение
func (s *String) Value() string {
	return s.value
}
