// Something is wrong with the way I've merged the logic of choice and cursor
// that's why my view selection is not updating, I'll have to figure out what's up with that
package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"
	"strconv"
	"sync/atomic"
	"sync"
	"bufio"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/indent"
)

//globals
var (
	smallestPrime int64 = -1
	largestPrime  int64 = -1
	primesMutex   sync.Mutex
	workersSpawned bool
	primes = make(chan int)
	updateChannel chan string
	workerCount int
)

// 📜📜📜📜📜📜~~WHAT CONTENT IS UPDATING~~📜📜📜📜📜📜

// Functions that return messages to other functions
type (
	frameMsg struct{
	}
)

func frame() tea.Cmd {
	return tea.Tick(time.Second/60, func(time.Time) tea.Msg {
		return frameMsg{}
	})
}

type model struct {
	programsPath string

	options        []string // list of options to choose from
	runningOptions []string // list of options after program starts running

	selected map[int]struct{} //unholy witchcraft, do not touch

	Choice       int //stores value of cursor position for first selection
	secondChoice int //stores value of cursor position for second selection

	Chosen       bool
	secondChosen bool

	Frames   int
	Quitting bool

	optionOne   string
	optionTwo   string

	renderFlag bool

	dirPath string
	running bool
	progress string
	numWorkers int
	currentFile chan string
}

func SelectModel() model {
	return model{
		options: []string{ // Options on program start
			"Run Program",
			"Increment Start Workers",
			"Decrement Start Workers",
			"Exit",
		},
		runningOptions: []string{ // Options while program is running
			"Increment Start Workers",
			"Decrement Start Workers",
			"Exit",
		},
		selected:     make(map[int]struct{}), //mathematical set mapping for choice selection
		Choice:       0,
		secondChoice: 0,
		Chosen:       false,
		secondChosen: false,
		Frames:       0,
		Quitting:     false,
		optionOne:    "",
		optionTwo:    "",
		renderFlag:   false, //used to wait and render checkmark before moving on
		dirPath:      "./rand",
		running:      false,
		numWorkers:   0,
		currentFile:  make(chan string, 1000),
	}
}

// 🏁🏁🏁🏁🏁🏁~~~~~Initialize Commands~~~~~🏁🏁🏁🏁🏁🏁
func (m model) Init() tea.Cmd {
	return nil //don't need to have an initially running function
}

// 🤔🤔🤔🤔🤔🤔 ~~~~~~~ Logik ~~~~~~~~~~~~~~🤔🤔🤔🤔🤔🤔

// Main update function.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Make sure these keys always quit
	if msg, ok := msg.(tea.KeyMsg); ok {
		k := msg.String()
		if k == "q" || k == "esc" || k == "ctrl+c" {
			m.Quitting = true
			return m, tea.Quit
		}
	}

	// Hand off the message and model to the appropriate update function for the
	// appropriate view based on the current state.
	if !m.Chosen && !m.secondChosen {
		return updateProgChoice(msg, m)
	} else  {
		return updateOptionChoice(msg, m)
	} 
}

// View Update 1 ~~~ Starting Runners, Incrementing Workers
func updateProgChoice(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			m.Choice++
			if m.Choice > len(m.options)-1 { //don't allow to exceed array bounds
				m.Choice = len(m.options) - 1
			}
			if m.Choice > (m.Frames+1)*10-1 && m.Frames < len(m.options)/4 {
				m.Frames++
			}
		case "k", "up":
			m.Choice--
			if m.Choice < 0 { //don't allow to exceed array bounds
				m.Choice = 0
			}
			if m.Choice < m.Frames*10 && m.Frames > 0 {
				m.Frames--
			}
		case "enter":
			// Store that we've chosen the first option, save the choice
			m.Chosen = true
			m.optionOne = m.options[m.Choice]
			switch m.optionOne {
			case "Run Program":
				m.running = true
				go primeRipper(m, updateChannel)
			case "Increment Start Workers":
				m.numWorkers++
				m.Chosen = false
			case "Decrement Start Workers":
				m.numWorkers--
				if m.numWorkers < 0 {
					m.numWorkers = 0
				}
				m.Chosen = false
			case "Exit":
				//quit the program
				m.Quitting = true
				return m, tea.Quit
			}
			return m, frame()
		}
	}
	return m, nil
}

