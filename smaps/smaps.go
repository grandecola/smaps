package smaps

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
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

// ProcInfo stores various aggregate information computed
// from /proc/<pid>/smaps file for a process.
type ProcInfo struct {
	Count uint64
	RSS   uint64
	PSS   uint64
	Total uint64
	Maps  []*MapInfo
}

// MapInfo stores information about one mapping.
type MapInfo struct {
	Name string
	RSS  uint64
	PSS  uint64
	Size uint64
}

// ReadSmaps reads the /proc/<pid>/smaps file and stores the information in a struct.
func ReadSmaps(pid int, filter string) (*ProcInfo, error) {
	var re *regexp.Regexp
	if filter != "" {
		var err error
		re, err = regexp.Compile(filter)
		if err != nil {
			return nil, fmt.Errorf("error in compiling regex [%v] :: %w", filter, err)
		}
	}

	fp := fmt.Sprintf("/proc/%v/smaps", pid)
	data, err := ioutil.ReadFile(fp)
	if err != nil {
		return nil, fmt.Errorf("error in reading file: %v :: %w", fp, err)
	}

	sf := new(ProcInfo)
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
			sf.Maps = append(sf.Maps, mf)
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

func readMapping(scanner *bufio.Scanner) (*MapInfo, error) {
	info := new(MapInfo)

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
