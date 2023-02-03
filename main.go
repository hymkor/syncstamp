package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/hymkor/syncstamp/internal/dupfile"
)

func findSameFileButTimeDiff(srcFiles []*dupfile.File, dstFile *dupfile.File) (sameTimes, diffTimes []*dupfile.File, err error) {
	for _, srcFile := range srcFiles {
		if srcFile.Sametime(dstFile) {
			sameTimes = append(sameTimes, srcFile)
			continue
		}
		equal, _err := srcFile.Equal(dstFile)
		if err != nil {
			return nil, nil, _err
		}
		if equal {
			diffTimes = append(diffTimes, srcFile)
		}
	}
	return
}

var flagBatch = flag.Bool("batch", false, "output batchfile to stdout")

var flagUpdate = flag.Bool("update", false, "update destinate-file's timestamp same as source-file's one")

func mains(args []string) error {
	if len(args) < 2 {
		fmt.Printf("Usage: %s <SRC-DIR> <DST-DIR>\n", os.Args[0])
		flag.PrintDefaults()
		return nil
	}

	srcRoot := args[0]
	dstRoot := args[1]
	dstCount := 0
	updCount := 0

	source := dupfile.NewTree()
	srcCount, err := source.Read(srcRoot)
	if err != nil {
		return err
	}

	err = dupfile.Walk(dstRoot, func(key *dupfile.Key, val *dupfile.File) error {
		dstCount++

		srcFiles, ok := source[*key]
		if !ok {
			return nil
		}

		sameTimes, diffTimes, err := findSameFileButTimeDiff(
			srcFiles,
			val)
		if err != nil {
			return err
		}
		if diffTimes == nil || len(diffTimes) <= 0 {
			return nil
		}
		if *flagBatch {
			fmt.Printf("touch -r \"%s\" \"%s\"\n",
				diffTimes[0].Path,
				val.Path)
		} else {
			fmt.Printf("   %s %s\n",
				val.ModTime().Format("2006/01/02 15:04:05"), val.Path)

			for _, s := range sameTimes {
				fmt.Printf("== %s %s\n",
					s.ModTime().Format("2006/01/02 15:04:05"), s.Path)
			}
			for _, d := range diffTimes {
				fmt.Printf("!= %s %s\n",
					d.ModTime().Format("2006/01/02 15:04:05"), d.Path)
			}

			if *flagUpdate {
				newTime := diffTimes[0].ModTime()
				err := os.Chtimes(val.Path, newTime, newTime)
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s: %w\n", val.Path, err)
				} else {
					fmt.Printf("touch -r %s \"%s\"\n",
						newTime.Format("200601021504.05"), val.Path)
					updCount++
				}
			}
			fmt.Println()
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
