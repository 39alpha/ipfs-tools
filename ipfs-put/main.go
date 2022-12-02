package main

import (
	"encoding/json"
	"flag"
	"fmt"
	ipfs "github.com/39alpha/ipfs-tools/ipfs-shell"
	"io"
	"io/ioutil"
	"os"
)

var (
	payloadfile = ""
	ipfsurl     = "127.0.0.1:5001"
)

func init() {
	flag.StringVar(&payloadfile, "o", payloadfile, "File to which to write the payload")
	flag.StringVar(&ipfsurl, "ipfsurl", ipfsurl, "URL to running IPFS node")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [OPTIONS] [FILE|DIR...]\n\n", os.Args[0])
		fmt.Fprintf(flag.CommandLine.Output(), "Add assets from one or more files or directories to an IPFS node and\nwrite a JSON object mapping CIDs to the path of the asset.\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Options:\n")
		flag.PrintDefaults()
	}
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

func (p Payload) WriteTo(w io.Writer) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(p)
}

func (p Payload) Save(filename string) error {
	if filename == "" {
		return p.WriteTo(os.Stdout)
	} else {
		file, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer file.Close()

		return p.WriteTo(file)
	}
}

func main() {
	flag.Parse()

	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}

	payload, err := ReadPayload(payloadfile)
	if err != nil {
		payload = make(Payload)
	}

	putter, err := ipfs.NewIpfsShell(ipfsurl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: cannot establish connection to IPFS shell — %v\n", err)
		os.Exit(2)
	}

	for _, entry := range flag.Args() {
		hash, err := putter.Put(entry)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: could not add %q to IPFS — %v\n", entry, err)
		} else {
			payload[hash] = entry
		}
	}

	if err = payload.Save(payloadfile); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: could not write payload to %q — %v\n", payload, err)
		os.Exit(3)
	}
}
