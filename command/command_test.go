package command

import (
	"reflect"
	"testing"

	"github.com/avito-tech/smart-redis-replication/data"
)

// TestCommandType проверяет правильное определение типа команды
func TestCommandType(t *testing.T) {
	t.Run("Ping", func(t *testing.T) {
		testCommandType(t, []string{"ping"}, Ping)
		testCommandType(t, []string{"PING"}, Ping)
		testCommandType(t, []string{"Ping"}, Ping)
	})
	t.Run("Select", func(t *testing.T) {
		testCommandType(t, []string{"select"}, Select)
		testCommandType(t, []string{"SELECT"}, Select)
		testCommandType(t, []string{"Select"}, Select)
	})
	t.Run("Zadd", func(t *testing.T) {
		testCommandType(t, []string{"zadd"}, Zadd)
		testCommandType(t, []string{"ZADD"}, Zadd)
		testCommandType(t, []string{"Zadd"}, Zadd)
	})
	t.Run("Sadd", func(t *testing.T) {
		testCommandType(t, []string{"sadd"}, Sadd)
		testCommandType(t, []string{"SADD"}, Sadd)
		testCommandType(t, []string{"Sadd"}, Sadd)
	})
	t.Run("Zrem", func(t *testing.T) {
		testCommandType(t, []string{"zrem"}, Zrem)
		testCommandType(t, []string{"ZREM"}, Zrem)
		testCommandType(t, []string{"Zrem"}, Zrem)
	})
	t.Run("Delete", func(t *testing.T) {
		testCommandType(t, []string{"delete"}, Delete)
		testCommandType(t, []string{"DELETE"}, Delete)
		testCommandType(t, []string{"Delete"}, Delete)
	})
	t.Run("Empty", func(t *testing.T) {
		testCommandType(t, []string{""}, Empty)
		testCommandType(t, []string{"   "}, Empty)
		testCommandType(t, []string{}, Empty)
	})
	t.Run("Undefined", func(t *testing.T) {
		testCommandType(t, []string{"TEST123"}, Undefined)
		testCommandType(t, []string{"test123"}, Undefined)
		testCommandType(t, []string{"_undefined_"}, Undefined)
		testCommandType(t, []string{"_other_"}, Undefined)
	})
}

func testCommandType(
	t *testing.T,
	args []string,
	expected Type,
) {
	c := New(args)
	result := c.Type()
	if result != expected {
		t.Errorf("expected %q but actual %q, args: %q", expected, result, args)
	}
}

// TestCommandConvertToSelectDB проверяет правильное конвертирование команды
func TestCommandConvertToSelectDB(t *testing.T) {
	t.Run("Normal", func(t *testing.T) {
		t.Run("0", func(t *testing.T) {
			testCommandConvertToSelectDB(
				t,
				[]string{"select", "0"},
				0,
				true,
			)
		})
		t.Run("1", func(t *testing.T) {
			testCommandConvertToSelectDB(
				t,
				[]string{"SELECT", "1"},
				1,
				true,
			)
		})
	})
	t.Run("Error", func(t *testing.T) {
		t.Run("Empty", func(t *testing.T) {
			testCommandConvertToSelectDB(
				t,
				[]string{},
				0,
				false,
			)
		})
		t.Run("NoDB", func(t *testing.T) {
			testCommandConvertToSelectDB(
				t,
				[]string{"select"},
				0,
				false,
			)
		})
		t.Run("NoSelectType", func(t *testing.T) {
			testCommandConvertToSelectDB(
				t,
				[]string{"PING", "1"},
				0,
				false,
			)
		})
		t.Run("IncorrectNumber", func(t *testing.T) {
			testCommandConvertToSelectDB(
				t,
				[]string{"select", "incorrect_number"},
				0,
				false,
			)
		})
	})
}

func testCommandConvertToSelectDB(
	t *testing.T,
	args []string,
	expected int,
	success bool,
) {
	c := New(args)
	result, err := c.ConvertToSelectDB()
	if success {
		if err != nil {
			t.Fatalf("conver error: %v", err)
		}
		if result != expected {
			t.Fatalf("expected %d db number but actual %d", expected, result)
		}
	} else {
		if err == nil {
			t.Fatalf("expected error")
		}
	}
}

// TestCommandConvertToSortedSetKey проверяет правильное конвертирование команды
func TestCommandConvertToSortedSetKey(t *testing.T) {

}

func testCommandConvertToSortedSetKey(
	t *testing.T,
	args []string,
	db int,
	expected data.SortedSetKey,
	success bool,
) {
	c := New(args)
	result, err := c.ConvertToSortedSetKey(db)
	if success {
		if err != nil {
			t.Fatalf("conver error: %v", err)
		}
		if !reflect.DeepEqual(expected, result) {
			t.Fatalf("expected key %#v but actual %#v", expected, result)
		}
	} else {
		if err == nil {
			t.Fatalf("expected error")
		}
	}
}

// TestCommandConvertToSetKey проверяет правильное конвертирование команды
func TestCommandConvertToSetKey(t *testing.T) {
	t.Run("Normal", func(t *testing.T) {

	})
	t.Run("Error", func(t *testing.T) {
		t.Run("Empty", func(t *testing.T) {
			testCommandConvertToSetKey(
				t,
				[]string{},
				0,
				nil,
				false,
			)
		})
		t.Run("NoSetType", func(t *testing.T) {
			testCommandConvertToSetKey(
				t,
				[]string{"ping", "key", "value", "value"},
				0,
				nil,
				false,
			)
		})
		t.Run("NoKeyName", func(t *testing.T) {
			testCommandConvertToSetKey(
				t,
				[]string{"sadd"},
				0,
				nil,
				false,
			)
		})
	})
}

func testCommandConvertToSetKey(
	t *testing.T,
	args []string,
	db int,
	expected data.SetKey,
	success bool,
) {
	c := New(args)
	result, err := c.ConvertToSetKey(db)
	if success {
		if err != nil {
			t.Fatalf("conver error: %v", err)
		}
		if !reflect.DeepEqual(expected, result) {
			t.Fatalf("expected key %#v but actual %#v", expected, result)
		}
	} else {
		if err == nil {
			t.Fatalf("expected error")
		}
	}
}
