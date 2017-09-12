package resp

import (
	"bytes"
	"fmt"
	"os"
	"time"
)

// Read это обёртка над bufio.Read для записи дампа
func (r *reader) Read(buf []byte) (int, error) {
	n, err := r.Reader.Read(buf)
	if r.debug && r.dump.start {
		r.dump.buf.Write(buf)
	}
	return n, err
}

// Read это обёртка над bufio.Read для записи дампа
func (r *reader) ReadByte() (byte, error) {
	b, err := r.Reader.ReadByte()
	if r.debug && r.dump.start {
		r.dump.buf.WriteByte(b)
	}
	return b, err
}

// Read это обёртка над bufio.Read для записи дампа
func (r *reader) ReadString(delim byte) (string, error) {
	s, err := r.Reader.ReadString(delim)
	if r.debug && r.dump.start {
		r.dump.buf.Write([]byte(s))
	}
	return s, err
}

// EnableDebug включает логирование resp команд
func (r *reader) EnableDebug(dir string) error {
	if !IsDir(dir) {
		err := os.MkdirAll(dir, os.ModeDir)
		if err != nil {
			return err
		}
	}
	r.debug = true
	r.dump.dir = dir

	buf := make([]byte, 0, 1024)
	r.dump.buf = bytes.NewBuffer(buf)
	return nil
}

// DisableDebug выключает логирование resp команд
func (r *reader) DisableDebug() {
	r.debug = false
}

// StartDump включает запись чтения из потока
func (r *reader) StartDump(name string) {
	if r.debug {
		r.dump.start = true
		r.dump.name = name
		r.dump.buf.Reset()
	}
}

// StopDump выключает запись чтения из потока
func (r *reader) StopDump(cmdError *error) {
	if r.debug && *cmdError != nil {
		b := make([]byte, 20)
		r.Read(b) //nolint:errcheck

		filename := fmt.Sprintf(
			"%s/%s-%s.dump",
			r.dump.dir,
			time.Now().Local().Format("2006-01-02"),
			r.dump.name,
		)
		r.saveDump(filename) //nolint:errcheck
		r.dump.start = false
		r.dump.buf.Reset()
	}
}

// saveDump сохраняет дамп в файл
func (r *reader) saveDump(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	_, err = file.Write(r.dump.buf.Bytes())
	if err != nil {
		return err
	}
	err = file.Sync()
	if err != nil {
		return err
	}
	err = file.Close()
	return err
}
