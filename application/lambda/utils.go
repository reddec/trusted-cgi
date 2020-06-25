package lambda

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func tarFiles(dir string, out io.Writer, excludeGlob []string) error {
	writer := tar.NewWriter(out)
	defer writer.Close()
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		for _, pat := range excludeGlob {
			if len(pat) != 0 {
				if ok, _ := filepath.Match(pat, rel); ok {
					return nil
				}
			}
		}
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = rel
		err = writer.WriteHeader(header)
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(writer, f)
		return err
	})
}

func untarFiles(src io.Reader, dest string) error {
	reader := tar.NewReader(src)
	dirs := make(map[string]bool)

	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		path := filepath.Join(dest, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			err = os.MkdirAll(path, header.FileInfo().Mode())
			if err != nil {
				return fmt.Errorf("create dir %s: %w", header.Name, err)
			}
			dirs[path] = true
		case tar.TypeReg:
			dir := filepath.Dir(path)
			if !dirs[dir] {
				err = os.MkdirAll(dir, header.FileInfo().Mode())
				if err != nil {
					return fmt.Errorf("create dir %s: %w", header.Name, err)
				}
				dirs[path] = true
			}
			f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, header.FileInfo().Mode())
			if err != nil {
				return fmt.Errorf("create file %s: %w", header.Name, err)
			}
			_, err = io.Copy(f, reader)
			if err != nil {
				_ = f.Close()
				return fmt.Errorf("finish file %s: %w", header.Name, err)
			}
			err = f.Close()
			if err != nil {
				return fmt.Errorf("close file %s: %w", header.Name, err)
			}
		default:
			return fmt.Errorf("unsupported type of file %s %v", header.Name, header.Typeflag)
		}
	}
	return nil
}
