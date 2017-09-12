package data

import (
	"reflect"
	"testing"
)

// TestSortedSet проверяет функциональность SortedSet
func TestSortedSet(t *testing.T) {
	t.Run("Data", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			data := make(map[string]float64)
			data["value1"] = float64(1.001)
			data["value2"] = float64(2.002)
			testSortedSetData(t, data)
		})
		t.Run("Error", testSortedSetDataError)
	})
	t.Run("Set", func(t *testing.T) {
		weight := float64(123989)
		value := "value1"
		testSortedSetSet(t, weight, value)
	})
	t.Run("Weight", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			value := "value1"
			weight := float64(123456.2393)
			testSortedSetWeight(t, weight, value)
		})
		t.Run("Undefined", func(t *testing.T) {
			value := "value1"
			testSortedSetWeightUndefined(t, value)
		})
	})
}

// testSortedSetData проверяет полную замену данных в SortedSet
func testSortedSetData(t *testing.T, data map[string]float64) {
	s := NewSortedSet("")
	err := s.SetData(data)
	if err != nil {
		t.Fatalf("set data error: %q", err)
	}
	result := s.Values()
	if !reflect.DeepEqual(data, result) {
		t.Fatalf("expected data %#v but actual %#v", data, result)
	}
}

// testSortedSetDataError проверяет что SortedSet.SetData() не принимает nil
func testSortedSetDataError(t *testing.T) {
	s := NewSortedSet("")
	err := s.SetData(nil)
	if err == nil {
		t.Fatalf("expected error")
	}
}

// testSortedSetSet проверяет установку значения
func testSortedSetSet(t *testing.T, weight float64, value string) {
	s := NewSortedSet("")
	err := s.Set(weight, value)
	if err != nil {
		t.Fatalf("set error: %q", err)
	}
	result, ok := s.Weight(value)
	if !ok {
		t.Fatalf("expected weight")
	}
	if result != weight {
		t.Fatalf("expected weight %0.4f but actual %0.4f", weight, result)
	}
}

// testSortedSetWeight порверяет получение веса значения
func testSortedSetWeight(t *testing.T, weight float64, value string) {
	data := make(map[string]float64)
	data[value] = weight

	s := NewSortedSet("")
	err := s.SetData(data)
	if err != nil {
		t.Fatalf("set data error: %q", err)
	}
	result, ok := s.Weight(value)
	if !ok {
		t.Fatalf("expected weight")
	}
	if result != weight {
		t.Fatalf("expected weight %0.4f but actual %0.4f", weight, result)
	}
}

// testSortedSetWeightUndefined проверяет что значения могут отсутствовать
func testSortedSetWeightUndefined(t *testing.T, value string) {
	s := NewSortedSet("")
	_, ok := s.Weight(value)
	if ok {
		t.Fatalf("expected undefined value")
	}
}

// TestIntegerSet проверяет функциональность IntegerSet
func TestIntegerSet(t *testing.T) {
	t.Run("Data", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			data := make(map[uint64]struct{})
			data[uint64(1)] = struct{}{}
			data[uint64(2)] = struct{}{}
			testIntegerSetData(t, data)
		})
		t.Run("Error", testIntegerSetDataError)
	})
	t.Run("Set", func(t *testing.T) {
		value := uint64(2938098)
		testIntegerSetSet(t, value)
	})
	t.Run("Is", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			value := uint64(8983984)
			testIntegerSetIs(t, value)
		})
		t.Run("Undefined", func(t *testing.T) {
			value := uint64(23989823)
			testIntegerSetIsUndefined(t, value)
		})
	})
}

// testInsertSetData проверяет полную замену данных в InsertSet
func testIntegerSetData(t *testing.T, data map[uint64]struct{}) {
	s := NewIntegerSet("")
	AssertData(t, s.SetData(data), s.Values(), data)
}

// testIntegerSetDataError проверяет что InsertSet.SetData() не принимает nil
func testIntegerSetDataError(t *testing.T) {
	s := NewIntegerSet("")
	err := s.SetData(nil)
	if err == nil {
		t.Fatalf("expected error")
	}
}