// View Update 2 ~~~ What to do with the program
func updateOptionChoice(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			m.secondChoice++
			if m.secondChoice > len(m.runningOptions)-1 { //don't allow cursor to exceed bounds
				m.secondChoice = len(m.runningOptions) - 1
			}
		case "k", "up":
			m.secondChoice--
			if m.secondChoice < 0 { // don't allow cursor to exceed bounds
				m.secondChoice = 0
			}
		case "enter":
			m.secondChosen = true
			m.optionTwo = m.runningOptions[m.secondChoice] //store the user's second choice
			switch m.optionTwo {
			case "Increment Start Workers":
				m.numWorkers++
				m.secondChosen = false
			case "Decrement Start Workers":
				m.numWorkers--
				if m.numWorkers < 0 {
					m.numWorkers = 0
				}
				m.secondChosen = false
			case "Exit":
				//quit the program
				m.Quitting = true
				return m, tea.Quit
			}
			return m, frame()
		}
	}

	return m, nil
}

// 👀👀👀👀👀~~~~~~~Wut MY USER SEES~~~~~~👀👀👀👀👀👀👀
// The main view, which just calls the appropriate sub-view
func (m model) View() string {
	var s string
	if m.Quitting {
		return "\n  Program Finished\n\n"
	}
	if !m.Chosen { //have we made our first choice? && !m.renderFlag, was a test
		s = listOptions(m)
	} 
	if m.Chosen {
		s = showStats(m)
	}
	return indent.String("\n"+s+"\n\n", 2)
}

// Subview 1 ~~~ Show initial options, allow incrementing workers
func listOptions(m model) string {
	header := headerStyle.
		Render("TUI Program to calculate primes:")

	// Max number of options to display per terminal page
	optionsPerPage := 5

	// Calculate the starting and ending index of options to display
	startIndex := m.Frames * optionsPerPage
	endIndex := (m.Frames + 1) * optionsPerPage

	// If the ending index is greater than the total number of programs, set it to the total number of programs
	if endIndex > len(m.options) {
		endIndex = len(m.options)
	}

	// Iterate over the programs in the current page
	optionsList := ""
	for i := startIndex; i < endIndex; i++ {
		// Is the cursor pointing at this choice?
		cursor := unselectedStyle
		icon := ""
		if m.Choice == i {
			icon = "🫠"
			cursor = selectedStyle
		} else {
			icon = ""
		}

		option := cursor.Render("ゝ " + m.options[i])
		optionsList += listStyle.Render(option + cursorStyle.Render(icon))
	}

	fillerSpace := ""
	emptyLines := 5
	for i := 0; i < emptyLines ; i++ {
		fillerSpace += listStyle.Render("")
	}

	debugVars := statStyle.
		Render(fmt.Sprint(
			"selected", "m.selected,", "\n",
			"choice:", "m.Choice", "\n",
			"secondChoice:", m.secondChoice, "\n",
			"Chosen:", m.Chosen, "\n",
			"secondChosen:", m.secondChosen, "\n",
			"Frames:", m.Frames, "\n",
			"Quitting:", m.Quitting, "\n",
			"optionOne:", m.optionOne, "\n",
			"optionTwo:", m.optionTwo, "\n",
			"renderFlag:", m.renderFlag, "\n",
			"dirPath:", m.dirPath, "\n",
			"running:", m.running, "\n",
			"numWorkers:", m.numWorkers, "\n",
			"spawned?:", workersSpawned , "\n",
			"currentFile: ", m.currentFile, "\n",
		))

	workers := statStyle.
		Render("workers:", fmt.Sprint(m.numWorkers))
	// The footer
	footer := footerStyle.
		Render("Press j or down arrow to scroll down, k or up arrow to scroll up.\n" + "Press enter to select an option.\n" + "Press q, esc, or ctrl-c to quit.\n")

	if m.Quitting {
		farewell := "\n  さようなら!\n\n"
		return farewell
	} else {
		return header + optionsList + fillerSpace + footer + workers + debugVars
	}
}

