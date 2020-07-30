package main

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type File struct {
	Path  string
	Stamp time.Time
	Size  int64
	hash  []byte
}

func hash(path string) ([]byte, error) {
	h := md5.New()

	fd, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	io.Copy(h, fd)

	return h.Sum(nil)[:], nil
}

func (f *File) Hash() ([]byte, error) {
	if f.hash == nil {
		var err error
		f.hash, err = hash(f.Path)
		if err != nil {
			return nil, err
		}
	}
	return f.hash, nil
}

func mains(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("Usage: %s <SRC-DIR> <DST-DIR>", os.Args[0])
	}

	srcPath := args[0]
	if path, err := filepath.EvalSymlinks(srcPath); err == nil {
		srcPath = path
	}
	source := map[string][]*File{}
	err := filepath.Walk(srcPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if name := filepath.Base(path); name[0] == '.' {
				return filepath.SkipDir
			}
			return nil
		}
		name := strings.ToUpper(filepath.Base(path))
		entry := &File{Path: path, Stamp: info.ModTime(), Size: info.Size()}
		s, ok := source[name]
		if ok {
			source[name] = append(s, entry)
		} else {
			source[name] = []*File{entry}
		}
		return nil
	})
	if err != nil {
		return err
	}

	dstPath := args[1]
	if path, err := filepath.EvalSymlinks(dstPath); err == nil {
		dstPath = path
	}
	err = filepath.Walk(dstPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if name := filepath.Base(path); name[0] == '.' {
				return filepath.SkipDir
			}
			return nil
		}
		name := strings.ToUpper(filepath.Base(path))

		s, ok := source[name]
		if !ok {
			return nil
		}
		var s1 *File
		var dstHash []byte
		for _, s2 := range s {
			if info.Size() != s2.Size {
				continue
			}
			if dstHash == nil {
				var err error
				dstHash, err = hash(path)
				if err != nil {
					return err
				}
			}
			hash2, err := s2.Hash()
			if err != nil {
				return err
			}
			if bytes.Equal(hash2, dstHash) {
				s1 = s2
				break
			}
		}
		if s1 == nil {
			return nil
		}
		srcTime := s1.Stamp.Truncate(time.Second)
		dstTime := info.ModTime().Truncate(time.Second)
		if !srcTime.Equal(dstTime) {
			fmt.Printf("\n   %s %s\n",
				srcTime.Format("2006/01/02 15:04:05"), s1.Path)
			fmt.Printf("-> %s %s\n",
				dstTime.Format("2006/01/02 15:04:05"), path)
		}
		return nil
	})
	return err
}

func main() {
	if err := mains(os.Args[1:]); err != nil {
		fmt.Println(os.Stderr, err.Error())
		os.Exit(1)
	}
}
