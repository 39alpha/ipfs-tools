package main

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
	"io/ioutil"
	"os"
	"syscall"
)

func credentials() ([]byte, error) {
	fmt.Printf("SSH Key Passphrase: ")
	passphrase, err := terminal.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	return passphrase, err
}

func ParsePrivateKey(pemBytes []byte) (ssh.Signer, error) {
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

func main() {
	var hostKey ssh.PublicKey

	key, err := ioutil.ReadFile("/home/dgm/.ssh/id_rsa")
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to read private key: %v", err)
		os.Exit(1)
	}

	signer, err := ParsePrivateKey(key)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to parse private key: %v\n", err)
		os.Exit(2)
	}

	config := &ssh.ClientConfig{
		User: "doug",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.FixedHostKey(hostKey),
	}

	client, err := ssh.Dial("tcp", "39alpharesearch.org:22", config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to connect: %v\n", err)
		os.Exit(3)
	}
	defer client.Close()
}