// testIntegerSetSet проверяет установку значения
func testIntegerSetSet(t *testing.T, value uint64) {
	s := NewIntegerSet("")
	AssertSetIs(t, s.Set(value), s.Is(value))
}

// testIntegerSetIs проверяет наличие значения
func testIntegerSetIs(t *testing.T, value uint64) {
	data := make(map[uint64]struct{})
	data[value] = struct{}{}

	s := NewIntegerSet("")
	AssertIs(t, s.SetData(data), s.Is(value))
}

// testIntegerSetIsUndefined проверяет что значения может и не быть
func testIntegerSetIsUndefined(t *testing.T, value uint64) {
	s := NewIntegerSet("")
	ok := s.Is(value)
	if ok {
		t.Fatalf("expected undefined value")
	}
}

// TestSet проверяет функциональность Set
func TestSet(t *testing.T) {
	t.Run("Data", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			data := make(map[string]struct{})
			data["value1"] = struct{}{}
			data["value2"] = struct{}{}
			testSetData(t, data)
		})
		t.Run("Error", testSetDataError)
	})
	t.Run("Set", func(t *testing.T) {
		value := "value1"
		testSetSet(t, value)
	})
	t.Run("Is", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			value := "value1"
			testSetIs(t, value)
		})
		t.Run("Undefined", func(t *testing.T) {
			value := "value1"
			testSetIsUndefined(t, value)
		})
	})
}

// testSetData проверяет полную замену данных в Set
func testSetData(t *testing.T, data map[string]struct{}) {
	s := NewSet("")
	AssertData(t, s.SetData(data), s.Values(), data)
}

// testSetDataError проверяет что Set.SetData() не принимает nil
func testSetDataError(t *testing.T) {
	s := NewSet("")
	err := s.SetData(nil)
	if err == nil {
		t.Fatalf("expected error")
	}
}

// testSetSet проверяет установку значения
func testSetSet(t *testing.T, value string) {
	s := NewSet("")
	AssertSetIs(t, s.Set(value), s.Is(value))
}

// testSetIs проверяет наличие значения
func testSetIs(t *testing.T, value string) {
	data := make(map[string]struct{})
	data[value] = struct{}{}

	s := NewSet("")
	AssertIs(t, s.SetData(data), s.Is(value))
}

// testSetIsUndefined проверяет что значения может и не быть
func testSetIsUndefined(t *testing.T, value string) {
	s := NewSet("")
	ok := s.Is(value)
	if ok {
		t.Fatalf("expected undefined value")
	}
}

// TestMap проверяет функциональность Map
func TestMap(t *testing.T) {
	t.Run("Data", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			data := make(map[string]string)
			data["key1"] = "value1"
			data["key2"] = "value2"
			testMapData(t, data)
		})
		t.Run("Error", testMapDataError)
	})
	t.Run("Set", func(t *testing.T) {
		testMapSet(t, "key1", "value1")
	})
	t.Run("Value", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			testMapValue(t, "key1", "value1")
		})
		t.Run("Undefined", func(t *testing.T) {
			testMapValueUndefined(t, "key1")
		})
	})
	t.Run("Is", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			testMapIs(t, "key1", "value1")
		})
		t.Run("Undefined", func(t *testing.T) {
			testMapIsUndefined(t, "key1")
		})
	})
}

// testMapData проверяет полную замену данных в Map
func testMapData(t *testing.T, data map[string]string) {
	s := NewMap("")
	AssertData(t, s.SetData(data), s.Values(), data)
}

// testMapDataError проверяет что Map.SetData() не принимает nil
func testMapDataError(t *testing.T) {
	s := NewMap("")
	err := s.SetData(nil)
	if err == nil {
		t.Fatalf("expected error")
	}
}

// testMapSet проверяет установку значения
func testMapSet(t *testing.T, key, value string) {
	s := NewMap("")
	AssertSetIs(t, s.Set(key, value), s.Is(key))
}

// testMapValue проверяет наличие ключа и значения
func testMapValue(t *testing.T, key, value string) {
	data := make(map[string]string)
	data[key] = value

	s := NewMap("")
	AssertIs(t, s.SetData(data), s.Is(key))
	result, ok := s.Value(key)
	if !ok {
		t.Fatalf("expected value")
	}
	if result != value {
		t.Fatalf("expected value %q but actual %q", value, result)
	}
}

