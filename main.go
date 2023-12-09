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
	// "os/signal"
	"strconv"
	"sync"
	"bufio"
	"runtime"
	// "syscall"
	// tea "github.com/charmbracelet/bubbletea"
)

var (
	numbers = make(chan int)
	largestPrime = -1
	largestPrimeMutex sync.Mutex
	smallestPrime = -1
	smallestPrimeMutex sync.Mutex
)


func main() {
	dirPath := "./rand"

	files := getFiles(dirPath)

	// fmt.Print(files)

	var consumer sync.WaitGroup
	var producer sync.WaitGroup
	var primality sync.WaitGroup
	primes := make(chan int)


	// Define the number of goroutines to use (e.g., 4 for quadrupling)
	numGoroutines := 5

	// Producer

	for _,file := range files {
		for i := 0; i < numGoroutines; i++ {
			producer.Add(1)
			go func(file string) {
				defer producer.Done()
				readFiles(file, numbers)
			}(file)
		}	
	}

	go func() {
		producer.Wait()
		// close(numbers) //fileChannel
	}()
	// Consumer I
	for i := 0; i < numGoroutines; i++ {
		consumer.Add(1)
		go func(workerID int) {
			defer consumer.Done()
            for j := workerID; ; j += numGoroutines {
                // Receive value from the channel, break if the channel is closed
                number := <-numbers
				// number, more := <-numbers
                // if !more {
                //     break
                // }

                if isPrime(number) {
                    primes <- number
                }
            }
		}(i)
	}

	go func() {
		consumer.Wait()
		// close(primes)
		// close(numbers) //fileChannel
	}()



	for i := 0; i < numGoroutines; i++ {
		primality.Add(1)
		go func(workerID int) {
			defer primality.Done()
            for k := workerID; ; k += numGoroutines {
			// primeResults := []int{}
			// primeResults = append(primeResults, prime)
				minMaxPrimeschan(primes)
            }
		}(i)
	}

	// go func() {
	// 	close(primes)
	// 	close(numbers) 
	// }()

	go func() {
		primality.Wait()
	}()

	fmt.Println("Largest:", largestPrime, "Smallest:", smallestPrime)
	routines := runtime.NumGoroutine()
	fmt.Println("Number of active goroutines:", routines)

}

// func incrementProgress() {
// 	processedMutex.Lock()
// 	numbersProcessed++
// 	processedMutex.Unlock()
// }

// 📑 what files are in a given directory
func getFiles(dirPath string) (the_files []string) {
	files, _ := os.ReadDir(dirPath)
	for _, file := range files {
		the_files = append(the_files, dirPath+"/"+file.Name()) // make a list of the files in the chosen directory
	}
	return the_files
}

// Producer
func readFiles(file string, nums chan<- int) {
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
			numbers <- number
		}
	}

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

func minMaxPrimeschan(primes <-chan int) {
	// largestPrime := -1
	// smallestPrime := -1

	for prime := range primes {
		// fmt.Println("Found prime:", prime)
		largestPrimeMutex.Lock()
		smallestPrimeMutex.Lock()
		if largestPrime == -1 || prime > largestPrime {
			largestPrime = prime
		}
		if smallestPrime == -1 || prime < smallestPrime {
			smallestPrime = prime

		}
		largestPrimeMutex.Unlock()
		smallestPrimeMutex.Unlock()
	}
}


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
