package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"

	"github.com/grandecola/smaps/smaps"
)

func main() {
	pidVar := flag.Int("pid", 0, "process pid to compute mem usage for (default this pid)")
	filterVar := flag.String("filter", "", "filter mapped files using regular expression")
	flag.Parse()

	if *pidVar == 0 {
		*pidVar = os.Getpid()
	}
	sf, err := smaps.ReadSmaps(*pidVar, *filterVar)
	if err != nil {
		log.Fatal(err)
	}

	printSmapsInfo(sf)
	printTop10Maps(sf)
}

func printSmapsInfo(sf *smaps.ProcInfo) {
	fmt.Println("Summary:")
	fmt.Printf("  Total mappings: %v\n", sf.Count)
	fmt.Printf("  Total size: %v\n", toStringMemory(sf.Total))
	fmt.Printf("  Total RSS: %v\n", toStringMemory(sf.RSS))
	fmt.Printf("  Total PSS: %v\n", toStringMemory(sf.PSS))
}

func printTop10Maps(sf *smaps.ProcInfo) {
	fmt.Println("Top 10 mappings:")
	sort.Slice(sf.Maps, func(i, j int) bool {
		return sf.Maps[i].PSS > sf.Maps[j].PSS
	})
	for i, mf := range sf.Maps[:min(10, len(sf.Maps))] {
		fmt.Printf("  %v. {%v} PSS: %v, RSS: %v, Size: %v\n", i+1, mf.Name, toStringMemory(mf.PSS),
			toStringMemory(mf.RSS), toStringMemory(mf.Size))
	}
}

func toStringMemory(m uint64) string {
	switch {
	case m > 1024*1024*1024:
		return strconv.Itoa(int(m)/1024/1024/1024) + " GB"
	case m > 1024*1024:
		return strconv.Itoa(int(m/1024/1024)) + " MB"
	case m > 1024:
		return strconv.Itoa(int(m/1024)) + " KB"
	default:
		return strconv.Itoa(int(m)) + " Bytes"
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}

	return b
}
