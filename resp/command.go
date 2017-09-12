package resp

import (
	"fmt"
	"io"

	"go.avito.ru/gl/smart-redis-replication/command"
)

// Command читает и возвращает следующую команду
func (r *reader) Command() (command.Command, error) {
	opcode, err := r.ReadOpcode()
	if err != nil {
		if err == io.EOF {
			err = nil
		}
		return command.Command{}, err
	}
	switch opcode {
	case LF:
		return r.ReadLFCommand()

	case BulkStringOpcode:
		return r.ReadRDBCommand()

	case ArrayOpcode:
		return r.ReadArrayCommand()

	case SimpleStringOpcode:
		return r.IgnoreSimpleStringCommand()

	case IntegerOpcode:
		return r.IgnoreIntegerCommand()

	case ErrorOpcode:
		return r.ReadErrorCommand()
	}
	return command.Command{}, fmt.Errorf("unexpected opcode %#v", opcode)
}

func (r *reader) ReadLFCommand() (command.Command, error) {
	return command.Command{}, nil
}

func (r *reader) ReadRDBCommand() (command.Command, error) {
	size, err := r.ReadInteger()
	if err != nil {
		return command.Command{}, err
	}
	return command.New([]string{
		"rdb",
		fmt.Sprintf("%d", size),
	}), nil
}

func (r *reader) ReadArrayCommand() (cmd command.Command, err error) {
	r.StartDump("ArrayCommand")
	defer r.StopDump(&err)

	var args []string
	args, err = r.ReadArray()
	if err != nil {
		return command.Command{}, err
	}
	return command.New(args), nil
}

func (r *reader) ReadSimpleStringCommand() (command.Command, error) {
	s, err := r.ReadString('\n')
	return command.New([]string{s}), err
}

func (r *reader) IgnoreSimpleStringCommand() (command.Command, error) {
	_, err := r.ReadSimpleStringCommand()
	return command.Command{}, err
}

func (r *reader) IgnoreIntegerCommand() (command.Command, error) {
	_, err := r.ReadInteger()
	return command.Command{}, err
}

func (r *reader) ReadErrorCommand() (command.Command, error) {
	dataError, err := r.ReadError()
	if err == nil {
		err = fmt.Errorf("server error: %v", dataError)
	}
	return command.Command{}, err
}
