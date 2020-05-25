package memlog

import (
	"github.com/reddec/trusted-cgi/stats"
	"github.com/tinylib/msgp/msgp"
	"io/ioutil"
	"os"
	"path/filepath"
)

func NewDumped(filename string, depth uint) (*dumped, error) {
	d := &dumped{
		filename: filename,
		mem:      New(depth),
	}

	err := d.readDump()
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return d, nil
}

type dumped struct {
	filename string
	mem      *statLogger
}

func (d *dumped) readDump() error {
	f, err := os.Open(d.filename)
	if err != nil {
		return err
	}
	defer f.Close()

	reader := msgp.NewReader(f)
	n, err := reader.ReadArrayHeader()
	if err != nil {
		return err
	}
	for i := 0; i < int(n); i++ {
		var item stats.Record
		err = item.DecodeMsg(reader)
		if err != nil {
			return err
		}
		d.mem.Track(item)
	}
	return nil
}

// Make atomic (fs by rename) dump
func (d *dumped) Dump() error {
	tmp, err := ioutil.TempFile(filepath.Dir(d.filename), "dump.*")
	if err != nil {
		return err
	}

	clone := d.mem.buffer.Flatten()
	writer := msgp.NewWriter(tmp)
	err = writer.WriteArrayHeader(uint32(len(clone)))
	if err != nil {
		_ = tmp.Close()
		_ = os.RemoveAll(tmp.Name())
		return err
	}

	for _, item := range clone {
		err = item.EncodeMsg(writer)
		if err != nil {
			_ = tmp.Close()
			_ = os.RemoveAll(tmp.Name())
			return err
		}
	}

	err = writer.Flush()
	if err != nil {
		_ = tmp.Close()
		_ = os.RemoveAll(tmp.Name())
		return err
	}

	err = tmp.Close()
	if err != nil {
		_ = os.RemoveAll(tmp.Name())
		return err
	}

	// atomically rename file to latest dump
	return os.Rename(tmp.Name(), d.filename)
}

func (d *dumped) Track(record stats.Record) {
	d.mem.Track(record)
}

func (d *dumped) LastByUID(uid string, limit int) ([]stats.Record, error) {
	return d.mem.LastByUID(uid, limit)
}

func (d *dumped) Last(limit int) ([]stats.Record, error) {
	return d.mem.Last(limit)
}
