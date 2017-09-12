package resp

import (
	"reflect"
	"testing"

	"github.com/avito-tech/smart-redis-replication/command"
)

func TestReaderCommand(t *testing.T) {
	t.Run("Normal", func(t *testing.T) {
		t.Run("PING", func(t *testing.T) {
			testReadCommand(
				t,
				"*1\r\n$4\r\nPING\r\n",
				command.New([]string{"PING"}),
				true,
			)
		})
		t.Run("Select", func(t *testing.T) {
			testReadCommand(
				t,
				"*2\r\n$6\r\nSELECT\r\n:10\r\n",
				command.New([]string{"SELECT", "10"}),
				true,
			)
		})
		// nolint:dupl
		t.Run("Zadd", func(t *testing.T) {
			t.Run("StringValue", func(t *testing.T) {
				testReadCommand(
					t,
					"*6\r\n$4\r\nZADD\r\n$9\r\nkey:1:2:3\r\n:123456\r\n$5\r\nID123\r\n:23456\r\n$6\r\nID2345\r\n", // nolint:lll
					command.New([]string{"ZADD", "key:1:2:3", "123456", "ID123", "23456", "ID2345"}),              // nolint:lll
					true,
				)
			})
			t.Run("IntegerValue", func(t *testing.T) {
				testReadCommand(
					t,
					"*6\r\n$4\r\nZADD\r\n$11\r\nkey:[1:2:]3\r\n:123456\r\n:123\r\n:23456\r\n:2345\r\n", // nolint:lll
					command.New([]string{"ZADD", "key:[1:2:]3", "123456", "123", "23456", "2345"}),     // nolint:lll
					true,
				)
			})
		})
		// nolint:dupl
		t.Run("Sadd", func(t *testing.T) {
			t.Run("StringValue", func(t *testing.T) {
				testReadCommand(
					t,
					"*6\r\n$4\r\nSADD\r\n$9\r\nkey:1:2:3\r\n:123456\r\n$5\r\nID123\r\n:23456\r\n$6\r\nID2345\r\n", // nolint:lll
					command.New([]string{"SADD", "key:1:2:3", "123456", "ID123", "23456", "ID2345"}),              // nolint:lll
					true,
				)
			})
			t.Run("IntegerValue", func(t *testing.T) {
				testReadCommand(
					t,
					"*6\r\n$4\r\nSADD\r\n$11\r\nkey:[1:2:]3\r\n:123456\r\n:123\r\n:23456\r\n:2345\r\n", // nolint:lll
					command.New([]string{"SADD", "key:[1:2:]3", "123456", "123", "23456", "2345"}),     // nolint:lll
					true,
				)
			})
		})
		// nolint:dupl
		t.Run("Set", func(t *testing.T) {
			t.Run("StringValue", func(t *testing.T) {
				testReadCommand(
					t,
					"*3\r\n$3\r\nSET\r\n$9\r\nkey:1:2:3\r\n$8\r\nID123456\r\n",
					command.New([]string{"SET", "key:1:2:3", "ID123456"}),
					true,
				)
			})
			t.Run("IntegerValue", func(t *testing.T) {
				testReadCommand(
					t,
					"*3\r\n$3\r\nSET\r\n$9\r\nkey:1:2:3\r\n:123456\r\n",
					command.New([]string{"SET", "key:1:2:3", "123456"}),
					true,
				)
			})
		})
	})
	t.Run("Error", func(t *testing.T) {

	})
}

// testReadCommand проверяет правильное чтение команды
func testReadCommand(
	t *testing.T,
	s string,
	expectedCommand command.Command,
	success bool,
) {
	r := NewStringReader(s)
	cmd, err := r.Command()
	if success {
		if err != nil {
			t.Fatalf("read command error: %q", err)
		}
		if !reflect.DeepEqual(cmd, expectedCommand) {
			t.Fatalf("expected %q but actual %q", expectedCommand, cmd)
		}
	} else {
		if err == nil {
			t.Fatalf("expected error")
		}
	}
}

// testReadFileCommand проверяет правильное чтение команды из файла дампа
// nolint:unused
func testReadFileCommand(
	t *testing.T,
	filename string,
	expectedCommand command.Command,
	success bool,
) {
	r, file, err := NewFileReader(filename)
	if err != nil {
		t.Fatalf("open file error: %v", err)
	}
	defer file.Close() //nolint:errcheck
	cmd, err := r.Command()
	if success {
		if err != nil {
			t.Fatalf("read command error: %v", err)
		}
		if !reflect.DeepEqual(cmd, expectedCommand) {
			t.Fatalf("expected %q but actual %q", expectedCommand, cmd)
		}
	} else {
		if err == nil {
			t.Fatalf("expected error")
		}
	}
}
