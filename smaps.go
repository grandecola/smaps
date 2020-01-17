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
}

func main() {
	pid := os.Getpid()
	pidVar := flag.Int("pid", pid, "process pid to compute mem usage for")
	filter := flag.String("filter", "", "filter mapped files using regular expression")
	flag.Parse()

	var re *regexp.Regexp
	if *filter != "" {
		var err error
		re, err = regexp.Compile(*filter)
		if err != nil {
			log.Fatalf("error in compiling regular expression [%v] :: %v", *filter, err)
		}
	}

	filePath := fmt.Sprintf("/proc/%v/smaps", *pidVar)
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatalf("error in reading file: %v :: %v", filePath, err)
	}

	var info SmapsInfo
	scanner := bufio.NewScanner(bytes.NewReader(data))
	first := true
	skip := false
	for scanner.Scan() {
		var vp *uint64
		var flag bool

		line := scanner.Text()
		switch {
		case first:
			first = false
			if re == nil {
				continue
			}
			tokens := strings.Fields(line)
			if len(tokens) == 5 {
				skip = true
				continue
			}
			if len(tokens) != 6 {
				log.Fatalf("unexpected first line [%v]", line)
			}
			if !re.Match([]byte(tokens[5])) {
				skip = true
			}
		case strings.HasPrefix(line, cVMFlagsPrefix):
			first = true
			skip = false
			continue
		case skip:
			continue
		case strings.HasPrefix(line, cSizePrefix):
			flag = true
			info.Count++
			vp = &info.Total
		case strings.HasPrefix(line, cRssPrefix):
			flag = true
			vp = &info.RSS
		case strings.HasPrefix(line, cPssPrefix):
			flag = true
			vp = &info.PSS
		}

		if flag {
			val, err := parseMemory(line)
			if err != nil {
				log.Fatal(err)
			}
			*vp += val
		}
	}

	fmt.Printf("%+v\n", info)
}

func parseMemory(line string) (uint64, error) {
	tokens := strings.Fields(line)
	if len(tokens) < 3 {
		return 0, fmt.Errorf("expected 3 tokens in line [%v]", line)
	}

	num, err := strconv.ParseUint(tokens[1], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("unable to parse value [%v] as uint64", tokens[1])
	}

	mul, err := findMultiplier(tokens[2])
	if err != nil {
		return 0, err
	}

	return num * mul, nil
}

func findMultiplier(s string) (uint64, error) {
	s = strings.ToLower(s)
	switch s {
	case "kb":
		return 1024, nil
	case "mb":
		return 1024 * 1024, nil
	case "gb":
		return 1024 * 1024 * 1024, nil
	default:
		return 0, fmt.Errorf("unable to parse value [%v]", s)
	}
}
