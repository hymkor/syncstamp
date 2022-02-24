package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/zat-kaoru-hayama/syncstamp/dupfile"
)

func findSameFileButTimeDiff(srcFiles []*dupfile.File, dstFile *dupfile.File) (*dupfile.File, error) {
	for _, srcFile := range srcFiles {
		if srcFile.Sametime(dstFile) {
			continue
		}
		equal, err := srcFile.Equal(dstFile)
		if err != nil {
			return nil, err
		}
		if equal {
			return srcFile, nil
		}
	}
	return nil, nil
}

var flagBatch = flag.Bool("batch", false, "output batchfile to stdout")

var flagUpdate = flag.Bool("update", false, "update destinate-file's timestamp same as source-file's one")

func mains(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("Usage: %s <SRC-DIR> <DST-DIR>", os.Args[0])
	}

	srcRoot := args[0]
	dstRoot := args[1]
	dstCount := 0
	updCount := 0

	source, srcCount, err := dupfile.GetTree(srcRoot)
	if err != nil {
		return err
	}

	err = dupfile.Walk(dstRoot, func(key *dupfile.Key, val *dupfile.File) error {
		dstCount++

		srcFiles, ok := source[*key]
		if !ok {
			return nil
		}

		matchSrcFile, err := findSameFileButTimeDiff(
			srcFiles,
			val)
		if err != nil {
			return err
		}
		if matchSrcFile == nil {
			return nil
		}
		if *flagBatch {
			fmt.Printf("touch -r \"%s\" \"%s\"\n",
				matchSrcFile.Path,
				val.Path)
		} else {
			fmt.Printf("   %s %s\n",
				matchSrcFile.ModTime().Format("2006/01/02 15:04:05"), matchSrcFile.Path)
			if *flagUpdate {
				fmt.Print("->")
			} else {
				fmt.Print("!=")
			}

			fmt.Printf(" %s %s\n\n",
				val.ModTime().Format("2006/01/02 15:04:05"), val.Path)

			if *flagUpdate {
				err := os.Chtimes(val.Path,
					matchSrcFile.ModTime(),
					matchSrcFile.ModTime())
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s: %w\n", val.Path, err)
				} else {
					updCount++
				}
			}
		}
		return nil
	})
	fmt.Fprintf(os.Stderr, "    Read %4d files on %s.\n", srcCount, srcRoot)
	fmt.Fprintf(os.Stderr, "Compared %4d files on %s.\n", dstCount, dstRoot)
	if updCount > 0 {
		fmt.Fprintf(os.Stderr, " Touched %4d files on %s.\n", updCount, dstRoot)
	}
	fmt.Fprintf(os.Stderr, "    Open %4d files.\n", dupfile.OpenCount())
	return err
}

func main() {
	flag.Parse()
	if err := mains(flag.Args()); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
