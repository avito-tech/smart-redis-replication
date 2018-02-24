package backlog

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/avito-tech/smart-redis-replication/command"
)

func TestBacklog(t *testing.T) {
	t.Run("Normal", func(t *testing.T) {
		testBacklog(t, 10, 10, 10, 0, true)
	})
	t.Run("Count", func(t *testing.T) {
		testBacklog(t, 10, 10, 5, 5, true)
	})
	t.Run("Error/Overflow", func(t *testing.T) {
		testBacklogOverflow(t)
	})
}

// testBacklogOverflow проверяет наличии ошибки при переполнении backlog
func testBacklogOverflow(t *testing.T) {
	backlog := New(2)
	for i := 0; i < 2; i++ {
		cmd := command.New([]string{fmt.Sprintf("command_%d", i)})
		err := backlog.Add(cmd)
		if err != nil {
			t.Fatalf("backlog error: %q", err)
		}
	}

	cmd := command.New([]string{fmt.Sprintf("command_%d", 3)})
	err := backlog.Add(cmd)
	if err == nil {
		t.Fatalf("expected error")
	}
	if err.Error() != "queue size exceeded" {
		t.Fatalf("expected error %q but actual %v", "queue size exceeded", err)
	}
}

// testBacklog проверяет работу backlog
func testBacklog(
	t *testing.T,
	size int,
	add int,
	get int,
	expectedCount int,
	success bool,
) {
	if get > add {
		t.Fatalf("expected add > get")
	}
	backlog := New(size)

	stack := []command.Command{}
	for i := 0; i < add; i++ {
		cmd := command.New([]string{fmt.Sprintf("command_%d", i)})
		stack = append(stack, cmd)
		err := backlog.Add(cmd)
		if err != nil {
			t.Fatalf("backlog error: %v", err)
		}
	}
	if len(stack) != add {
		t.Fatalf("expected stack size %d but actual %d", add, len(stack))
	}

	for i := 0; i < get; i++ {
		cmd := backlog.Get()
		if !reflect.DeepEqual(stack[i], cmd) {
			t.Fatalf("expected command %q but actual %q", stack[i], cmd)
		}
	}
	count := backlog.Count()
	if count != expectedCount {
		t.Fatalf("expected count %d but actual %d", expectedCount, count)
	}
}
