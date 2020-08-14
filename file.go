package main

import (
	"bytes"
	"crypto/md5"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type File struct {
	os.FileInfo
	Path string
	hash []byte
}

var openCount int = 0

func hash(path string) ([]byte, error) {
	h := md5.New()

	fd, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	io.Copy(h, fd)
	openCount++

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

func (this *File) Equal(other *File) (bool, error) {
	hash1, err := this.Hash()
	if err != nil {
		return false, err
	}
	hash2, err := other.Hash()
	if err != nil {
		return false, err
	}
	return bytes.Equal(hash1, hash2), nil
}

func (this *File) Sametime(other *File) bool {
	time1 := this.ModTime().Truncate(time.Second)
	time2 := other.ModTime().Truncate(time.Second)
	return time1.Equal(time2)
}

type keyT struct {
	Name string
	Size int64
}

func walk(root string, callback func(*keyT, *File) error) error {
	if path, err := filepath.EvalSymlinks(root); err == nil {
		root = path
	}
	return filepath.Walk(root, func(path string, file1 os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if file1.IsDir() {
			if name := filepath.Base(path); name[0] == '.' {
				return filepath.SkipDir
			}
			return nil
		}
		key := &keyT{
			Name: strings.ToUpper(filepath.Base(path)),
			Size: file1.Size(),
		}
		val := &File{Path: path, FileInfo: file1}
		return callback(key, val)
	})
}

func getTree(root string) (map[keyT][]*File, int, error) {
	files := map[keyT][]*File{}
	count := 0

	err := walk(root, func(key *keyT, value *File) error {
		files[*key] = append(files[*key], value)
		count++
		return nil
	})
	return files, count, err
}
