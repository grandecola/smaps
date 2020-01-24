package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const (
	cSizePrefix    = "Size:"
	cRssPrefix     = "Rss:"
	cPssPrefix     = "Pss:"
	cVMFlagsPrefix = "VmFlags:"
)

// SmapsInfo stores various aggregate information computed from /proc/<pid>/smaps file.
type SmapsInfo struct {
	Count uint64
	RSS   uint64
	PSS   uint64
	Total uint64
	maps  []*MapsInfo
}

// MapsInfo stores information about one mapping.
type MapsInfo struct {
	Name string
	RSS  uint64
	PSS  uint64
	Size uint64
}

func main() {
	pid := os.Getpid()
	pidVar := flag.Int("pid", pid, "process pid to compute mem usage for")
	filter := flag.String("filter", "", "filter mapped files using regular expression")
	flag.Parse()

	filePath := fmt.Sprintf("/proc/%v/smaps", *pidVar)
	sf, err := readSmaps(filePath, *filter)
	if err != nil {
		log.Fatal(err)
	}

	printSmapsInfo(sf)
	printTop10Maps(sf)
}

func readSmaps(fp string, filter string) (*SmapsInfo, error) {
	var re *regexp.Regexp
	if filter != "" {
		var err error
		re, err = regexp.Compile(filter)
		if err != nil {
			return nil, fmt.Errorf("error in compiling regex [%v] :: %w", filter, err)
		}
	}

	data, err := ioutil.ReadFile(fp)
	if err != nil {
		return nil, fmt.Errorf("error in reading file: %v :: %w", fp, err)
	}

	sf := new(SmapsInfo)
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		tokens := strings.Fields(line)

		switch {
		case re != nil && len(tokens) == 5:
		case re != nil && len(tokens) == 6 && !re.Match([]byte(tokens[5])):
			if err := skipMapping(scanner); err != nil {
				return nil, err
			}

		case re != nil && len(tokens) != 6:
			return nil, fmt.Errorf("unexpected first line [%v]", line)

		default:
			mf, err := readMapping(scanner)
			if err != nil {
				return nil, err
			}
			sf.maps = append(sf.maps, mf)
			sf.Count++
			sf.RSS += mf.RSS
			sf.PSS += mf.PSS
			sf.Total += mf.Size
		}
	}

	return sf, nil
}

func skipMapping(scanner *bufio.Scanner) error {
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, cVMFlagsPrefix) {
			return nil
		}
	}

	return fmt.Errorf("unexpected data in smaps file")
}

func readMapping(scanner *bufio.Scanner) (*MapsInfo, error) {
	info := new(MapsInfo)

	line := scanner.Text()
	tokens := strings.Fields(line)
	if len(tokens) != 6 {
		info.Name = "anonymous"
	} else {
		info.Name = tokens[5]
	}

	var vp *uint64
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, cVMFlagsPrefix):
			return info, nil
		case strings.HasPrefix(line, cSizePrefix):
			vp = &info.Size
		case strings.HasPrefix(line, cRssPrefix):
			vp = &info.RSS
		case strings.HasPrefix(line, cPssPrefix):
			vp = &info.PSS
		default:
			continue
		}

		val, err := parseMemory(line)
		if err != nil {
			return nil, err
		}
		*vp += val
	}

	return nil, fmt.Errorf("unexpected data in smaps file")
}

func parseMemory(line string) (uint64, error) {
	tokens := strings.Fields(line)
	if len(tokens) < 3 {
		return 0, fmt.Errorf("expected 3 tokens in line [%v]", line)
	}

	return toUintMemory(tokens[1], tokens[2])
}

func toUintMemory(val, str string) (uint64, error) {
	num, err := strconv.ParseUint(val, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("unable to parse value [%v] as uint64", val)
	}

	str = strings.ToLower(str)
	switch str {
	case "kb":
		return num * 1024, nil
	case "mb":
		return num * 1024 * 1024, nil
	case "gb":
		return num * 1024 * 1024 * 1024, nil
	default:
		return 0, fmt.Errorf("unable to parse value [%v]", str)
	}
}

func printSmapsInfo(sf *SmapsInfo) {
	fmt.Println("Summary:")
	fmt.Printf("  Total mappings: %v\n", sf.Count)
	fmt.Printf("  Total size: %v\n", toStringMemory(sf.Total))
	fmt.Printf("  Total RSS: %v\n", toStringMemory(sf.RSS))
	fmt.Printf("  Total PSS: %v\n", toStringMemory(sf.PSS))
}

func printTop10Maps(sf *SmapsInfo) {
	fmt.Println("Top 10 mappings:")
	sort.Slice(sf.maps, func(i, j int) bool {
		return sf.maps[i].PSS > sf.maps[j].PSS
	})
	for i, mf := range sf.maps[:min(10, len(sf.maps))] {
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
