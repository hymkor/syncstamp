package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/zat-kaoru-hayama/syncstamp/dupfile"
)

func mains(args []string, w io.Writer) error {
	files := map[dupfile.Key][]*dupfile.File{}
	count := 0
	for _, arg1 := range args {
		count1, err := dupfile.ReadTree(arg1, files)
		if err != nil {
			return err
		}
		count += count1
	}
	for _, sameNameSizeFiles := range files {
		if len(sameNameSizeFiles) <= 1 {
			continue
		}
		sameHashFiles := map[string][]*dupfile.File{}
		for _, file1 := range sameNameSizeFiles {
			hash1, err := file1.Hash()
			if err == nil {
				h := string(hash1)
				sameHashFiles[h] = append(sameHashFiles[h], file1)
			}
		}
		for _, dup1 := range sameHashFiles {
			if len(dup1) <= 1 {
				continue
			}
			for _, file1 := range dup1 {
				fmt.Fprintf(w, "rem \"%s\"\r\n", file1.Path)
			}
		}
		fmt.Fprintln(w)
	}
	return nil
}

var flagTee = flag.String("tee", "", "filename to tee output")

func main() {
	flag.Parse()
	var out io.Writer = os.Stdout
	if *flagTee != "" {
		fd, err := os.Create(*flagTee)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		defer fd.Close()
		out = io.MultiWriter(fd, os.Stdout)
	}
	if err := mains(flag.Args(), out); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
