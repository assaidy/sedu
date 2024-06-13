package main

import (
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

// FIX: ignore symLinks
func isSymLink(path string) bool {
	_, err := os.Readlink(path)
	return err == nil
}

func listDirFiles(dirPath string) []string {
	var res []string
	var stack []string
	stack = append(stack, dirPath)

	for len(stack) > 0 {
		curDir := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		enteries, err := os.ReadDir(curDir)
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}

		for _, entery := range enteries {
			path := filepath.Join(curDir, entery.Name())
			if entery.IsDir() {
				stack = append(stack, path)
			} else {
				// fmt.Println(path)
				res = append(res, path)
			}
		}
	}
	return res
}

func getHash(path string) (string, error) {
	h := sha1.New()

	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if _, err := io.Copy(h, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

type FileHash struct {
	hash, path string
}

func worker(files <-chan string, results chan<- *FileHash, wg *sync.WaitGroup) {
	defer wg.Done()
	for file := range files {
		hash, err := getHash(file)
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
		fmt.Println("ERROR: messing argument: dir-path")
		os.Exit(1)
	}

	dirPath := args[1]

	fmt.Println("[INFO] collecting all files...")
	allFiles := listDirFiles(dirPath)

	fmt.Println("[INFO] generating file hashes...")
	hasheToPaths := make(map[string][]string)
	var mutex sync.Mutex

	numWorkers := runtime.NumCPU()
	jobs := make(chan string, len(allFiles))
	results := make(chan *FileHash, len(allFiles))
	var wg sync.WaitGroup

	for w := 1; w <= numWorkers; w++ {
		// start workers
		wg.Add(1)
		go worker(jobs, results, &wg)
	}

	// send jobs to worker
	go func() {
		for _, file := range allFiles {
			jobs <- file
		}
		close(jobs)
	}()

	// wait for all workers
	go func() {
		wg.Wait()
		close(results)
	}()

	for result := range results {
		hash, path := result.hash, result.path
		mutex.Lock()
		if _, exists := hasheToPaths[hash]; !exists {
			hasheToPaths[hash] = []string{}
		}
		hasheToPaths[hash] = append(hasheToPaths[hash], path)
		mutex.Unlock()
	}

	fmt.Println("[INFO] printing all duplicate files...")
	for _, v := range hasheToPaths {
		if len(v) > 1 {
			fmt.Println("{")
			for _, p := range v {
				fmt.Println("   " + p)
			}
			fmt.Println("}")
		}
	}
}
