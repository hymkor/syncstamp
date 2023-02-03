package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/hymkor/syncstamp/dupfile"
	"github.com/nyaosorg/go-windows-mbcs"
)

func mains(args []string, printer func(string) error) error {
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
				err := printer(fmt.Sprintf("rem \"%s\"", file1.Path))
				if err != nil {
					return err
				}
			}
		}
		if err := printer(""); err != nil {
			return err
		}
	}
	return nil
}

var flagTee = flag.String("tee", "", "filename to tee output")

func main() {
	flag.Parse()
	printer := func(line string) error { fmt.Println(line); return nil }
	if *flagTee != "" {
		fd, err := os.Create(*flagTee)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		defer fd.Close()
		printer = func(line string) error {
			fmt.Println(line)
			sjis, err := mbcs.UtoA(line, mbcs.ACP)
			if err != nil {
				return err
			}
			fd.Write(sjis)
			fd.Write([]byte{'\r', '\n'})
			return nil
		}
	}
	if err := mains(flag.Args(), printer); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
