package main

import (
	"encoding/json"
	"flag"
	"fmt"
	ipfs "github.com/39alpha/ipfs-tools/ipfs-shell"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var (
	nopin   = false
	nofetch = false
	dryrun  = false
	verbose = false
	ipfsurl = "127.0.0.1:5001"
)

func init() {
	flag.StringVar(&ipfsurl, "ipfsurl", ipfsurl, "URL to running IPFS node")
	flag.BoolVar(&nopin, "nopin", nopin, "Do not pin the fetched asset")
	flag.BoolVar(&nofetch, "nofetch", nofetch, "Do not download files; only pin them to the IPFS node")
	flag.BoolVar(&verbose, "verbose", verbose, "Generate verbose output")
	flag.BoolVar(&dryrun, "dryrun", dryrun, "Pretend to fetch and pin, but don't really do it (useful with the -verbose flag)")
}

type Payload map[string]string

func ReadPayload(filename string) (Payload, error) {
	blob, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var payload Payload
	if err := json.Unmarshal(blob, &payload); err != nil {
		return nil, err
	}

	return payload, nil
}

type PathValidityError int

const (
	NotRelative PathValidityError = 123 + iota
	NotBelow
)

func (e PathValidityError) Error() string {
	switch e {
	case NotRelative:
		return "path is not and cannot be made relative"
	case NotBelow:
		return "path is above the current working directory"
	}
	return "unrecognized error"
}

func Normalize(p string) (string, error) {
	if path.IsAbs(p) {
		return "", NotRelative
	}

	rel, err := filepath.Rel(".", p)
	if err != nil {
		return "", NotRelative
	}

	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", NotBelow
	}

	return rel, nil
}

func Usage() {
	fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [OPTION]... [FILE]...\n\n", os.Args[0])
	fmt.Fprintf(flag.CommandLine.Output(), "Fetch assets specified by [FILE]...\n\n")
	fmt.Fprintf(flag.CommandLine.Output(), "Options:\n\n")
	flag.PrintDefaults()
}

func NoFetch(fetcher *ipfs.IpfsShell, hash, p string, verbose, dryrun bool) error {
	if verbose {
		fmt.Printf("INFO: Pinning asset %q (%q)\n", hash, p)
	}
	if !dryrun {
		if err := fetcher.Pin(hash); err != nil {
			return fmt.Errorf("failed to pin asset %q — %v\n", hash, err)
		}
	}
	return nil
}

func NoPin(fetcher *ipfs.IpfsShell, hash, p string, verbose, dryrun bool) error {
	if verbose {
		fmt.Printf("INFO: Fetching asset %q to %q\n", hash, p)
	}
	if !dryrun {
		if err := fetcher.Fetch(hash, p); err != nil {
			return fmt.Errorf("failed to fetch asset %q to path %q — %v\n", hash, p, err)
		}
	}
	return nil
}

func FetchAndPin(fetcher *ipfs.IpfsShell, hash, p string, verbose, dryrun bool) error {
	if verbose {
		fmt.Printf("INFO: Fetching and pinning asset %q to %q\n", hash, p)
	}
	if !dryrun {
		if err := fetcher.FetchAndPin(hash, p); err != nil {
			return fmt.Errorf("failed to fetch asset %q to path %q — %v\n", hash, p, err)
		}
	}
	return nil
}

func main() {
	flag.Parse()

	if flag.NArg() == 0 {
		Usage()
		os.Exit(1)
	}

	if nofetch && nopin && !verbose {
		fmt.Fprintf(os.Stderr, "WARNING: -nopin and -nofetch were both provided; no disk or IPFS node modifications will be performed\n")
	}

	fetcher, err := ipfs.NewIpfsShell(ipfsurl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: cannot establish connection to IPFS shell — %v\n", err)
		os.Exit(3)
	}

	for _, file := range flag.Args() {
		payload, err := ReadPayload(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: cannot read data payload — %v\n", err)
			os.Exit(2)
		}

		exitcode := 0
		for hash, dest := range payload {
			if p, err := Normalize(dest); err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: cannot fetch asset %q to path %q — %v\n", hash, dest, err)
				exitcode = 4
			} else {
				err = nil
				if nofetch && !nopin {
					err = NoFetch(fetcher, hash, p, verbose, dryrun)
				} else if !nofetch && nopin {
					err = NoPin(fetcher, hash, p, verbose, dryrun)
				} else if !nofetch {
					err = FetchAndPin(fetcher, hash, p, verbose, dryrun)
				} else if verbose {
					fmt.Printf("INFO: Ignoring asset %q (%q)\n", hash, p)
				}
				if err != nil {
					fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
					exitcode = 5
				}
			}
		}

		os.Exit(exitcode)
	}
}
