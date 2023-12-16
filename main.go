package main

import (
	"fmt"
	"os"
	// "os/signal"
	"strconv"
	"sync/atomic"
	"sync"
	"bufio"
	"time"
	// "syscall"
	// tea "github.com/charmbracelet/bubbletea"
)

var (
	smallestPrime int64 = -1
	largestPrime  int64 = -1
	primesMutex   sync.Mutex
	workersSpawned bool
	primes = make(chan int)
)

func main() {
	StartNow := time.Now
	dirPath := "./rand"

	files := getFiles(dirPath)

	fileChannel := make(chan string, 1000) // channel to store the files in buffer 1000
	for _,file := range files {
		fileChannel <- file
	}
	fmt.Print("Closing file channel")
	// close(fileChannel)

	// Define the number of goroutines to use (e.g., 4 for quadrupling throughput (limited severely by IOPs))
	numGoroutines := 1

	var fileGroup sync.WaitGroup

	workers := func(id int, fileChannel chan string, wg *sync.WaitGroup) {
		fmt.Println("Deferring fileGroup")
		defer fileGroup.Done()

		filesDone := false
		for !filesDone {
			select{
			case filePath, filesOk := <- fileChannel:
				if !filesOk{
					filesDone = true
					return
				} 

				numbers, err := readFiles(filePath)
				if err != nil {
					fmt.Println("Error reading files:", err)
					os.Exit(1)
				}

				
				for _, number := range numbers {
					if isPrime(number) == true {
						primes <- number		// stash the primes in a channel for reading later
					}
				}
				fmt.Println("Worker:", id, "Processing", filePath)
			}
		}
	}

	// go func() {

	if workersSpawned == false {
		for i := 0; i < numGoroutines; i++ {
			fmt.Println("spawn a file worker")
			fileGroup.Add(1)
			go workers(i, fileChannel, &fileGroup)
			fmt.Println("File Worker spawned")
			workersSpawned = true
		}
	}

	// Asynchronously read the primes out of the channel while the other function is writing them in
	go func () {
		for prime := range primes {
			prime64 := int64(prime)
			updatePrimes(prime64)
			fmt.Print("\033[H\033[2J")
			fmt.Println("largest:", largestPrime, "smallest:", smallestPrime)
		}
	}()

	fileGroup.Wait()
	// close(fileChannel)


	fmt.Println("progress should be done now")
	fmt.Println(smallestPrime, largestPrime)
	fmt.Println("that took:", time.Since(StartNow()))

}

// 📑 what files are in a given directory
func getFiles(dirPath string) (the_files []string) {
	files, _ := os.ReadDir(dirPath)
	for _, file := range files {
		the_files = append(the_files, dirPath+"/"+file.Name()) // make a list of the files in the chosen directory
	}
	return the_files
}

// Producer
func readFiles(filePath string) ([]int, error) {
	var numbers []int

	
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fmt.Println("Reading:", filePath)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		number, err := strconv.Atoi(line)
		if err == nil {
			numbers = append(numbers, number)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return numbers, nil
}

// Consumer
func isPrime(n int) bool {
	if n <= 1 {
		return false
	}
	if n <= 3 {
		return true
	}
	if n%2 == 0 || n%3 == 0 {
		return false
	}
	i := 5
	for i*i <= n {
		if n%i == 0 || n%(i+2) == 0 {
			return false
		}
		i += 6
	}
	return true
}

func updatePrimes(prime int64) {
	primesMutex.Lock()
	defer primesMutex.Unlock()
	// Update smallestPrim
	for {
	
		currentSmallest := atomic.LoadInt64(&smallestPrime)
		if currentSmallest == -1 || prime < currentSmallest {
			fmt.Println("debug")
			
			if atomic.CompareAndSwapInt64(&smallestPrime, currentSmallest, prime) {
				// fmt.Println(smallestPrime)
				break
			}
		} else {
			break
		}
	}

	// Update largestPrime
	for {
		currentLargest := atomic.LoadInt64(&largestPrime)
		if currentLargest == -1 || prime > currentLargest {
			if atomic.CompareAndSwapInt64(&largestPrime, currentLargest, prime) {
				// fmt.Println(largestPrime)
				break
			}
		} else {
			break
		}
	}
}