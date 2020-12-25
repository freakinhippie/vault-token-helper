/*
Copyright © 2020 Joshua Colson <joshua.colson@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mitchellh/go-homedir"
)

var (
	tokens      map[string]string
	tokenDir    string
	tokenFile   string
	vaultConfig string
)

func get(addr string) error {
	token, ok := tokens[addr]
	if ok {
		fmt.Print(token)
	}
	return nil
}

func save() error {
	contents, _ := json.MarshalIndent(tokens, "", "  ")
	if err := ioutil.WriteFile(tokenFile, contents, 0600); err != nil {
		return err
	}
	return nil
}

func store(addr, token string) error {
	tokens[addr] = token
	return save()
}

func erase(addr string) error {
	delete(tokens, addr)
	return save()
}

func enable(vaultConfig, contents string) error {
	if err := ioutil.WriteFile(vaultConfig, []byte(contents), 0644); err != nil {
		return err
	}
	return nil
}

func disable(vaultConfig string) error {
	if _, err := os.Stat(vaultConfig); os.IsNotExist(err) {
		return nil
	}
	return os.Remove(vaultConfig)
}

func main() {
	// get the path to this executable
	binPath, _ := exec.LookPath(os.Args[0])
	// convert the path to an absolute path
	helperAbsPath, _ := filepath.Abs(binPath)
	// extract the base name
	helper := filepath.Base(helperAbsPath)
	if len(os.Args) != 2 {
		// the vault token helper api accepts of of three arguments
		fmt.Fprintf(os.Stderr, "usage: %s get|store|erase|enable|disable\n", helper)
		os.Exit(1)
	}

	vaultAddr := os.Getenv("VAULT_ADDR")
	vaultConfig := os.Getenv("VAULT_CONFIG_PATH")

	if len(vaultAddr) == 0 {
		// the environment variable VAULT_ADDR must be set
		fmt.Fprintln(os.Stderr, "environment variable 'VAULT_ADDR' is required")
		os.Exit(2)
	}

	tokens = make(map[string]string)
	tokenDir, err := homedir.Expand("~/.config/vault.d")
	if err != nil {
		fmt.Fprintln(os.Stderr, "unable to expand home directory path '~/.config/vault.d'")
		os.Exit(3)
	}
	tokenFile = filepath.Join(tokenDir, "tokens")

	// VAULT_CONFIG_PATH settings
	if len(vaultConfig) == 0 {
		vaultConfig, err = homedir.Expand("~/.vault")
		if err != nil {
			fmt.Fprintln(os.Stderr, "unable to expand home directory path '~/.vault'")
			os.Exit(3)
		}
	}

	// create the directory, if needed
	if _, err := os.Stat(tokenDir); os.IsNotExist(err) {
		if err := os.Mkdir(tokenDir, 0700); err != nil {
			fmt.Fprintf(os.Stderr, "unable to create tokens directory [%s]: %s", tokenDir, err.Error())
			os.Exit(4)
		}
	}

	// read any tokens
	if content, err := ioutil.ReadFile(tokenFile); err == nil {
		if err = json.Unmarshal(content, &tokens); err != nil {
			fmt.Fprintf(os.Stderr, "error reading tokens file: %s", err.Error())
			os.Exit(5)
		}
	}

	// process command argument
	switch os.Args[1] {
	case "get":
		if err := get(vaultAddr); err != nil {
			fmt.Fprintf(os.Stderr, "error getting token for %s: %s", vaultAddr, err.Error())
			os.Exit(6)
		}
	case "store":
		reader := bufio.NewScanner(os.Stdin)
		reader.Scan()
		token := strings.TrimSpace(reader.Text())
		if len(token) > 0 {
			if err := store(vaultAddr, token); err != nil {
				fmt.Fprintf(os.Stderr, "error storing token for %s: %s", vaultAddr, err.Error())
				os.Exit(7)
			}
		} else {
			// erase when the token is an empty string,
			// as outlined in the [Storing] section here:
			//   https://www.hashicorp.com/blog/building-a-vault-token-helper
			if err := erase(vaultAddr); err != nil {
				fmt.Fprintf(os.Stderr, "error erasing token for %s: %s", vaultAddr, err.Error())
				os.Exit(8)
			}
		}
	case "erase":
		if err := erase(vaultAddr); err != nil {
			fmt.Fprintf(os.Stderr, "error erasing token for %s: %s", vaultAddr, err.Error())
			os.Exit(9)
		}
	case "enable":
		if err := enable(vaultConfig, fmt.Sprintf("token_helper = \"%s\"\n", helperAbsPath)); err != nil {
			fmt.Fprintf(os.Stderr, "error enabling vault token helper: %s", err.Error())
			os.Exit(10)
		}
	case "disable":
		if err := disable(vaultConfig); err != nil {
			fmt.Fprintf(os.Stderr, "error disabling vault token helper: %s", err.Error())
			os.Exit(11)
		}
	}
}
