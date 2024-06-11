package main

// TODO: do it in parallel

import (
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func isSymLink(path string) bool {
	_, err := os.Readlink(path)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Error:", err)
		} else {
			return false
		}
	}
	return true
}

func getHash(path string) string {
	h := sha1.New()

	file, err := os.Open(path)
	if err != nil {
		fmt.Println("Error:", err)
		return ""
	}
	defer file.Close()

	if _, err := io.Copy(h, file); err != nil {
		fmt.Println("Error:", err)
		return ""
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}

func main() {
	args := os.Args

	if len(args) != 2 {
		fmt.Println("Usage: sedu <file-path>")
		fmt.Println("ERROR: messing argument: file-path")
		os.Exit(1)
	}

	dirPath := args[1]

	hashes := make(map[string][]string)

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Ignore errors that indicate the file was not found
			if os.IsNotExist(err) || os.IsPermission(err) {
				return nil
			}
			return err
		}
		if !info.IsDir() && !isSymLink(path) {
			absPath, err := filepath.Abs(path)
			if err != nil {
				fmt.Println("Error:", err)
			}
			// NOTE: storing hashes + file pathes
			hash := getHash(absPath)
			if _, ex := hashes[hash]; !ex {
				hashes[hash] = []string{}
			}
			hashes[hash] = append(hashes[hash], absPath)
		}
		return nil
	})

	if err != nil {
		fmt.Println("Error:", err)
		// os.Exit(1)
	}

	for _, v := range hashes {
		if len(v) > 1 {
			for _, p := range v {
				fmt.Println(p)
			}
			fmt.Println()
		}
	}
}
