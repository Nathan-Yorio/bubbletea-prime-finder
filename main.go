package main

import (
	"fmt"
	"os"
	"sync/atomic"
	// "os/signal"
	"strconv"
	"sync"
	"bufio"
	// "time"
	// "runtime"
	// "syscall"
	// tea "github.com/charmbracelet/bubbletea"
)

var (
	// largestPrime = -1
	// largestPrimeMutex sync.Mutex
	// smallestPrime = -1
	// smallestPrimeMutex sync.Mutex
	smallestPrime int64 = -1
	largestPrime  int64 = -1
	primesMutex   sync.Mutex
)


func main() {
	dirPath := "./rand"

	files := getFiles(dirPath)

	// fmt.Print(files)

	var wg sync.WaitGroup
	// var disp sync.WaitGroup
	primes := make(chan int)
	numbers := make(chan int)
	fileChannel := make(chan string, 1000) // channel to store the files in buffer 1000

	for _,file := range files {
		fileChannel <- file
	}
	fmt.Print("Closing file channel")
	close(fileChannel)

	// Define the number of goroutines to use (e.g., 4 for quadrupling)
	numGoroutines := 5

	// Workers
	// Buffered channel, size of numGoRoutines
	workerPool := make(chan struct{}, numGoroutines)
	progressChannel := make(chan string, numGoroutines)
	progressDone := make(chan struct{})

// Worker function
	worker := func(workerID int, numbers chan int, primes chan int, wg *sync.WaitGroup) {
		// fmt.Println("Enter worker function")
			// fmt.Println("Release worker pool slot")
			// fmt.Println("worker killed")
		defer func() {
			wg.Done()
			// <-workerPool
		}()
		

		
		for {
			select {
			// case prime, primesOk := <-primes:
			// 	fmt.Println("primes", primesOk)
			// 	if !primesOk {
			// 		// No more primes to process, exit the loop
			// 		// return
			// 		fmt.Println("something")
			// 	} else {
			// 							// Process the prime
				
			// 	// prime64 := int64(prime)
			// 	// updatePrimes(prime64)
			// 	}
			case filePath, filesOk := <-fileChannel:
				if !filesOk{
					close(fileChannel)
				} else {
				// fmt.Println("files:", filesOk)
				 // Continue processing primes once files are done
					fmt.Println("parsing files")
					readFiles(filePath, numbers)
					fmt.Println("files parsed")

					// Send progress information to the channel
					fmt.Println(workerID, filePath)
					progressChannel <- fmt.Sprintf("Worker %d processed file: %s", workerID, filePath)
				}
			case number, numsOk := <- numbers:
				fmt.Println("nums:", numsOk)
				if numsOk {
					fmt.Println("reading:", number)
					readNumbers(number, primes)
				}
			}
		}
	}

	go func(numbers chan int, primes chan int) {
		fmt.Println("prespawn")
		for i := 0; i < numGoroutines ; i++ {
			fmt.Println("spawn a boye")
			workerPool <- struct{}{}  // Acquire a slot in the worker pool
			wg.Add(1)                 // Add 1 worker to the pool
			go worker(i, numbers, primes, &wg)         // Set worker to work
			fmt.Println("boye spawend")
			// <-workerPool
		}
	}(numbers, primes)

	// Create a goroutine to display progress information
	go func() {
		fmt.Println("any kind of progress display")
		// defer close(progressDone)
		for progress := range progressChannel {
			fmt.Print("Progress:")
			fmt.Println(progress)
		}
	}()

	// go func() {
	// 	for prime := range primes {
	// 		fmt.Println("Received Prime:", prime)
	// 		// prime := int64(prime)
	// 		// updatePrimes(prime)
	// 	// 	// Process the prime number as needed
	// 	}
	// 	for number := range numbers {
	// 		fmt.Println("Received number:", number)
	// 	}
	// }()
	
// unbuffered channels will block if not actively consumed
	go func() {
		for {
			select {
			case num := <-numbers:
				// Process the received number
				// time.Sleep(500 * time.Millisecond)
				// fmt.Println("Received num:", num)
				UNUSED(num)
			// Add more cases or logic as needed
			}
		}
	}()

	go func() {
		for {
			select {
			case prime := <-primes:
				// Process the received number
				// time.Sleep(500 * time.Millisecond)
				fmt.Println("aaaaaaaaaaaaaaaaaaaaaaReceived prime:", prime)
			// Add more cases or logic as needed
				prime64 := int64(prime)
				updatePrimes(prime64)
			}
		}
	}()

	// defer wg.Done()
	fmt.Println("enter async wait")
	go func() {
		wg.Wait()
		// close(numbers)
	}()


	<-progressDone
	fmt.Println("progress should be done now")

}

func UNUSED(x ...interface{}) {}

// 📑 what files are in a given directory
func getFiles(dirPath string) (the_files []string) {
	files, _ := os.ReadDir(dirPath)
	for _, file := range files {
		the_files = append(the_files, dirPath+"/"+file.Name()) // make a list of the files in the chosen directory
	}
	return the_files
}

// Producer
func readFiles(file string, numbers chan<- int) {
	// fmt.Println("opening file:", file)
	data, err := os.Open(file)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer data.Close()

	// fmt.Println("scanning file:", file)
	scanner := bufio.NewScanner(data)
	for scanner.Scan() {
		line := scanner.Text()
		number, err := strconv.Atoi(line)
		if err == nil {
			// if isPrime(number) {
			// 	// fmt.Println("Stashing Prime:", number)
			// 	primes <- number
			// }
			// fmt.Println("stashing num:", number)
			numbers <- number //stash the numbers in a channel
			// fmt.Println("num", number, "stashed")
		}
	}

	// fmt.Println("post stash")
	if err := scanner.Err(); err != nil {
		fmt.Println(err)
		return
	}
}

func readNumbers(number int, primes chan<- int) {
	if isPrime(number) {
		// fmt.Println("Stashing Prime:", number)
		primes <- number // stashing the primes in a channel
	}
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
				fmt.Println(smallestPrime)
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
				fmt.Println(largestPrime)
				break
			}
		} else {
			break
		}
	}
}