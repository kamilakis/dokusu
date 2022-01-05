package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"log"
)

// Cell represents each of 81 board's cells
type Cell struct {
	Number   int
	row      int
	col      int
	invalid  bool // those have a red or green color
	active   bool
	selected bool  // used for cross-hatching
	candid   bool  // possible solution for current number
	marks    []int // other solutions (if number = 0)
	solved   bool
	blink    bool
}

const (
	cReset      = "0m"
	cBright     = "1m"
	cDim        = "2m"
	cUnderscore = "4m"
	cBlink      = "5m"
	cReverse    = "7m"
	cHidden     = "8m"

	cFgBlack   = "30m"
	cFgRed     = "31m"
	cFgGreen   = "32m"
	cFgYellow  = "33m"
	cFgBlue    = "34m"
	cFgMagenta = "35m"
	cFgCyan    = "36m"
	cFgWhite   = "37m"

	cBgBlack   = "40m"
	cBgRed     = "41m"
	cBgGreen   = "42m"
	cBgYellow  = "43m"
	cBgBlue    = "44m"
	cBgMagenta = "45m"
	cBgCyan    = "46m"
	cBgWhite   = "47m"
)

// stateFile is where games are saved before exit
var stateFile = "state.json"

// puzzleFile is where games are loaded from
var puzzleFile = "puzzle.json"

// ValuesMap holds all cells that a number appears in
var ValuesMap map[int][]Cell

// debug (log) level
var debug bool

// Board holds the current state of the sudoku puzzle
var Board [9][9]Cell

// user input
var scanner *bufio.Scanner

func ilog(cat string, msg string, o ...interface{}) {
	switch cat {

	case "debug":
		if debug == true {
			log.Println("--- DEBUG --------------")
			log.Println("---")
			log.Printf("---%+v\n", msg)
			log.Println("---")
			log.Println("--- DEBUG --------------")
		}
		return

	default:
		log.Printf(msg, o...)
		return
	}
}

// load puzzle from file
func load(f string) error {

	j, err := ioutil.ReadFile(f)
	if err != nil {
		return err
	}

	err = json.Unmarshal(j, &Board)
	if err != nil {
		return err
	}

	return nil
}

// save puzzle state
func save() error {
	clearCells()

	j, err := json.MarshalIndent(&Board, "", "\t")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(stateFile, j, 0600)
}

// Content prints a cell's number
// see structs for available colors
func (c Cell) Content() string {
	var number, color string
	color = cFgWhite // default is white foreground color

	if c.invalid {
		color = cFgRed
	}
	if c.solved {
		color = cFgYellow
	}
	if c.active {
		color = cFgMagenta
	}
	if c.Number == 0 {
		number = " " // zero-numbered cells shown as empty
	} else {
		number = fmt.Sprintf("%d", c.Number)
	}
	if c.candid {
		color = cBgGreen
	}
	if c.blink {
		color = cBlink
	}

	return "\033[0;" + color + number + "\033[0m"
}

// prints the board with the cells contents if num not zero
func printBoard() {
	fmt.Printf("\n\n")
	// START first row of boxes
	fmt.Printf("   \033[0;2m" + "  0   1   2   3   4   5   6   7   8\n" + "\033[0m")
	// upper border
	fmt.Printf("   \u250F\u2501\u2501\u2501\u252F\u2501\u2501\u2501\u252F\u2501\u2501\u2501\u2533\u2501\u2501\u2501\u252F\u2501\u2501\u2501\u252F\u2501\u2501\u2501\u2533\u2501\u2501\u2501\u252F\u2501\u2501\u2501\u252F\u2501\u2501\u2501\u2513\n")
	// first row of numbers
	printNumberRow(0)
	// first middle row
	fmt.Printf("   \u2520\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2542\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2542\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2528\n")
	// second row of numbers
	printNumberRow(1)
	// second middle row
	fmt.Printf("   \u2520\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2542\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2542\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2528\n")
	// third row of numbers
	printNumberRow(2)
	// lower border
	fmt.Printf("   \u2523\u2501\u2501\u2501\u253F\u2501\u2501\u2501\u253F\u2501\u2501\u2501\u254B\u2501\u2501\u2501\u253F\u2501\u2501\u2501\u253F\u2501\u2501\u2501\u254B\u2501\u2501\u2501\u253F\u2501\u2501\u2501\u253F\u2501\u2501\u2501\u252B\n")
	// END first row of boxes

	// REPEAT
	// START second row of boxes (no border)
	// first row of numbers
	printNumberRow(3)
	// first middle row
	fmt.Printf("   \u2520\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2542\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2542\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2528\n")
	// second row of numbers
	printNumberRow(4)
	// second middle row
	fmt.Printf("   \u2520\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2542\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2542\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2528\n")
	// third row of numbers
	printNumberRow(5)
	// lower border
	fmt.Printf("   \u2523\u2501\u2501\u2501\u253F\u2501\u2501\u2501\u253F\u2501\u2501\u2501\u254B\u2501\u2501\u2501\u253F\u2501\u2501\u2501\u253F\u2501\u2501\u2501\u254B\u2501\u2501\u2501\u253F\u2501\u2501\u2501\u253F\u2501\u2501\u2501\u252B\n")
	// END second row of boxes
	// REPEAT
	// START third row of boxes (no border)

	// first row of numbers
	printNumberRow(6)
	// first middle row
	fmt.Printf("   \u2520\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2542\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2542\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2528\n")
	// second row of numbers
	printNumberRow(7)
	// second middle row
	fmt.Printf("   \u2520\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2542\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2542\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2528\n")
	// third row of numbers
	printNumberRow(8)
	// lower border
	fmt.Printf("   \u2517\u2501\u2501\u2501\u2537\u2501\u2501\u2501\u2537\u2501\u2501\u2501\u253B\u2501\u2501\u2501\u2537\u2501\u2501\u2501\u2537\u2501\u2501\u2501\u253B\u2501\u2501\u2501\u2537\u2501\u2501\u2501\u2537\u2501\u2501\u2501\u251B\n")
	// END third row of boxes

	fmt.Printf("\n\n")
}

