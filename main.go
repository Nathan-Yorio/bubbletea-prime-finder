// type Model struct {
// 	FilesProcessed  int
// 	MinPrime        int
// 	MaxPrime        int
// 	NumThreads      int
// 	Running         bool
// 	ConsumerThreads int
// }

// type Msg string

// const (
// 	StartConsumer Msg = "StartConsumer"
// 	StopConsumer  Msg = "StopConsumer"
// 	Quit          Msg = "Quit"
// )

// func (m Model) Init() tea.Model {
// 	return m
// }

// func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
// 	switch msg.(type) {
// 	case tea.KeyMsg:
// 		keyMsg := msg.(tea.KeyMsg)
// 		switch keyMsg.String() {
// 		case "q":
// 			return m, tea.Quit
// 		case "up":
// 			return Model{NumThreads: m.NumThreads + 1}, tea.Batch(StartConsumer)
// 		case "down":
// 			if m.NumThreads > 0 {
// 				return Model{NumThreads: m.NumThreads - 1}, tea.Batch(StopConsumer)
// 			}
// 		}
// 	case Msg:
// 		switch msg {
// 		case StartConsumer:
// 			// Implement starting a new consumer thread
// 			go startConsumer()
// 			return Model{ConsumerThreads: m.ConsumerThreads + 1}, nil
// 		case StopConsumer:
// 			// Implement stopping a consumer thread
// 			// You need to implement a way to gracefully stop a consumer
// 			return Model{ConsumerThreads: m.ConsumerThreads - 1}, nil
// 		case Quit:
// 			os.Exit(0)
// 		}
// 	}

// 	return m, nil
// }

// func (m Model) View() string {
// 	return fmt.Sprintf(`
//     Number of Files Processed: %d
//     Min Prime: %d
//     Max Prime: %d
//     Number of Threads: %d (Press + to increase, - to decrease)
//     Number of Active Threads: %d
//     Running: %t (Press 'q' to quit)
//     `, m.FilesProcessed, m.MinPrime, m.MaxPrime, m.NumThreads, m.ConsumerThreads, m.Running)
// }

// func main() {
// 	m := Model{
// 		NumThreads: 4,
// 	}

// 	p := tea.NewProgram(m)
// 	go p.Start()

// 	// Handle termination signals to gracefully quit the program
// 	c := make(chan os.Signal, 1)
// 	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
// 	<-c

// 	p.Send(tea.KeyMsg{Type: tea.KeyRelease, Name: "q"})
// }

package main

import (
	"fmt"
	"os"
	"sync/atomic"
	// "os/signal"
	"strconv"
	"sync"
	"bufio"
	// "runtime"
	// "syscall"
	// tea "github.com/charmbracelet/bubbletea"
)