func showStats(m model) string {
	header := headerStyle.
		Render("Calculating Primes in dir:", m.dirPath)

	// Max number of options to display per terminal page
	optionsPerPage := 5

	// Calculate the starting and ending index of options to display
	startIndex := m.Frames * optionsPerPage
	endIndex := (m.Frames + 1) * optionsPerPage

	// If the ending index is greater than the total number of programs, set it to the total number of programs
	if endIndex > len(m.runningOptions) {
		endIndex = len(m.runningOptions)
	}

	// Iterate over the programs in the current page
	optionsList := ""
	for i := startIndex; i < endIndex; i++ {
		// Is the cursor pointing at this choice?
		cursor := unselectedStyle
		icon := ""
		if m.secondChoice == i {
			icon = "🔥"
			cursor = selectedStyle
		} else {
			icon = ""
		}

		option := cursor.Render("ゝ " + m.runningOptions[i])
		optionsList += listStyle.Render(option + cursorStyle.Render(icon))
	}

	fillerSpace := ""
	emptyLines := 5
	for i := 0; i < emptyLines ; i++ {
		fillerSpace += listStyle.Render("")
	}

	workers := statStyle.
		Render("workers:", fmt.Sprint(m.numWorkers))
	threadProg := statStyle.
		Render(m.progress)
	sPrime 	:= statStyle.
		Render(fmt.Sprint("Smallest Prime: ",smallestPrime))
	lPrime  := statStyle.
		Render(fmt.Sprint("Largest Prime: ", largestPrime))

	debugVars := statStyle.
		Render(fmt.Sprint(
			"selected", "m.selected,", "\n",
			"choice:", "m.Choice", "\n",
			"secondChoice:", m.secondChoice, "\n",
			"Chosen:", m.Chosen, "\n",
			"secondChosen:", m.secondChosen, "\n",
			"Frames:", m.Frames, "\n",
			"Quitting:", m.Quitting, "\n",
			"optionOne:", m.optionOne, "\n",
			"optionTwo:", m.optionTwo, "\n",
			"renderFlag:", m.renderFlag, "\n",
			"dirPath:", m.dirPath, "\n",
			"running:", m.running, "\n",
			"numWorkers:", m.numWorkers, "\n",
			"spawned?:", workersSpawned , "\n",
			"runningWorkers: ", workerCount, "\n",
		))

	currentFile := footerStyle.
		Render(fmt.Sprint("currentFile: ", <-m.currentFile, "\n",))

	footer := footerStyle.
		Render("Press j or down arrow to scroll down, k or up arrow to scroll up.\n" + "Press enter to select an option.\n" + "Press q, esc, or ctrl-c to quit.\n")

	if m.Quitting {
		farewell := "\n  さようなら!\n\n"
		return farewell
	} else {
		return header + optionsList + fillerSpace +  footer + workers + sPrime + lPrime + threadProg + debugVars + currentFile
	}
}

// 🔥🔥🔥🔥🔥~~~~~~Make the Magic Happen~~~~~~🔥🔥🔥🔥🔥
func main() {
	//clear the terminal window before starting MVC
	clear := exec.Command("clear")
	clear.Stdout = os.Stdout
	clear.Run()

	p := tea.NewProgram(SelectModel())
	if _, err := p.Run(); err != nil {
		fmt.Println("could not start program:", err)
	}
}

func primeRipper(m model, updateChannel chan string) {
	StartNow := time.Now()
	files := getFiles(m.dirPath)

	fileChannel := make(chan string, 1000) // channel to store the files in buffer 1000
	for _,file := range files {
		fileChannel <- file
	}
	workersSpawned = true

	// Define the number of goroutines to use (e.g., 4 for quadrupling throughput (limited severely by IOPs))
	numGoroutines := m.numWorkers

	var fileGroup sync.WaitGroup

	workers := func(id int, fileChannel chan string, wg *sync.WaitGroup) {
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
				m.currentFile <- fmt.Sprint("Worker: ", id, " Processing ", filePath)
			}
		}
	}

	for i := 0; i < numGoroutines; i++ {
		fileGroup.Add(1)
		go workers(i, fileChannel, &fileGroup)
		workerCount += 1
	}
	
	// Asynchronously read the primes out of the channel while the other function is writing them in
	go func () {
		for prime := range primes {
			prime64 := int64(prime)
			updatePrimes(prime64)
		}
	}()

	fileGroup.Wait()

	fmt.Println("that took:", time.Since(StartNow))
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
			if atomic.CompareAndSwapInt64(&smallestPrime, currentSmallest, prime) {
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
				break
			}
		} else {
			break
		}
	}
}

// 💄
// Lip Gloss styles
var (
	// Define some common styles
	cursorStyle = lipgloss.NewStyle().
			Align(lipgloss.Right).
			Margin(0, 0, 0, 0).
			Padding(0, 0, 0, 0)

	selectedStyle = lipgloss.NewStyle().
			Width(67). //room for the icon
			Foreground(lipgloss.Color("#ffffff")).
			Background(lipgloss.Color("#6B1509"))
	
	unselectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Background(lipgloss.Color("#0B0815")).
			Width(70)

	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E0E2E4")).
			Background(lipgloss.Color("#194F3B")).
			Width(70).
			Height(0).
			PaddingBottom(0).
			Align(lipgloss.Center)

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E0E2E4")).
			Background(lipgloss.Color("#194F3B")).
			Width(70).
			Padding(0, 0, 0, 0).
			Margin(1, 0, 0, 0).
			Align(lipgloss.Center)

	statStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E0E2E4")).
			Background(lipgloss.Color("#080B5F")).
			Width(70).
			Padding(0, 0, 0, 0).
			Margin(1, 0, 0, 0).
			Align(lipgloss.Center)


	listStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Background(lipgloss.Color("232")).
			Width(70).
			Margin(1, 0, 0, 0).
			Align(lipgloss.Left)

	centerStyle = lipgloss.NewStyle().
			Align(lipgloss.Center)
)