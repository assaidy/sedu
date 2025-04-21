package main

import (
	"bufio"
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

// isSymlink checks if a given path is a symbolic link.
func isSymlink(path string) bool {
	info, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeSymlink != 0
}

// listFiles recursively lists all files in a directory, ignoring symlinks.
func listFiles(dirPath string) []string {
	var files []string
	var stack []string
	stack = append(stack, dirPath)

	for len(stack) > 0 {
		currentDir := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		entries, err := os.ReadDir(currentDir)
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}

		for _, entry := range entries {
			path := filepath.Join(currentDir, entry.Name())
			if !isSymlink(path) {
				if entry.IsDir() {
					stack = append(stack, path)
				} else {
					files = append(files, path)
				}
			}
		}
	}
	return files
}

// computeHash computes the SHA-1 hash of a file using buffered I/O.
func computeHash(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	bufferedReader := bufio.NewReader(file)
	hasher := sha1.New()

	if _, err := io.Copy(hasher, bufferedReader); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

// FileHash represents a file and its computed hash.
type FileHash struct {
	hash, path string
}

// worker processes files from the jobs channel and sends the results to the results channel.
func worker(jobs <-chan string, results chan<- *FileHash, wg *sync.WaitGroup) {
	defer wg.Done()
	for file := range jobs {
		hash, err := computeHash(file)
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}
		results <- &FileHash{hash, file}
	}
}

func main() {
	args := os.Args

	if len(args) != 2 {
		fmt.Println("Usage: sedu <dir-path>")
		fmt.Println("ERROR: missing argument: dir-path")
		os.Exit(1)
	}

	dirPath := args[1]

	fmt.Println("Info: Collecting all files...")
	allFiles := listFiles(dirPath)

	fmt.Println("Info: Generating file hashes...")
	hashToPaths := make(map[string][]string)

	numWorkers := runtime.NumCPU()
	jobs := make(chan string, len(allFiles))
	results := make(chan *FileHash, len(allFiles))
	var wg sync.WaitGroup

	// Start background workers
	for range numWorkers {
		wg.Add(1)
		go worker(jobs, results, &wg)
	}

	// Send jobs to workers
	go func() {
		for _, file := range allFiles {
			jobs <- file
		}
		close(jobs)
	}()

	// Wait for all workers to finish
	go func() {
		wg.Wait()
		close(results)
	}()

	for result := range results {
		hashToPaths[result.hash] = append(hashToPaths[result.hash], result.path)
	}

	fmt.Println("Info: Printing all duplicate files...")
	for _, paths := range hashToPaths {
		if len(paths) > 1 {
			fmt.Println("{")
			for _, p := range paths {
				fmt.Println("   " + p)
			}
			fmt.Println("}")
		}
	}
}
