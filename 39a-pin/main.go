package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
)

var (
	home           = ""
	user           = ""
	domain         = "39alpharesearch.org"
	port           = 22
	knownHostsFile = ""
	privateKey     = ""
)

func init() {
	user = os.Getenv("USER")
	if h, ok := os.LookupEnv("HOME"); !ok {
		home = filepath.Join("home", user)
	} else {
		home = h
	}
	privateKey = filepath.Join(home, ".ssh", "id_rsa")

	flag.StringVar(&user, "user", os.Getenv("USER"), "username on the server")
	flag.StringVar(&domain, "domain", domain, "the server domain name or ip address")
	flag.IntVar(&port, "port", port, "the port to use")
	flag.StringVar(&privateKey, "i", privateKey, "the private ssh key to use")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [OPTIONS] [CID...]\n\n", os.Args[0])
		fmt.Fprintf(flag.CommandLine.Output(), "Pin assets to a remote IPFS node. If no CIDs are provided at the\ncommand line, a JSON-formatted data mapping is read from standard\ninput and all CIDs in the mapping are pinned.\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Options:\n")
		flag.PrintDefaults()
	}
}

type Payload map[string]string

func ParsePayload(blob []byte) (Payload, error) {
	var payload Payload
	if err := json.Unmarshal(blob, &payload); err != nil {
		return nil, err
	}

	return payload, nil
}

func ReadPayload(r io.Reader) (Payload, error) {
	blob, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return ParsePayload(blob)
}

func credentials() ([]byte, error) {
	fmt.Printf("SSH Key Passphrase: ")
	file, err := os.OpenFile("/dev/tty", os.O_RDONLY, 0755)
	if err != nil {
		return nil, err
	}
	passphrase, err := terminal.ReadPassword(int(file.Fd()))
	fmt.Println()
	return passphrase, err
}

func ParsePrivateKey() (ssh.Signer, error) {
	pemBytes, err := ioutil.ReadFile(privateKey)
	if err != nil {
		return nil, err
	}

	key, err := ssh.ParsePrivateKey(pemBytes)
	if err != nil {
		switch err.(type) {
		case *ssh.PassphraseMissingError:
			password, err := credentials()
			if err != nil {
				return nil, err
			}
			return ssh.ParsePrivateKeyWithPassphrase(pemBytes, password)
		default:
			return nil, err
		}
	}
	return ssh.NewSignerFromKey(key)
}

func unlockKeyring() (agent.ExtendedAgent, error) {
	if socket, ok := os.LookupEnv("SSH_AUTH_SOCK"); ok {
		conn, err := net.Dial("unix", socket)
		if err != nil {
			return nil, err
		}
		return agent.NewClient(conn), nil
	} else {
		return nil, fmt.Errorf("cannot find ssh agent socket")
	}
}

func findHostKeys(filename, host string) ([]ssh.PublicKey, error) {
	rest, err := ioutil.ReadFile(filename)
	if err != nil {
		os.Exit(1)
	}

	hostKeys := []ssh.PublicKey{}
	for {
		_, hosts, key, _, r, err := ssh.ParseKnownHosts(rest)
		if err != nil {
			break
		}
		rest = r
		for _, knownHost := range hosts {
			if knownHost == host {
				hostKeys = append(hostKeys, key)
			}
		}
	}

	return hostKeys, nil
}

func startClient() (*ssh.Client, error) {
	hostKeys, err := findHostKeys(filepath.Join(filepath.Dir(privateKey), "known_hosts"), domain)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: unable to read known hosts — %v\n", err)
		os.Exit(1)
	}

	config := &ssh.ClientConfig{
		User: user,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) (err error) {
			for _, hostKey := range hostKeys {
				err = ssh.FixedHostKey(hostKey)(hostname, remote, key)
				if err == nil {
					break
				}
			}
			return
		},
	}

	keyring, err := unlockKeyring()
	if err != nil {
		signer, err := ParsePrivateKey()
		if err != nil {
			return nil, err
		}

		config.Auth = []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		}

		return ssh.Dial("tcp", fmt.Sprintf("%s:%d", domain, port), config)
	}

	config.Auth = []ssh.AuthMethod{
		ssh.PublicKeysCallback(keyring.Signers),
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", domain, port), config)
	if err != nil {
		signer, err := ParsePrivateKey()
		if err != nil {
			return nil, err
		}

		config.Auth = []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		}

		return ssh.Dial("tcp", fmt.Sprintf("%s:%d", domain, port), config)
	}

	return client, nil
}

func main() {
	flag.Parse()

	client, err := startClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: cannot connect to ssh server — %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	hashes := flag.Args()
	if len(hashes) == 0 {
		payload, err := ReadPayload(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: cannot read data payload from standard input — %v", err)
			os.Exit(2)
		}
		hashes = make([]string, 0, len(payload))
		for hash := range payload {
			hashes = append(hashes, hash)
		}
	}

	for _, hash := range hashes {
		session, err := client.NewSession()
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: unable to create session — %v", err)
			os.Exit(3)
		}
		defer session.Close()

		var stderr bytes.Buffer
		session.Stderr = &stderr

		cmd := fmt.Sprintf("ipfs pin add %s", hash)
		if err = session.Run(cmd); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: failed to pin %s\n  %s", hash, stderr.String())
		} else {
			fmt.Printf("INFO: Pinned %s\n", hash)
		}
	}
}
