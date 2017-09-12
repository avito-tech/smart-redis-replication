package data

import (
	"regexp"
)

// key это общая часть для работы с ключами
type key struct {
	db     int
	name   string
	expiry Expiry
}

// ReplaceName заменяет название ключа по регулярному выражению
func (k *key) ReplaceName(srcRegexp *regexp.Regexp, repl string) error {
	k.name = string(srcRegexp.ReplaceAll([]byte(k.name), []byte(repl)))
	return nil
}

// DB возвращает номер базы данных
func (k *key) DB() int {
	return k.db
}

// SetDB устанавливает номер базы данных
func (k *key) SetDB(db int) error {
	k.db = db
	return nil
}

// GetKey возвращает название ключа
func (k *key) Name() string {
	return k.name
}

// SetKey устанавливает новое название ключа
func (k *key) SetName(name string) error {
	k.name = name
	return nil
}

// SetExpiry устанавливает время жизни ключа
func (k *key) SetExpiry(expiry Expiry) error {
	k.expiry = expiry
	return nil
}

// Expiry возвращает время жизни ключа
func (k *key) Expiry() Expiry {
	return k.expiry
}
