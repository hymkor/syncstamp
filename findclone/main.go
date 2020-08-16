package main

import (
	"fmt"
	"os"

	"github.com/zat-kaoru-hayama/syncstamp/dupfile"
)

func mains(args []string) error {
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
				fmt.Printf("rem \"%s\"\r\n", file1.Path)
			}
		}
		fmt.Println()
	}
	return nil
}

func main() {
	if err := mains(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
