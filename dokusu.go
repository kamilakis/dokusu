package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"
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

// generate a box (9 cells range)
func genBox(board [9][9]Cell, cell Cell) [9][9]Cell {
	ints := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	ints = shuffle(ints)
	i := 0
	for row := cell.row; row < cell.row+3; row++ {
		for col := cell.col; col < cell.col+3; col++ {
			board[row][col].Number = ints[i]
			i++
		}
	}

	return board
}

// add a number as a used (not available)
// this needed because a used number for a cell
// should not appear more than once
func addAsUsed(used []int, n int) []int {
	if len(used) == 0 {
		ilog("debug", "adding %d\n", n)
		return append(used, n)
	}

	for i := 0; i < len(used); i++ {
		ilog("debug", "checking %d with %d: ", n, used[i])
		if used[i] == n {
			ilog("debug", "skipping %d\n", n)
			return used
		}
	}
	used = append(used, n)
	ilog("debug", "adding %d, used: %#+v\n", n, used)
	return used
}

// find used numbers (not available) for a cell
func findUsed(board [9][9]Cell, cell Cell) []int {
	var used []int
	if board[cell.row][cell.col].Number > 0 {
		ilog("info", "cell [%d%d] not empty", cell.row, cell.col)
		return used
	}
	if cell.row > 9 || cell.row < 0 {
		ilog("error", "not a valid cell: [%d%d]", cell.row, cell.col)
		return used
	}
	if cell.col > 9 || cell.col < 0 {
		ilog("error", "not a valid cell: [%d%d]", cell.row, cell.col)
		return used
	}

	// check row first
	for i := 0; i < 9; i++ {
		if board[cell.row][i].Number > 0 {
			used = addAsUsed(used, board[cell.row][i].Number)
		}
	}

	// check column
	for i := 0; i < 9; i++ {
		if board[i][cell.col].Number > 0 {
			used = addAsUsed(used, board[i][cell.col].Number)
		}
	}

	// finally check box
	brow, bcol := box(cell.row, cell.col)
	for i := brow; i < brow+3; i++ {
		for j := bcol; j < bcol+3; j++ {
			if board[i][j].Number > 0 {
				used = addAsUsed(used, board[i][j].Number)
			}
		}
	}

	return used
}

// find free (available) numbers for a cell
func findFree(board [9][9]Cell, cell Cell, used []int) []int {
	var free []int

	for i := 1; i < 10; i++ {
	out:
		for j := 0; j < len(used); j++ {
			ilog("debug", "checking %d against used %d\n", i, used[j])
			if i == used[j] {
				ilog("debug", "%d is equal to used %d, check next\n", i, used[j])
				break out
			}
			if j == len(used)-1 {
				ilog("debug", "reached end of used numbers, add this %d to free\n", i)
				free = append(free, i)
				continue
			}
		}
	}
	return free
}

// load puzzle from file
func load(f string) ([9][9]Cell, error) {
	var board [9][9]Cell

	j, err := ioutil.ReadFile(f)
	if err != nil {
		return board, err
	}

	err = json.Unmarshal(j, &board)
	if err != nil {
		return board, err
	}

	return board, nil
}

// save puzzle(i.e the board) state
func save(board [9][9]Cell) error {
	clearCells(board)

	j, err := json.MarshalIndent(&board, "", "\t")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(stateFile, j, 0600)
}

