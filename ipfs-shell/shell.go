package shell

import (
	"fmt"
	ipfs "github.com/ipfs/go-ipfs-api"
	"io/fs"
	"os"
	"path/filepath"
)

type IpfsShell struct {
	s *ipfs.Shell
}

func NewIpfsShell(url string) (*IpfsShell, error) {
	shell := ipfs.NewShell(url)
	if shell != nil && shell.IsUp() {
		return &IpfsShell{shell}, nil
	}
	return nil, fmt.Errorf("shell is not up")
}

func (shell *IpfsShell) IsUp() bool {
	if shell.s == nil {
		return false
	}

	return shell.s.IsUp()
}

func (shell *IpfsShell) List(path string) ([]*ipfs.LsLink, error) {
	if !shell.IsUp() {
		return nil, fmt.Errorf("shell is not up")
	}

	return shell.s.List(path)
}

func (shell *IpfsShell) IsIpfsDir(path string) (bool, error) {
	if !shell.IsUp() {
		return false, fmt.Errorf("shell is not up")
	}

	entries, err := shell.List(path)
	if err != nil {
		return false, err
	}

	for _, link := range entries {
		if link.Name != "" {
			return true, nil
		}
	}
	return false, nil
}

func (shell *IpfsShell) Get(hash, outdir string) error {
	if !shell.IsUp() {
		return fmt.Errorf("shell is not up")
	}

	return shell.s.Get(hash, outdir)
}

func (shell *IpfsShell) Fetch(hash, path string) error {
	if !shell.IsUp() {
		return fmt.Errorf("shell is not up")
	}

	isdir, err := shell.IsIpfsDir(hash)
	if err != nil {
		return err
	} else if isdir {
		if err = os.MkdirAll(path, fs.ModeDir|fs.ModePerm); err != nil {
			return err
		}

		return shell.Get(hash, path)
	} else {
		outdir := filepath.Dir(path)

		if err = os.MkdirAll(outdir, fs.ModeDir|fs.ModePerm); err != nil {
			return err
		}

		if err = shell.Get(hash, outdir); err != nil {
			return err
		}

		outpath := filepath.Join(outdir, hash)
		if _, err := os.Stat(outpath); err == nil {
			if err = os.Rename(outpath, path); err != nil {
				if err = os.Remove(outpath); err != nil {
					return fmt.Errorf("could not cleanup after failed fetch; %v", err)
				}
			}
		} else {
			return fmt.Errorf("no blocks downloaded to %q", outpath)
		}

		return nil
	}
}

func (shell *IpfsShell) Put(path string) (string, error) {
    if !shell.IsUp() {
        return "", fmt.Errorf("shell is not up")
    }

    if stat, err := os.Stat(path); err != nil {
        return "", err
    } else if stat.IsDir() {
        return shell.s.AddDir(path)
    } else if stat.Mode().IsRegular() {
        file, err := os.Open(path)
        if err != nil {
            return "", err
        }
        defer file.Close()

        return shell.s.Add(file)
    } else {
        return "", fmt.Errorf("path must be a directory or regular file")
    }
}