var (
	numbers = make(chan int)
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



	// // Worker function
	// worker := func(workerID int) {
	// 	// fmt.Println("Enter worker function")
	// 	defer func() {
	// 		// fmt.Println("Release worker pool slot")
	// 		<-workerPool // Release the slot in the worker pool
	// 		wg.Done()
	// 	}()

	// 	for {
	// 		// fmt.Println("Check if files remain")
	// 		filePath, ok := <-fileChannel

	// 		if !ok {
	// 			// No more files to process
	// 			break
	// 		}

	// 		// Process the file
	// 		// fmt.Println("parsing files")
	// 		readFiles(filePath, numbers, primes)
	// 		// fmt.Println("files parsed")

	// 		// Send progress information to the channel
	// 		fmt.Println(workerID, filePath)
	// 		progressChannel <- fmt.Sprintf("Worker %d processed file: %s", workerID, filePath)
	// 	}
	// }

// Worker function
worker := func(workerID int) {
    // fmt.Println("Enter worker function")
    defer func() {
        // fmt.Println("Release worker pool slot")
        <-workerPool // Release the slot in the worker pool
        wg.Done()
    }()

    for {
        select {
        case filePath, filesOk := <-fileChannel:
            if !filesOk { // Continue processing primes once files are done
                select {
                case prime, primesOk := <-primes:
                    if !primesOk {
                        // No more primes to process, exit the loop
                        return
                    }
                    // Process the prime
                    // fmt.Println(workerID, "Processing prime:", prime)
					prime64 := int64(prime)
					updatePrimes(prime64)
                }
            } else { // Process Files from channel
                // Process the file
                // fmt.Println("parsing files")
                readFiles(filePath, numbers, primes)
                // fmt.Println("files parsed")

                // Send progress information to the channel
                fmt.Println(workerID, filePath)
                progressChannel <- fmt.Sprintf("Worker %d processed file: %s", workerID, filePath)
            }
        }
    }
}


	// // Spawn initial goroutines
	// go func() {
	// 	fmt.Println("prespawn")
	// 	for i := 0; i < numGoroutines; i++ {
	// 		fmt.Println("initial spawn")
	// 		workerPool <- struct{}{} // Acquire a slot in the worker pool
	// 		wg.Add(1)				 // Add 1 worker to the pool
	// 		go worker(i)				 // Set worker to work
	// 		<-workerPool
	// 	}
	// }()

	// // Spawn additional goroutines dynamically
	// go func() {
	// 	fmt.Println("pre dynamic-spawn")
	// 	for {
	// 		fmt.Println("boye")
	// 		if len(fileChannel) > 0 {
	// 			fmt.Println("spawning a boye")
	// 			workerPool <- struct{}{} // Acquire a slot in the worker pool
	// 			wg.Add(1)
	// 			go worker(len(workerPool))
	// 			<-workerPool
	// 		} else {
	// 			// No more files to process
	// 			break
	// 		}
	// 	}
	// }()

	go func() {
		fmt.Println("prespawn")
		for i := 0; i < numGoroutines || (len(fileChannel) > 0 && i < numGoroutines*2); i++ {
			fmt.Println("spawn a boye")
			workerPool <- struct{}{} // Acquire a slot in the worker pool
			wg.Add(1)                 // Add 1 worker to the pool
			go worker(i)              // Set worker to work
			// <-workerPool
		}
	}()

	// fmt.Println("Closing Progress Channel")
	// close(progressChannel)
	// fmt.Println("Progress Channel Closed")

	// Create a goroutine to display progress information
	go func() {
		fmt.Println("any kind of progress display")
		defer close(progressDone)
		for progress := range progressChannel {
			fmt.Print("Progress:")
			fmt.Println(progress)
		}
	}()

	// Wait for all goroutines to complete
	// go func(){

	// 	// close(numbers)
	// 	// close(primes)
	// 	// close(progressChannel)
	// }()

	go func() {
		for prime := range primes {
			fmt.Println("Received Prime:", prime)
			// prime := int64(prime)
			// updatePrimes(prime)
		// 	// Process the prime number as needed
		}
	}()
	

	fmt.Println("enter async wait")
	wg.Wait()

	<-progressDone
	fmt.Println("progress should be done now")

	fmt.Println(primes)

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
func readFiles(file string, nums chan<- int, primes chan<- int) {
	data, err := os.Open(file)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer data.Close()

	scanner := bufio.NewScanner(data)
	for scanner.Scan() {
		line := scanner.Text()
		number, err := strconv.Atoi(line)
		if err == nil {
			// numbers = append(numbers, number)
			// fmt.Println("stashing number", number)
			// numbers <- number
			// fmt.Println("number", number, "stashed")
			if isPrime(number) {
				// fmt.Println("Stashing Prime:", number)
				primes <- number
			}
		}
	}

	// fmt.Println("post stash")
	if err := scanner.Err(); err != nil {
		fmt.Println(err)
		return
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

// func minMaxPrimeschan(primes <-chan int) {
// 	// largestPrime := -1
// 	// smallestPrime := -1

// 	for prime := range primes {
// 		// fmt.Println("Found prime:", prime)
// 		largestPrimeMutex.Lock()
// 		smallestPrimeMutex.Lock()
// 		if largestPrime == -1 || prime > largestPrime {
// 			largestPrime = prime
// 			// fmt.Print(largestPrime)
// 		}
// 		if smallestPrime == -1 || prime < smallestPrime {
// 			smallestPrime = prime

// 		}
// 		largestPrimeMutex.Unlock()
// 		smallestPrimeMutex.Unlock()
// 	}
// }


// func minMaxPrimes(primes []int) (int, int) {
// 	largestPrime := -1
// 	smallestPrime := -1

// 	for prime := range primes {
// 		if largestPrime == -1 || prime > largestPrime {
// 			largestPrime = prime
// 		}
// 		if smallestPrime == -1 || prime < smallestPrime {
// 			smallestPrime = prime
// 		}
// 	}

// 	return largestPrime, smallestPrime
// }

func updatePrimes(prime int64) {
	primesMutex.Lock()
	defer primesMutex.Unlock()
	// Update smallestPrime
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