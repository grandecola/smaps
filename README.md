[![Build Status](https://travis-ci.com/grandecola/smaps.svg?branch=master)](https://travis-ci.com/grandecola/smaps) [![GoDoc](https://godoc.org/github.com/grandecola/smaps/smaps?status.svg)](https://godoc.org/github.com/grandecola/smaps/smaps) [![MIT license](http://img.shields.io/badge/license-MIT-brightgreen.svg)](http://opensource.org/licenses/MIT) [![Go Report Card](https://goreportcard.com/badge/github.com/grandecola/smaps)](https://goreportcard.com/report/github.com/grandecola/smaps) [![golangci](https://golangci.com/badges/github.com/grandecola/smaps.svg)](https://golangci.com/r/github.com/grandecola/smaps)

## smaps

Script and a library to compute memory usage of a process due to mmap (using `/proc/<pid>/smaps`)

## How to Use

`go run main.go --help`

```
Usage:
  -filter string
        filter mapped files using regular expression
  -pid int
        process pid to compute mem usage for (default this pid)
```

## Example Usage

### Print memory usage of current process

```
⇒  go run main.go
Summary:
  Total mappings: 30
  Total size: 687 MB
  Total RSS: 2 MB
  Total PSS: 2 MB
Top 10 mappings:
  1. {/tmp/go-build158759953/b001/exe/main} PSS: 828 KB, RSS: 828 KB, Size: 828 KB
  2. {/tmp/go-build158759953/b001/exe/main} PSS: 768 KB, RSS: 768 KB, Size: 968 KB
  3. {anonymous} PSS: 200 KB, RSS: 200 KB, Size: 300 KB
  4. {/tmp/go-build158759953/b001/exe/main} PSS: 96 KB, RSS: 96 KB, Size: 96 KB
  5. {anonymous} PSS: 80 KB, RSS: 80 KB, Size: 63 MB
  6. {anonymous} PSS: 76 KB, RSS: 76 KB, Size: 108 KB
  7. {anonymous} PSS: 60 KB, RSS: 60 KB, Size: 35 MB
  8. {anonymous} PSS: 52 KB, RSS: 52 KB, Size: 384 KB
  9. {anonymous} PSS: 40 KB, RSS: 40 KB, Size: 180 KB
  10. {anonymous} PSS: 32 KB, RSS: 32 KB, Size: 60 KB
```

### Filter mappings using regex

```
⇒  go run main.go -filter "main"
Summary:
  Total mappings: 3
  Total size: 1 MB
  Total RSS: 1 MB
  Total PSS: 1 MB
Top 10 mappings:
  1. {/tmp/go-build121173397/b001/exe/main} PSS: 828 KB, RSS: 828 KB, Size: 828 KB
  2. {/tmp/go-build121173397/b001/exe/main} PSS: 772 KB, RSS: 772 KB, Size: 968 KB
  3. {/tmp/go-build121173397/b001/exe/main} PSS: 96 KB, RSS: 96 KB, Size: 96 KB
```

### Print memory usage of another process

This needs privileged permissions (sudo)

```
⇒  sudo go run main.go -pid 1
Summary:
  Total mappings: 212
  Total size: 164 MB
  Total RSS: 11 MB
  Total PSS: 3 MB
Top 10 mappings:
  1. {[heap]} PSS: 1 MB, RSS: 2 MB, Size: 2 MB
  2. {/usr/lib/systemd/systemd} PSS: 228 KB, RSS: 736 KB, Size: 740 KB
  3. {/usr/lib/systemd/libsystemd-shared-245.so} PSS: 216 KB, RSS: 1 MB, Size: 1 MB
  4. {/usr/lib/systemd/systemd} PSS: 83 KB, RSS: 252 KB, Size: 280 KB
  5. {/usr/lib/systemd/systemd} PSS: 76 KB, RSS: 252 KB, Size: 344 KB
  6. {/usr/lib/systemd/systemd} PSS: 66 KB, RSS: 200 KB, Size: 200 KB
  7. {/usr/lib/x86_64-linux-gnu/libcrypto.so.1.1} PSS: 58 KB, RSS: 176 KB, Size: 176 KB
  8. {[stack]} PSS: 49 KB, RSS: 52 KB, Size: 132 KB
  9. {/usr/lib/systemd/libsystemd-shared-245.so} PSS: 38 KB, RSS: 340 KB, Size: 564 KB
  10. {/usr/lib/x86_64-linux-gnu/libseccomp.so.2.4.3} PSS: 35 KB, RSS: 108 KB, Size: 108 KB
```

## Usage as a Library

You can import the library in your application as well.

```go
import (
    "github.com/grandecola/smaps/smaps"
)

sf, err := smaps.ReadSmaps(os.Getpid(), "")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Total RSS due to mmap: %v bytes\n", sf.RSS)
fmt.Printf("Total PSS due to mmap: %v bytes\n", sf.PSS)
```

Checkout an example [here](https://github.com/grandecola/smaps/blob/master/main.go).
