package data

import (
	"testing"
)

// TestInteface проверяет соответствие структур интерфейсам
func TestInterface(t *testing.T) {
	t.Run("Key", func(t *testing.T) {
		t.Run("SortedSet", func(t *testing.T) {
			testInterfaceKey(t, new(SortedSet))
		})
		t.Run("IntegerSet", func(t *testing.T) {
			testInterfaceKey(t, new(IntegerSet))
		})
		t.Run("Set", func(t *testing.T) {
			testInterfaceKey(t, new(Set))
		})
		t.Run("Map", func(t *testing.T) {
			testInterfaceKey(t, new(Map))
		})
		t.Run("List", func(t *testing.T) {
			testInterfaceKey(t, new(List))
		})
		t.Run("String", func(t *testing.T) {
			testInterfaceKey(t, new(String))
		})
	})
	t.Run("SortedSetKey", func(t *testing.T) {
		testInterfaceSortedSetKey(t, new(SortedSet))
	})
	t.Run("IntegerSetKey", func(t *testing.T) {
		testInterfaceIntegerSetKey(t, new(IntegerSet))
	})
	t.Run("SetKey", func(t *testing.T) {
		testInterfaceSetKey(t, new(Set))
	})
	t.Run("MapKey", func(t *testing.T) {
		testInterfaceMapKey(t, new(Map))
	})
	t.Run("ListKey", func(t *testing.T) {
		testInterfaceListKey(t, new(List))
	})
	t.Run("StringKey", func(t *testing.T) {
		testInterfaceStringKey(t, new(String))
	})
}

// testInterfaceKey проверяет принадлежность к интерфейсу Key
func testInterfaceKey(t *testing.T, key interface{}) {
	if _, ok := key.(Key); !ok {
		t.Fatalf("does not implement the interface")
	}
}

// testInterfaceSortedSetKey проверяет принадлежность к интерфейсу SortedSetKey
func testInterfaceSortedSetKey(t *testing.T, key interface{}) {
	if _, ok := key.(SortedSetKey); !ok {
		t.Fatalf("does not implement the interface")
	}
}

// testInterfaceIntegerSetKey
// проверяет принадлежность к интерфейсу IntegerSetKey
func testInterfaceIntegerSetKey(t *testing.T, key interface{}) {
	if _, ok := key.(IntegerSetKey); !ok {
		t.Fatalf("does not implement the interface")
	}
}

// testInterfaceSetKey проверяет принадлежность к интерфейсу SetKey
func testInterfaceSetKey(t *testing.T, key interface{}) {
	if _, ok := key.(SetKey); !ok {
		t.Fatalf("does not implement the interface")
	}
}

// testInterfaceMapKey проверяет принадлежность к интерфейсу MapKey
func testInterfaceMapKey(t *testing.T, key interface{}) {
	if _, ok := key.(MapKey); !ok {
		t.Fatalf("does not implement the interface")
	}
}

// testInterfaceListKey проверяет принадлежность к интерфейсу ListKey
func testInterfaceListKey(t *testing.T, key interface{}) {
	if _, ok := key.(ListKey); !ok {
		t.Fatalf("does not implement the interface")
	}
}

// testInterfaceStringKey проверяет принадлежность к интерфейсу StringKey
func testInterfaceStringKey(t *testing.T, key interface{}) {
	if _, ok := key.(StringKey); !ok {
		t.Fatalf("does not implement the interface")
	}
}
