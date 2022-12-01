package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	ipfs "github.com/39alpha/ipfs-tools/ipfs-shell"
	"net/http"
	"os"
	"regexp"
	"time"
)

var (
	ipfsurl = "127.0.0.1:5001"
	api     = "api/v0"
	verbose = false
)

const ThirtyNineAlpha = "https://39alpharesearch.org/"

func init() {
	flag.StringVar(&api, "api", api, "39A API version")
	flag.StringVar(&ipfsurl, "ipfsurl", ipfsurl, "URL to running IPFS node")
	flag.BoolVar(&verbose, "v", verbose, "Print verbose status messages")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [OPTIONS]\n\n", os.Args[0])
		fmt.Fprintf(flag.CommandLine.Output(), "Establish a peer-to-peer connection to the 39 Alpha gateway node.\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Options:\n")
		flag.PrintDefaults()
	}
}

type Addrs struct {
	Addresses []string `json:"addresses"`
}

func (a Addrs) ChooseAddress() (string, error) {
	if len(a.Addresses) == 0 {
		return "", fmt.Errorf("no addresses return from gateway")
	}

	udp := regexp.MustCompile(`^.*/udp/`)

	for _, addr := range a.Addresses {
		if udp.MatchString(addr) {
			return addr, nil
		}
	}

	return a.Addresses[0], nil
}

func main() {
	flag.Parse()

	shell, err := ipfs.NewIpfsShell(ipfsurl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: cannot establish connection to IPFS shell — %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Get(ThirtyNineAlpha + api + "/ipfs/addr")
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: cannot fetch IPFS address from gateway — %v\n", err)
		os.Exit(2)
	}

	if resp.StatusCode != 200 {
		fmt.Fprintf(os.Stderr, "ERROR: cannot fetch IPFS address from gateway — response code: %v\n", resp.StatusCode)
		os.Exit(3)
	}

	var addrs Addrs

	dec := json.NewDecoder(resp.Body)
	if err = dec.Decode(&addrs); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: could not parse response from gateway — %v\n", err)
		os.Exit(4)
	}

	addr, err := addrs.ChooseAddress()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(5)
	}
	if verbose {
		fmt.Printf("Connecting to %s ... ", addr)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err = shell.SwarmConnect(ctx, addr); err != nil {
		if verbose {
			fmt.Println()
		}
		fmt.Fprintf(os.Stderr, "\nERROR: connection attempt timed out")
		os.Exit(6)
	}

	if verbose {
		fmt.Println("done")
	}
}
