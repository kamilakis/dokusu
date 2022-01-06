package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
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

// debug (log) level
var debug bool

// user input
var scanner *bufio.Scanner

func ilog(cat string, msg string, o ...interface{}) {
	switch cat {

	case "debug":
		if debug == true {
			m := fmt.Sprintf("%#v\n", msg) // TODO: not working as intended
			log.Println("--- DEBUG --------------")
			log.Println("---")
			log.Printf(m, o...)
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
func load(f string) ([][]Cell, error) {
	var board [][]Cell

	j, err := ioutil.ReadFile(f)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(j, &board)
	if err != nil {
		return nil, err
	}

	return board, nil
}

// save puzzle(i.e the board) state
func save(board [][]Cell) error {
	clearCells(board)

	j, err := json.MarshalIndent(&board, "", "\t")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(stateFile, j, 0600)
}

// Content prints a cell's number depending on the cell's state
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
func printBoard(b [][]Cell) {
	fmt.Printf("\n\n")
	// START first row of boxes
	fmt.Printf("   \033[0;2m" + "  0   1   2   3   4   5   6   7   8\n" + "\033[0m")
	// upper border
	fmt.Printf("   \u250F\u2501\u2501\u2501\u252F\u2501\u2501\u2501\u252F\u2501\u2501\u2501\u2533\u2501\u2501\u2501\u252F\u2501\u2501\u2501\u252F\u2501\u2501\u2501\u2533\u2501\u2501\u2501\u252F\u2501\u2501\u2501\u252F\u2501\u2501\u2501\u2513\n")
	// first row of numbers
	printNumberRow(0, b)
	// first middle row
	fmt.Printf("   \u2520\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2542\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2542\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2528\n")
	// second row of numbers
	printNumberRow(1, b)
	// second middle row
	fmt.Printf("   \u2520\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2542\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2542\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2528\n")
	// third row of numbers
	printNumberRow(2, b)
	// lower border
	fmt.Printf("   \u2523\u2501\u2501\u2501\u253F\u2501\u2501\u2501\u253F\u2501\u2501\u2501\u254B\u2501\u2501\u2501\u253F\u2501\u2501\u2501\u253F\u2501\u2501\u2501\u254B\u2501\u2501\u2501\u253F\u2501\u2501\u2501\u253F\u2501\u2501\u2501\u252B\n")
	// END first row of boxes

	// REPEAT
	// START second row of boxes (no border)
	// first row of numbers
	printNumberRow(3, b)
	// first middle row
	fmt.Printf("   \u2520\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2542\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2542\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2528\n")
	// second row of numbers
	printNumberRow(4, b)
	// second middle row
	fmt.Printf("   \u2520\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2542\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2542\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2528\n")
	// third row of numbers
	printNumberRow(5, b)
	// lower border
	fmt.Printf("   \u2523\u2501\u2501\u2501\u253F\u2501\u2501\u2501\u253F\u2501\u2501\u2501\u254B\u2501\u2501\u2501\u253F\u2501\u2501\u2501\u253F\u2501\u2501\u2501\u254B\u2501\u2501\u2501\u253F\u2501\u2501\u2501\u253F\u2501\u2501\u2501\u252B\n")
	// END second row of boxes
	// REPEAT
	// START third row of boxes (no border)

	// first row of numbers
	printNumberRow(6, b)
	// first middle row
	fmt.Printf("   \u2520\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2542\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2542\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2528\n")
	// second row of numbers
	printNumberRow(7, b)
	// second middle row
	fmt.Printf("   \u2520\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2542\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2542\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2528\n")
	// third row of numbers
	printNumberRow(8, b)
	// lower border
	fmt.Printf("   \u2517\u2501\u2501\u2501\u2537\u2501\u2501\u2501\u2537\u2501\u2501\u2501\u253B\u2501\u2501\u2501\u2537\u2501\u2501\u2501\u2537\u2501\u2501\u2501\u253B\u2501\u2501\u2501\u2537\u2501\u2501\u2501\u2537\u2501\u2501\u2501\u251B\n")
	// END third row of boxes

	fmt.Printf("\n\n")
}

// print each row between cell borders separately
// so the cell's numbers are printed (with color)
// replace 2502 with 250A or 2506 for vertical lines
func printNumberRow(row int, board [][]Cell) {
	// print row number row in gray color
	fmt.Printf(" \033[0;2m%d\033[0m", row)

	// this one line printed; broken into three for better readability
	fmt.Printf(" \u2503 %s \u2502 %s \u2502 %s \u2503", board[row][0].Content(), board[row][1].Content(), board[row][2].Content())
	fmt.Printf(" %s \u2502 %s \u2502 %s \u2503", board[row][3].Content(), board[row][4].Content(), board[row][5].Content())
	fmt.Printf(" %s \u2502 %s \u2502 %s \u2503\n", board[row][6].Content(), board[row][7].Content(), board[row][8].Content())
}

// mapValues makes a map of numbers in cells
func mapValues(board [][]Cell) map[int][]Cell {
	m := make(map[int][]Cell)

	// iterate over all cells
	for row := 0; row < 9; row++ {
		for col := 0; col < 9; col++ {
			// current cell
			c := board[row][col]

			// initial number in this cell
			number := c.Number

			// row and col fields set here;
			// initial puzzle.json file has only numbers set for each cell
			c.row = row
			c.col = col

			// store this appearance of this number
			m[number] = append(m[number], c)
		}
	}

	return m
}

// difficulty measured by the count of 0's;
// > 35 considered easy, < 25 hard
func difficulty(m map[int][]Cell) string {
	if len(m[0]) < 25 {
		return "easy"
	}
	if len(m[0]) > 30 {
		return "hard"
	}

	return fmt.Sprintf("empty cells: %d", len(m[0]))
}

func clearConsole() {
	fmt.Println("\033[2J")
}

// clear state from all cells;
// Number, row, col and marks remain
func clearCells(board [][]Cell) {
	for row := 0; row < 9; row++ {
		for col := 0; col < 9; col++ {
			board[row][col].invalid = false
			board[row][col].active = false
			board[row][col].selected = false
			board[row][col].candid = false
			board[row][col].solved = false
			board[row][col].blink = false
		}
	}
}

// get user input
func getInput() string {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Printf("\tYour choice: ")
		scanner.Scan()
		option := scanner.Text()
		return option
	}
}

func getNumber() int {
	scanner := bufio.NewScanner(os.Stdin)
	var num int
	for {
		fmt.Print("Enter a number or x to exit: ")
		scanner.Scan()

		input := scanner.Text()
		ilog("debug", "got %#+v", input)
		if input == "x" {
			num = 0
			break
		}

		i, err := strconv.Atoi(input)
		if err != nil {
			fmt.Printf("Must enter a number from 1 to 9\n")
			continue
		}
		if i > 9 || i < 1 {
			// fmt.Printf("\n")
			fmt.Print("Must enter a number from 1 to 9\n")
			continue
		}
		num = i
		break
	}
	return num
}

func play(board [][]Cell) {
	printBoard(board)
	for {
		num := getNumber()
		ilog("debug", "got %#+v", num)
		if num == 0 {
			err := save(board)
			if err != nil {
				ilog("error", "error saving: %s", err)
			}
			break // TODO: how to break loop and return to main?
		}

		// check number
		// check in row
		// check in column
		// check in 9-cell box
		// cross hatch
		// set board's state
		printBoard(board)
	}
}

func quit(board [][]Cell) {

}

func main() {
	debug = true

	// main loop
	fmt.Printf("\tOptions: (n)ew, (r)esume, e(x)it\n")
	input := getInput()
	for {
		switch input {
		case "n":
			// load puzzle from puzzle.json file
			board, err := load(puzzleFile)
			if err != nil {
				panic(err)
			}
			play(board)
			// // make a map of existing numbers in cells
			// mapv := mapValues(board)
			// ilog("info", "\tNew puzzle, difficulty: %s\n", difficulty(mapv))
			// printBoard(board)
			// input = getInput()

		case "r":
			// load previously saved puzzle in state.json
			board, err := load(stateFile)
			if err != nil {
				panic(err)
			}
			play(board)

		case "x":
			return // exit program

		default:
			fmt.Printf("\tOptions: (n)ew, (r)esume, e(x)it\n")
			input = getInput()

		}
	}
}
