package internal

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

func ReadJson(filename string, target interface{}) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(target)
}

func AtomicWriteJson(filename string, config interface{}) error {
	// write to file with atomic swap
	tmp, err := ioutil.TempFile(filepath.Dir(filename), "")
	if err != nil {
		return err
	}
	enc := json.NewEncoder(tmp)
	enc.SetIndent("", "  ")
	err = enc.Encode(config)

	_ = tmp.Close()
	if err != nil {
		_ = os.Remove(tmp.Name())
		return err
	}

	err = os.Rename(tmp.Name(), filename) // swap temp file to target file
	if err != nil {
		_ = os.Remove(tmp.Name())
		return err
	}
	return nil
}