// shuffle a slice of ints
func shuffle(ints []int) []int {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(ints), func(i, j int) {
		ints[i], ints[j] = ints[j], ints[i]
	})
	return ints
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
func print(b [9][9]Cell) {
	fmt.Printf("\n\n")
	// START first row of boxes
	fmt.Printf("   \033[0;2m" + "  0   1   2   3   4   5   6   7   8\n" + "\033[0m")
	// upper border
	fmt.Printf("   \u250F\u2501\u2501\u2501\u252F\u2501\u2501\u2501\u252F\u2501\u2501\u2501\u2533\u2501\u2501\u2501\u252F\u2501\u2501\u2501\u252F\u2501\u2501\u2501\u2533\u2501\u2501\u2501\u252F\u2501\u2501\u2501\u252F\u2501\u2501\u2501\u2513\n")
	// first row of numbers
	printRow(0, b)
	// first middle row
	fmt.Printf("   \u2520\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2542\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2542\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2528\n")
	// second row of numbers
	printRow(1, b)
	// second middle row
	fmt.Printf("   \u2520\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2542\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2542\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2528\n")
	// third row of numbers
	printRow(2, b)
	// lower border
	fmt.Printf("   \u2523\u2501\u2501\u2501\u253F\u2501\u2501\u2501\u253F\u2501\u2501\u2501\u254B\u2501\u2501\u2501\u253F\u2501\u2501\u2501\u253F\u2501\u2501\u2501\u254B\u2501\u2501\u2501\u253F\u2501\u2501\u2501\u253F\u2501\u2501\u2501\u252B\n")
	// END first row of boxes

	// REPEAT
	// START second row of boxes (no border)
	// first row of numbers
	printRow(3, b)
	// first middle row
	fmt.Printf("   \u2520\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2542\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2542\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2528\n")
	// second row of numbers
	printRow(4, b)
	// second middle row
	fmt.Printf("   \u2520\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2542\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2542\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2528\n")
	// third row of numbers
	printRow(5, b)
	// lower border
	fmt.Printf("   \u2523\u2501\u2501\u2501\u253F\u2501\u2501\u2501\u253F\u2501\u2501\u2501\u254B\u2501\u2501\u2501\u253F\u2501\u2501\u2501\u253F\u2501\u2501\u2501\u254B\u2501\u2501\u2501\u253F\u2501\u2501\u2501\u253F\u2501\u2501\u2501\u252B\n")
	// END second row of boxes
	// REPEAT
	// START third row of boxes (no border)

	// first row of numbers
	printRow(6, b)
	// first middle row
	fmt.Printf("   \u2520\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2542\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2542\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2528\n")
	// second row of numbers
	printRow(7, b)
	// second middle row
	fmt.Printf("   \u2520\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2542\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2542\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2528\n")
	// third row of numbers
	printRow(8, b)
	// lower border
	fmt.Printf("   \u2517\u2501\u2501\u2501\u2537\u2501\u2501\u2501\u2537\u2501\u2501\u2501\u253B\u2501\u2501\u2501\u2537\u2501\u2501\u2501\u2537\u2501\u2501\u2501\u253B\u2501\u2501\u2501\u2537\u2501\u2501\u2501\u2537\u2501\u2501\u2501\u251B\n")
	// END third row of boxes

	fmt.Printf("\n\n")
}

// print each row between cell borders separately
// so the cell's numbers are printed (with color)
// replace 2502 with 250A or 2506 for vertical lines
func printRow(row int, board [9][9]Cell) {
	// print row number row in gray color
	fmt.Printf(" \033[0;2m%d\033[0m", row)

	// this one line printed; broken into three for better readability
	fmt.Printf(" \u2503 %s \u2502 %s \u2502 %s \u2503", board[row][0].Content(), board[row][1].Content(), board[row][2].Content())
	fmt.Printf(" %s \u2502 %s \u2502 %s \u2503", board[row][3].Content(), board[row][4].Content(), board[row][5].Content())
	fmt.Printf(" %s \u2502 %s \u2502 %s \u2503\n", board[row][6].Content(), board[row][7].Content(), board[row][8].Content())
}

// mapValues makes a map of numbers in cells
func mapValues(board [9][9]Cell) map[int][]Cell {
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
func clearCells(board [9][9]Cell) {
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
		fmt.Print("Enter a number or Ctrl-c to exit: ")
		scanner.Scan()
		input := scanner.Text()
		i, err := strconv.Atoi(input)
		if err != nil {
			fmt.Printf("Must enter a number from 1 to 9\n")
			continue
		}
		if i > 9 || i < 1 {
			fmt.Print("Must enter a number from 1 to 9\n")
			continue
		}
		num = i
		break
	}
	return num
}

func checkRow(board [9][9]Cell, num int, row int) interface{} {
	for col := 0; col < 9; col++ {
		if board[row][col].Number == num {
			return fmt.Sprintf("number %d found in cell [%d%d]", num, row, col)
		}
	}
	return nil
}

func checkCol(board [9][9]Cell, num int, col int) interface{} {
	for row := 0; row < 9; row++ {
		if board[row][col].Number == num {
			return fmt.Sprintf("number %d found in cell [%d%d]", num, row, col)
		}
	}
	return nil
}

// box returns the 9-cell range (box) that a cell belongs in
// i.e the upper left cell's row and column
func box(row, col int) (int, int) {
	crd3 := row / 3
	ccd3 := col / 3

	srow := crd3 * 3
	scol := ccd3 * 3

	return srow, scol
}

func checkBox(board [9][9]Cell, num int, row int, col int) interface{} {
	srow, scol := box(row, col)
	for row := srow; row < srow+3; row++ {
		for col := scol; col < scol+3; col++ {
			if board[row][col].Number == num {
				return fmt.Sprintf("number %d found in cell [%d%d]", num, row, col)
			}
		}
	}
	return nil
}

func play(board [9][9]Cell) {
	print(board)
	for {
		num := getNumber()
		ilog("debug", "got %d", num)
		err := save(board)
		if err != nil {
			ilog("error", "error saving: %s", err)
		}

		// check number
		// check in row
		// check in column
		// check in 9-cell box
		// cross hatch
		// set board's state
		print(board)
	}
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
			// print(board)
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
