package smaps

import (
	"os"
	"strings"
	"syscall"
	"testing"

	"github.com/grandecola/mmap"
)

var (
	protPage = syscall.PROT_READ | syscall.PROT_WRITE
	testPath = "/tmp/m.txt"
	size     = 16 * 1024
)

func TestReadMaps(t *testing.T) {
	defer func() {
		if err := os.Remove(testPath); err != nil {
			t.Fatalf("error in deleting file :: %v", err)
		}
	}()

	fd, err := os.OpenFile(testPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		t.Fatalf("error in opening file :: %v", err)
	}
	defer func() {
		if err := fd.Close(); err != nil {
			t.Fatalf("error in closing the file :: %v", err)
		}
	}()

	if err := os.Truncate(testPath, int64(size)); err != nil {
		t.Fatalf("error in truncating the file :: %v", err)
	}

	aa, err := mmap.NewSharedFileMmap(fd, 0, size, protPage)
	if err != nil {
		t.Fatalf("error in mapping file :: %v", err)
	}
	defer func() {
		if err := aa.Unmap(); err != nil {
			t.Fatalf("error in un-mapping the file :: %v", err)
		}
	}()

	// Nothing in memory yet
	pid := os.Getpid()
	pi, err := ReadSmaps(pid, "")
	if err != nil {
		t.Fatalf("error in reading smaps file :: %v", err)
	}

	for _, mf := range pi.Maps {
		if !strings.Contains(mf.Name, "m.txt") {
			continue
		}

		if mf.Size != uint64(size) || mf.RSS != 0 || mf.PSS != 0 {
			t.Fatalf("unexpected mapping values :: %+v", mf)
		}
	}

	// Only the first page in memory
	_, _ = aa.WriteAt([]byte("aman"), 0)
	pi, err = ReadSmaps(pid, "m.txt")
	if err != nil {
		t.Fatalf("error in reading smaps file :: %v", err)
	}
	if len(pi.Maps) != 1 {
		t.Fatalf("unexpected number of mappings")
	}
	if pi.Maps[0].Size != uint64(size) || pi.Maps[0].RSS != 4*1024 || pi.Maps[0].PSS != 4*1024 {
		t.Fatalf("unexpected mapping values :: %+v", pi.Maps[0])
	}

	// Touch another page
	_, _ = aa.WriteAt([]byte("aman"), 8*1024)
	pi, err = ReadSmaps(pid, "m.txt")
	if err != nil {
		t.Fatalf("error in reading smaps file :: %v", err)
	}
	if len(pi.Maps) != 1 {
		t.Fatalf("unexpected number of mappings")
	}
	if pi.Maps[0].Size != uint64(size) || pi.Maps[0].RSS != 8*1024 || pi.Maps[0].PSS != 8*1024 {
		t.Fatalf("unexpected mapping values :: %+v", pi.Maps[0])
	}

	// Open the same file again
	aa2, err := mmap.NewSharedFileMmap(fd, 0, size, protPage)
	if err != nil {
		t.Fatalf("error in mapping file :: %v", err)
	}
	defer func() {
		if err := aa2.Unmap(); err != nil {
			t.Fatalf("error in un-mapping the file :: %v", err)
		}
	}()

	// and check memory usage
	pi, err = ReadSmaps(pid, "m.txt")
	if err != nil {
		t.Fatalf("error in reading smaps file :: %v", err)
	}
	if len(pi.Maps) != 2 {
		t.Fatalf("unexpected number of mappings")
	}
	if pi.Total != uint64(size)*2 || pi.RSS != 8*1024 || pi.PSS != 8*1024 {
		t.Fatalf("unexpected mapping values :: %+v", pi.Maps[0])
	}

	// Touch an already in mem page
	_, _ = aa2.WriteAt([]byte("aman"), 8*1024)
	pi, err = ReadSmaps(pid, "m.txt")
	if err != nil {
		t.Fatalf("error in reading smaps file :: %v", err)
	}
	if len(pi.Maps) != 2 {
		t.Fatalf("unexpected number of mappings")
	}
	if pi.Total != uint64(size)*2 || pi.RSS != 12*1024 || pi.PSS != 8*1024 {
		t.Fatalf("unexpected mapping values :: %+v", pi.Maps[0])
	}
}
