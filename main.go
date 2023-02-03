package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/hymkor/syncstamp/internal/dupfile"
)

var (
	flagBatch  = flag.Bool("batch", false, "output batchfile to stdout")
	flagDryRun = flag.Bool("dry-run", false, "dry-run")
	flagOldest = flag.Bool("oldest", false, "update timestamps of source and destinate files all same as the oldest one")
	flagUpdate = flag.Bool("update", false, "update destinate-file's timestamp same as source-file's one")
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

func touch(path string, stamp time.Time, count *int) {
	var err error
	if !*flagDryRun {
		err = os.Chtimes(path, stamp, stamp)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %w\n", path, err)
	} else {
		fmt.Printf("touch -r %s \"%s\"\n", stamp.Format("200601021504.05"), path)
		if count != nil {
			*count++
		}
	}
}

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

			if *flagOldest {
				stamp := val.ModTime()
				for _, p := range srcFiles {
					if tm := p.ModTime(); tm.Before(stamp) {
						stamp = tm
					}
				}
				if val.ModTime() != stamp {
					touch(val.Path, stamp, &updCount)
				}
				for _, p := range srcFiles {
					if p.ModTime() != stamp {
						touch(p.Path, stamp, &updCount)
					}
				}
			} else if *flagUpdate {
				touch(val.Path, diffTimes[0].ModTime(), &updCount)
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
