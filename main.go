package main

import (
	"bytes"
	"crypto/md5"
	"flag"
	"fmt"
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

func findSameFileButTimeDiff(srcFiles []*File, dstFile *File) (*File, error) {
	dstTime := dstFile.ModTime().Truncate(time.Second)
	for _, srcFile := range srcFiles {
		srcTime := srcFile.ModTime().Truncate(time.Second)
		if srcTime.Equal(dstTime) {
			continue
		}
		srcHash, err := srcFile.Hash()
		if err != nil {
			return nil, err
		}
		dstHash, err := dstFile.Hash()
		if err != nil {
			return nil, err
		}
		if bytes.Equal(srcHash, dstHash) {
			return srcFile, nil
		}
	}
	return nil, nil
}

var flagBatch = flag.Bool("batch", false, "output batchfile to stdout")

var flagUpdate = flag.Bool("update", false, "update destinate-file's timestamp same as source-file's one")

type keyT struct {
	Name string
	Size int64
}

func mains(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("Usage: %s <SRC-DIR> <DST-DIR>", os.Args[0])
	}

	srcRoot := args[0]
	if path, err := filepath.EvalSymlinks(srcRoot); err == nil {
		srcRoot = path
	}
	source := map[keyT][]*File{}
	srcCount := 0
	err := filepath.Walk(srcRoot, func(path string, srcFile os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if srcFile.IsDir() {
			if name := filepath.Base(path); name[0] == '.' {
				return filepath.SkipDir
			}
			return nil
		}
		key := keyT{
			Name: strings.ToUpper(filepath.Base(path)),
			Size: srcFile.Size(),
		}
		entry := &File{Path: path, FileInfo: srcFile}
		source[key] = append(source[key], entry)
		srcCount++
		return nil
	})
	if err != nil {
		return err
	}

	dstRoot := args[1]
	if path, err := filepath.EvalSymlinks(dstRoot); err == nil {
		dstRoot = path
	}
	dstCount := 0
	updCount := 0
	err = filepath.Walk(dstRoot, func(path string, dstFile os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if dstFile.IsDir() {
			if name := filepath.Base(path); name[0] == '.' {
				return filepath.SkipDir
			}
			return nil
		}
		dstCount++

		key := keyT{
			Name: strings.ToUpper(filepath.Base(path)),
			Size: dstFile.Size(),
		}

		srcFiles, ok := source[key]
		if !ok {
			return nil
		}

		matchSrcFile, err := findSameFileButTimeDiff(
			srcFiles,
			&File{Path: path, FileInfo: dstFile})

		if err != nil {
			return err
		}
		if matchSrcFile == nil {
			return nil
		}
		if *flagBatch {
			fmt.Printf("touch -r \"%s\" \"%s\"\n",
				matchSrcFile.Path,
				path)
		} else {
			fmt.Printf("   %s %s\n",
				matchSrcFile.ModTime().Format("2006/01/02 15:04:05"), matchSrcFile.Path)
			if *flagUpdate {
				fmt.Print("->")
			} else {
				fmt.Print("!=")
			}

			fmt.Printf(" %s %s\n\n",
				dstFile.ModTime().Format("2006/01/02 15:04:05"), path)

			if *flagUpdate {
				os.Chtimes(path,
					matchSrcFile.ModTime(),
					matchSrcFile.ModTime())
				updCount++
			}
		}
		return nil
	})
	fmt.Fprintf(os.Stderr, "    Read %4d files on %s.\n", srcCount, srcRoot)
	fmt.Fprintf(os.Stderr, "Compared %4d files on %s.\n", dstCount, dstRoot)
	if updCount > 0 {
		fmt.Fprintf(os.Stderr, " Touched %4d files on %s.\n", updCount, dstRoot)
	}
	fmt.Fprintf(os.Stderr, "    Open %4d files.\n", openCount)
	return err
}

func main() {
	flag.Parse()
	if err := mains(flag.Args()); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