// print each row between cell borders separately
// so the cell's numbers are printed (with color)
// replace 2502 with 250A or 2506 for vertical lines as cells separators
func printNumberRow(n int) {
	// print row number n in gray color
	fmt.Printf(" \033[0;2m%d\033[0m", n)

	// this one line printed; broken into three for better readability
	fmt.Printf(" \u2503 %s \u2502 %s \u2502 %s \u2503", Board[n][0].Content(), Board[n][1].Content(), Board[n][2].Content())
	fmt.Printf(" %s \u2502 %s \u2502 %s \u2503", Board[n][3].Content(), Board[n][4].Content(), Board[n][5].Content())
	fmt.Printf(" %s \u2502 %s \u2502 %s \u2503\n", Board[n][6].Content(), Board[n][7].Content(), Board[n][8].Content())
}

// validates given puzzle and makes a map of numbers in cells
func mapValues() error {
	ValuesMap = make(map[int][]Cell)

	// iterate over all cells
	for row := 0; row < 9; row++ {
		for col := 0; col < 9; col++ {
			// current cell
			c := Board[row][col]

			// initial number in this cell
			number := c.Number

			// row and col fields set here;
			// initial puzzle.json file has only numbers set for each cell
			c.row = row
			c.col = col

			// check for invalid numbers
			// for empty cells we use 0
			if number < 0 || number > 9 {
				return fmt.Errorf("invalid number %d in cell %d%d", number, c.row, c.col)
			}

			// check if initial number is valid for this cell
			// puzzleErrors = c.checkCell(number)

			// register this appearance of this number
			ValuesMap[number] = append(ValuesMap[number], c)
		}
	}

	// if len(puzzleErrors) > 0 {
	// 	for i := 0; i < len(puzzleErrors); i++ {
	// 		// ilog("info", puzzleErrors[i])
	// 		fmt.Println(puzzleErrors[i])
	// 	}
	// }

	return nil
}

// difficulty measured by the count of 0's;
// > 35 considered easy, < 25 hard
func difficulty() string {
	if len(ValuesMap[0]) < 25 {
		return "easy"
	}
	if len(ValuesMap[0]) > 30 {
		return "hard"
	}

	return fmt.Sprintf("empty cells: %d", len(ValuesMap[0]))
}

func clearConsole() {
	fmt.Println("\033[2J")
}

// clear state from all cells;
// Number, row, col and marks remain 
func clearCells() {
	for i := 0; i < 9; i++ {
		for j := 0; j < 9; j++ {
			Board[i][j].invalid = false
			Board[i][j].active = false
			Board[i][j].selected = false
			Board[i][j].candid = false
			Board[i][j].solved = false
			Board[i][j].blink = false
		}
	}
}

// get user input
func getInput() string {
	var option string

	fmt.Printf("\tOptions: (n)ew, (r)esume, (q)uit\n")
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Printf("\tOr a number: ")
		scanner.Scan()
		option = scanner.Text()
		return option
	}
}

func main() {
	var input string
	// main loop
	input = getInput()
	for {
		switch input {
		case "n":
			// load puzzle from puzzle.json file
			if err := load(puzzleFile); err != nil {
				panic(err)
			}
			// make a map of existing numbers in cells
			if err := mapValues(); err != nil {
				panic(err)
			}
			ilog("info", "\n\tNew puzzle, difficulty: %s\n", difficulty())
			printBoard()
			input = getInput()

		case "r":
			// load previously saved puzzle in state.json
			if err := load(stateFile); err != nil {
				panic(err)
			}
			// make a map of existing numbers in cells
			if err := mapValues(); err != nil {
				panic(err)
			}
			printBoard()
			input = getInput()

		case "q":
			// save puzzle state
			if err := save(); err != nil {
				ilog("error", "error saving puzzle: %s", err)
			}
			ilog("info", "puzzle saved at %s.", stateFile)
			return // exit program

		default:
			i, err := strconv.Atoi(input)
			if err != nil {
				input = getInput()
			}
			if i > 9 || i < 1 {
				// fmt.Printf("\n")
				fmt.Print("Must enter a number from 1 to 9\n")
				input = getInput()
			}
			ilog("info", "chose %d", i)
			input = getInput()
		}
	}
}