// testMapValueUndefined проверяет что ключа может и не быть в данных
func testMapValueUndefined(t *testing.T, key string) {
	s := NewMap("")
	_, ok := s.Value(key)
	if ok {
		t.Fatalf("expected undefined value")
	}
}

// testMapIs проверяет наличие значения
func testMapIs(t *testing.T, key, value string) {
	data := make(map[string]string)
	data[key] = value

	s := NewMap("")
	AssertIs(t, s.SetData(data), s.Is(key))
}

// testMapIsUndefined проверяет что ключа может и не быть в данных
func testMapIsUndefined(t *testing.T, key string) {
	s := NewMap("")
	ok := s.Is(key)
	if ok {
		t.Fatalf("expected undefined value")
	}
}

// TestList проверяет функциональность List
func TestList(t *testing.T) {
	t.Run("Data", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			data := []string{
				"value1",
				"value2",
			}
			testListData(t, data)
		})
		t.Run("Error", testListDataError)
	})
	t.Run("Rpush", func(t *testing.T) {
		testListPush(t, "right", []string{"a", "b", "c"})
	})
	t.Run("Lpush", func(t *testing.T) {
		testListPush(t, "left", []string{"a", "b", "c"})
	})
}

// testListData проверяет полную замену данных в List
func testListData(t *testing.T, data []string) {
	s := NewList("")
	AssertData(t, s.SetData(data), s.Values(), data)
}

// testListDataError проверяет что List.SetData() не принимает nil
func testListDataError(t *testing.T) {
	s := NewList("")
	err := s.SetData(nil)
	if err == nil {
		t.Fatalf("expected error")
	}
}

// testListPush проверяет добавление данных в List
func testListPush(t *testing.T, direction string, data []string) {
	list := []string{"123", "234", "345"}
	var expected []string

	s := NewList("")
	err := s.SetData(list)
	if err != nil {
		t.Fatalf("set data error: %q", err)
	}
	switch direction {
	case "left":
		err = s.Lpush(data...)
		if err != nil {
			t.Fatalf("lpush error: %q", err)
		}
		expected = append(data, list...)
	case "right":
		err = s.Rpush(data...)
		if err != nil {
			t.Fatalf("lpush error: %q", err)
		}
		expected = append(list, data...)
	default:
		t.Fatalf("undefined direction")
	}
	result := s.Values()
	if !reflect.DeepEqual(expected, result) {
		t.Fatalf("expected data %#v but actual %#v", expected, result)
	}
}

// TestString проверяет функциональность String
func TestString(t *testing.T) {
	t.Run("Name", func(t *testing.T) {
		name := "name"
		s := NewString(name, "value")
		result := s.Name()
		if result != name {
			t.Fatalf("expected name %q but actual %q", name, result)
		}
	})
	t.Run("Value", func(t *testing.T) {
		value := "value"
		s := NewString("name", value)
		result := s.Value()
		if result != value {
			t.Fatalf("expected value %q but actual %q", value, result)
		}
	})
	t.Run("Set", func(t *testing.T) {
		value := "value"
		s := NewString("", "")
		err := s.Set(value)
		if err != nil {
			t.Fatalf("set value error: %q", err)
		}
		result := s.Value()
		if result != value {
			t.Fatalf("expected value %q but actual %q", value, result)
		}
	})
}

// AssertData проверяет полную замену данных
func AssertData(t *testing.T, err error, result, expected interface{}) {
	if err != nil {
		t.Fatalf("set data error: %q", err)
	}
	if !reflect.DeepEqual(expected, result) {
		t.Fatalf("expected data %#v but actual %#v", expected, result)
	}
}

// AssertSetIs проверяет что значение удалось установить и проверить
func AssertSetIs(t *testing.T, err error, ok bool) {
	if err != nil {
		t.Fatalf("set error: %q", err)
	}
	if !ok {
		t.Fatalf("expected value")
	}
}

// AssertIs проверяет наличие значения
func AssertIs(t *testing.T, err error, ok bool) {
	if err != nil {
		t.Fatalf("set data error: %q", err)
	}
	if !ok {
		t.Fatalf("expected value")
	}
}
