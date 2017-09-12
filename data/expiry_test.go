package data

import (
	"testing"
)

func TestExpiry(t *testing.T) {
	t.Run("Milliseconds", func(t *testing.T) {
		t.Run("1492780026", func(t *testing.T) {
			testExpiryMilliseconds(t, 1492780026)
		})
	})
	t.Run("Seconds", func(t *testing.T) {
		t.Run("1492780026", func(t *testing.T) {
			testExpirySeconds(t, 1492780026, 1492780.026)
		})
		t.Run("1492780526", func(t *testing.T) {
			testExpirySeconds(t, 1492780526, 1492780.526)
		})
		t.Run("1492780909", func(t *testing.T) {
			testExpirySeconds(t, 1492780909, 1492780.909)
		})
	})
}

// testExpiryMilliseconds проверяет значение времени жизни в миллисекундах
func testExpiryMilliseconds(t *testing.T, milliseconds uint64) {
	e := NewExpiry(milliseconds)
	result := e.Milliseconds()
	if result != milliseconds {
		t.Fatalf("expected expiry %d but actual %d", milliseconds, result)
	}
}

// testExpiryMilliseconds проверяет значение времени жизни в секундах
func testExpirySeconds(t *testing.T, milliseconds uint64, expected float64) {
	e := NewExpiry(milliseconds)
	result := e.Seconds()
	if result != expected {
		t.Fatalf("expected expiry %0.4f but actual %0.4f", expected, result)
	}
}
