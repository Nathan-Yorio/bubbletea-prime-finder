package main

import (
	"fmt"
	"os"
	// "os/signal"
	"strconv"
	// "sync"
	"bufio"
	// "syscall"
	// tea "github.com/charmbracelet/bubbletea"
)

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

func main() {
	dirPath := "./rand"

	files := getFiles(dirPath)

	fmt.Print(files)

	numbers, err := readFiles(files)
	if err != nil {
		fmt.Println("Error reading files:", err)
		os.Exit(1)
	}

	// fmt.Print(numbers)
	primes := []int{}
	for _, number := range numbers {
		if isPrime(number) == true {
			prime := number
			primes = append(primes, prime)
			// fmt.Println(primes)
			largest, smallest := minMaxPrimes(primes)
			fmt.Println("Largest:", largest, "Smallest: ", smallest)
		}
	}
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
func readFiles(filePaths []string) ([]int, error) {
	//Iterate over all the files in the directory given as an input list / array
	//go into each file and read all of the lines, grabbing each number from each line
	//this function should return all of the numbers from a single given file
	var numbers []int

	for _, filePath := range filePaths {
		file, err := os.Open(filePath)
		if err != nil {
			return nil, err
		}
		defer file.Close()

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

func minMaxPrimes(primes []int) (int, int) {
	largestPrime := -1
	smallestPrime := -1

	for _, prime := range primes {
		if isPrime(prime) {
			// fmt.Println("Found prime:", prime)
			if largestPrime == -1 || prime > largestPrime {
				largestPrime = prime
			}
			if smallestPrime == -1 || prime < smallestPrime {
				smallestPrime = prime
			}
		}
	}

	return largestPrime, smallestPrime
}
