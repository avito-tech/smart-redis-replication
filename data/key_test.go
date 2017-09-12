package data

import (
	"reflect"
	"regexp"
	"testing"
)

// newKey возвращает новый key
func newKey() *key {
	return &key{}
}

// TestKey проверяет функциональность ключа
func TestKey(t *testing.T) {
	// nolint:dupl
	t.Run("DB", func(t *testing.T) {
		t.Run("0", func(t *testing.T) {
			testKeyDB(t, 0)
		})
		t.Run("1", func(t *testing.T) {
			testKeyDB(t, 1)
		})
	})
	// nolint:dupl
	t.Run("Name", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			testKeyName(t, "name123")
		})
		t.Run("Empty", func(t *testing.T) {
			testKeyName(t, "")
		})
	})
	// nolint:dupl
	t.Run("ReplaceName", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			name := "a:b:c:d"
			src := "^a:b:(.*)"
			repl := "$1:e:f"
			expected := "c:d:e:f"
			testKeyReplaceName(t, name, src, repl, expected)
		})
	})
	// nolint:dupl
	t.Run("Expiry", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			testKeyExpiry(t, NewExpiry(1234567))
		})
		t.Run("Empty", func(t *testing.T) {

		})
	})
}

// testKeyDB проверяет установку и получение номера базы данных
// nolint:dupl
func testKeyDB(t *testing.T, db int) {
	k := newKey()
	err := k.SetDB(db)
	if err != nil {
		t.Fatalf("set db error: %q", err)
	}
	result := k.DB()
	if result != db {
		t.Fatalf("expected db number %q but actual %q", db, result)
	}
}

// testKeyName проверяет установку и получение названия ключа
// nolint:dupl
func testKeyName(t *testing.T, name string) {
	k := newKey()
	err := k.SetName(name)
	if err != nil {
		t.Fatalf("set name error: %q", err)
	}
	result := k.Name()
	if result != name {
		t.Fatalf("expected name %q but actual %q", name, result)
	}
}

// testKeyReplaceName проверяет замену названия ключа
func testKeyReplaceName(t *testing.T, name, src, repl, expected string) {
	reg, err := regexp.Compile(src)
	if err != nil {
		t.Fatalf("create regexp error: %q", err)
	}
	k := newKey()
	err = k.SetName(name)
	if err != nil {
		t.Fatalf("set name error: %q", err)
	}
	err = k.ReplaceName(reg, repl)
	if err != nil {
		t.Fatalf("replace name error :%q", err)
	}
	result := k.Name()
	if result != expected {
		t.Fatalf("expected name %q but actual %q", expected, result)
	}
}

// testKeyExpiry проверяет установку и получение времени жизни ключа
func testKeyExpiry(t *testing.T, expiry Expiry) {
	k := newKey()
	err := k.SetExpiry(expiry)
	if err != nil {
		t.Fatalf("set expiry error: %q", err)
	}
	result := k.Expiry()
	if !reflect.DeepEqual(expiry, result) {
		t.Fatalf("expected expiry %#v but actual %#v", expiry, result)
	}
}
