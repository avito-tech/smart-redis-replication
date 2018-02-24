package rdb

import (
	"fmt"
	"io"
	"os"

	"github.com/avito-tech/smart-redis-replication/data"
)

const (
	tokenLevelStart = iota
	tokenLevelInit
	tokenLevelDB
)

// decoder реализует интерфейс Decoder
type decoder struct {
	r               Reader
	tokenLevelState int
	file            *os.File
}

// NewDecoder возвращает новый Decoder
func NewDecoder(r io.Reader) Decoder {
	return &decoder{
		r: NewReader(r),
	}
}

// NewStringDecoder возвращает новый Decoder на основании строки
func NewStringDecoder(data string) Decoder {
	return &decoder{
		r: NewStringReader(data),
	}
}

// NewLimitDecoder возвращает новый Decoder ограниченный по размеру
func NewLimitDecoder(r io.Reader, size int64) Decoder {
	return &decoder{
		r: NewReader(io.LimitReader(r, size)),
	}
}

// NewFileDecoder возвращает новый Decoder на основании файла,
// файл закрывается в конце Decode
// Для возможности преждевременного закрытия файла воспользуйтесь NewDecoder
// передав в него файл
func NewFileDecoder(filename string) (Decoder, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	return &decoder{
		r:    NewReader(file),
		file: file,
	}, nil
}

// nolint:gocyclo
func (d *decoder) DecodeKeys(consumer KeyConsumer) error {
	var db uint32
	for {
		token, err := d.Next()
		if err == io.EOF {
			_, ok := token.(EOF)
			if ok {
				err = nil
			} else {
				err = fmt.Errorf(`unexpected io.EOF, actual token "%#v"`, token)
			}
			return err
		}
		if err != nil {
			return fmt.Errorf("error get next token: %q", err)
		}
		switch op := token.(type) {
		case Magic, AuxField, ResizeDB:
		case DBSelector:
			db = op.GetDBNumber()
		case data.Key:
			err = op.SetDB(int(db))
			if err == nil {
				err = consumer.Key(op)
			}
		default:
			return fmt.Errorf("unexpected token %#v", op)
		}
		if err != nil {
			return err
		}
	}
}

// nolint:gocyclo
func (d *decoder) Decode(consumer Consumer) error {
	if d.file != nil {
		defer func() {
			_ = d.file.Close()
		}()
	}
	var db uint32
	for {
		token, err := d.Next()
		if err == io.EOF {
			eof, ok := token.(EOF)
			if ok {
				err = consumer.SetEOF(eof)
			} else {
				err = fmt.Errorf(`unexpected io.EOF, actual token "%#v"`, token)
			}
			return err
		}
		if err != nil {
			return fmt.Errorf("error get next token: %q", err)
		}
		switch op := token.(type) {
		case Magic:
			err = consumer.SetMagic(op)
		case AuxField:
			err = consumer.SetAuxField(op)
		case DBSelector:
			db = op.GetDBNumber()
		case ResizeDB:
			err = consumer.SetResizeDB(db, op)
		case data.Key:
			err = op.SetDB(int(db))
			if err == nil {
				err = consumer.Key(op)
			}
		default:
			return fmt.Errorf("unexpected token %#v", op)
		}
		if err != nil {
			return err
		}
	}
}

// checkTokenLevelState проверяет что токен находится в определённом уровне
// вложенности в RDB файле
func (d *decoder) checkTokenLevelState(tokenLevels ...int) error {
	for _, tokenLevel := range tokenLevels {
		if tokenLevel == d.tokenLevelState {
			return nil
		}
	}
	return fmt.Errorf(
		"expected token level %q but actual %d",
		tokenLevels,
		d.tokenLevelState,
	)
}

// nolint:gocyclo
func (d *decoder) Next() (interface{}, error) {
	if d.tokenLevelState == tokenLevelStart {
		d.tokenLevelState = tokenLevelInit
		return d.r.ReadMagic()
	}

	opcode, err := d.r.ReadOpcode()
	if err != nil {
		return nil, err
	}
	switch opcode {
	case AuxFieldOpcode:
		err = d.checkTokenLevelState(tokenLevelInit)
		if err != nil {
			return nil, err
		}
		return d.r.ReadAuxField()
	case DBSelectorOpcode:
		err = d.checkTokenLevelState(tokenLevelInit, tokenLevelDB)
		if err != nil {
			return nil, err
		}
		d.tokenLevelState = tokenLevelDB
		return d.r.ReadDBSelector()
	case ResizeDBOpcode:
		err = d.checkTokenLevelState(tokenLevelDB)
		if err != nil {
			return nil, err
		}
		return d.r.ReadResizeDB()
	case ExpirySecondsOpcode, ExpiryMillisecondsOpcode:
		err = d.checkTokenLevelState(tokenLevelDB)
		if err != nil {
			return nil, err
		}
		var expiry data.Expiry
		expiry, err = d.r.ReadExpiry(opcode)
		if err != nil {
			return nil, err
		}

		var opcodeNext byte
		opcodeNext, err = d.r.ReadByte()
		if err != nil {
			return nil, err
		}
		return d.readKey(opcodeNext, expiry)
	case EOFOpcode:
		err = d.checkTokenLevelState(tokenLevelInit, tokenLevelDB)
		if err != nil {
			return nil, err
		}
		return d.r.ReadEOF()
	}
	err = d.checkTokenLevelState(tokenLevelDB)
	if err != nil {
		return nil, err
	}
	return d.readKey(opcode, data.NewExpiry(0))
}

// nolint:gocyclo
func (d *decoder) readKey(opcode byte, expiry data.Expiry) (data.Key, error) {
	switch opcode {
	// SortedSet
	case ZipListSortedSetOpcode:
		return d.r.ReadZipListSortedSet(expiry)
	case SortedSetOpcode:
		return d.r.ReadSortedSet(expiry)

	// HashMap
	case ListHashMapOpcode:
		return d.r.ReadListHashMap(expiry)
	case ZipListHashMapOpcode:
		return d.r.ReadZipListHashMap(expiry)
	case ZipMapHashMapOpcode:
		return d.r.ReadZipMapHashMap(expiry)

	// List
	case ListOpcode:
		return d.r.ReadList(expiry)
	case ZipListOpcode:
		return d.r.ReadZipList(expiry)
	case QuickListOpcode:
		return d.r.ReadQuickList(expiry)

	case SetOpcode:
		return d.r.ReadSet(expiry)
	case IntSetOpcode:
		return d.r.ReadIntSet(expiry)
	case StringValueOpcode:
		return d.r.ReadStringValue(expiry)
	}
	return nil, fmt.Errorf("unsupported key opcode: %#v", opcode)
}
