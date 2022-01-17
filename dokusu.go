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
	color    string
	invalid  bool // those have a red or green color
	active   bool
	selected bool  // used for cross-hatching
	candid   bool  // possible solution for current number
	marks    []int // other solutions (if number = 0)
	solved   bool
	blink    bool
}

// Board 
type Board [9][9]Cell

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

// print cell in [rowcol] format, e.g. [04]
func (c Cell) String() string {
	return fmt.Sprintf("[%d%d]", c.row, c.col)
}

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
		fmt.Printf(msg, o...)
		return
	}
}

// create a new empty board
func board() Board {
	b := Board{}
	for row := 0; row < 9; row++ {
		for col := 0; col < 9; col++ {
			b[row][col].row = row
			b[row][col].col = col
		}
	}
	return b
}

// load puzzle from file
func (b *Board) load(f string) error {
	j, err := ioutil.ReadFile(f)
	if err != nil {
		return err
	}

	err = json.Unmarshal(j, b)
	if err != nil {
		return err
	}

	return nil
}

// save puzzle(i.e the board) state
func (b *Board) save() error {
	b.clear()

	j, err := json.MarshalIndent(b, "", "\t")
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

// generate randomly a 3x3 box (9 cells range)
func (b *Board) genBox(c Cell) {
	ilog("debug", "show *Board b: %#+v", b)
	ints := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	ints = shuffle(ints)
	i := 0
	for row := c.row; row < c.row+3; row++ {
		for col := c.col; col < c.col+3; col++ {
			b[row][col].Number = ints[i]
			i++
		}
	}
}

// generate randomly first three 3x3 boxes
func (b *Board) gen3boxes() {
	ints := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	ints = shuffle(ints)

	c := Cell{row: 0, col: 0}
	b.genBox(c)
	c = Cell{row: 3, col: 3}
	b.genBox(c)
	c = Cell{row: 6, col: 6}
	b.genBox(c)
}

// set value for a cell
func (b *Board) setValue(r int, c int, v int) {
	b[r][c].Number = v
	ilog("info", " [%d%d] set to %d\n", r, c, v)
}

// check marks for a cell
func (b *Board) checkMarks(c Cell) int {
	ilog("info", "check marks for %s%v:\n", c, c.marks)
	if len(c.marks) == 0 {
		ilog("info", "no marks available for %s\n", c)
		return 0 // returning 0 actually means a failed check
	}
	if len(c.marks) == 1 {
		ilog("info", "only one available mark for %s\n", c)
		return c.marks[0] // returning the one and only available mark
	}
	for _, mark := range c.marks {
		check := b.checkNum(mark, c.row, c.col)
		if check != nil {
			ilog("info", "mark %d not fit for %s: %v\n", mark, c, check)
		} else {
			return mark
		}
	}

	ilog("info", "marks exchausted for %s\n", c)
	return 0
}

// swap values of two cells; an attempt to different solutions
// TODO: exclude completed boxes?
func (b *Board) swap(c1 Cell, c2 Cell) {
	ilog("info", "swap %s with %s\n", c1, c2)
	b.setValue(c1.row, c1.col, c2.Number)
	b.setValue(c2.row, c2.col, c1.Number)
}

// find conflicting number in this cell
// check only row and column, cannot swap a box
func (b *Board) findConflict(c Cell) Cell {
	for col := 0; col < 9; col++ {
		if b[c.row][col].col == col {
			continue
		}
		if b[c.row][col].Number == c.Number {
			return Cell{row:c.row, col:col}
		}
	}
	for row := 0; row < 9; row++ {
		if b[row][c.col].row == row {
			continue
		}
		if b[row][c.col].Number == c.Number {
			return Cell{row:row, col:c.col}
		}
	}

	// TODO: change function return value
	return Cell{row:0, col:0}
}

// try setting all cells one by one
func (b *Board) setAll(retry int) bool {
	// out:
	for row := 0; row < 9; row++ {
		col:
		for col := 0; col < 9; col++ {
			c := b[row][col]
			if c.Number > 0 {
				ilog("info", "cell %s already set with %d\n", c, c.Number)
				continue col
			}
			mark := b.checkMarks(c)
			if mark == 0 { // marks exchausted or not available for this cell
				b.swap(c, b.findConflict(c))
				return false // no marks available for cell

			}
			b.setValue(row, col, mark)
			continue col
		}
	}

	return true
}
// recursively fill a 3x3 box
// find free numbers available for each cell
// if none found start again with next free number
func (b *Board) fillBox(c Cell, t int, seq []int) {
	// stop madness
	if t > 10 {
		ilog("info", "cannot fill; quitting.")
		return
	}

	if t == 0 && b[c.row][c.col].Number > 0 {
		ilog("error", "[%d%d] has number: %d\n", c.row, c.col, b[c.row][c.col].Number)
		return
	}

	if t > 0 {
		// recursive call, previous try failed, reset to 0
		seq = []int{}
		for i := c.row; i < c.row+3; i++ {
			for j := c.col; j < c.col+3; j++ {
				b[i][j].Number = 0
			}
		}
	}

	// out:
	for i := c.row; i < c.row+3; i++ {
		col:
		for j := c.col; j < c.col+3; j++ {
			c := b[i][j]
			used := b.findUsed(c)
			free := b.findFree(c, used)
			ilog("info", "free numbers for [%d%d]: %v\n", i, j, free)
			if len(free) == 0 {
				t++
				ilog("info", "re-start\n")
				ilog("info", "seq: %v, try #%d", seq, t)
				// TODO:
				// return b.fillBox(c, t, seq)
			}

			if t > len(free)-1 && i == c.row && j == c.col {
				ilog("info", "no more tries for [%d%d]\n", i, j)
				b.setValue(i, j, free[0])
				seq = append(seq, free[0])
				// t = 2
				break col
			}

			var trynew int
			if t > 0 {
				trynew = t
			}
			if t > len(free)-1 {
				trynew = len(free) - 1
			}

			b.setValue(i, j, free[trynew])
			seq = append(seq, free[trynew])
		}
	}

	ilog("info", "seq: %v, try #%d", seq, t)
	return
}

// add number in a list, no duplicates
func addOnce(listn []int, n int) []int {
	for i := 0; i < len(listn); i++ {
		if listn[i] == n {
			return listn
		}
	}

	return append(listn, n)
}

// find used numbers (not available) for a cell
func (b *Board) findUsed(c Cell) []int {
	var used []int
	if b[c.row][c.col].Number > 0 {
		// ilog("info", "c [%d%d] not empty", c.row, c.col)
		return used
	}
	if c.row > 9 || c.row < 0 {
		ilog("error", "not a valid c: [%d%d]", c.row, c.col)
		return used
	}
	if c.col > 9 || c.col < 0 {
		ilog("error", "not a valid c: [%d%d]", c.row, c.col)
		return used
	}

	// check row first
	for i := 0; i < 9; i++ {
		if b[c.row][i].Number > 0 {
			used = addOnce(used, b[c.row][i].Number)
		}
	}

	// check column
	for i := 0; i < 9; i++ {
		if b[i][c.col].Number > 0 {
			used = addOnce(used, b[i][c.col].Number)
		}
	}

	// finally check box
	brow, bcol := box(c.row, c.col)
	for i := brow; i < brow+3; i++ {
		for j := bcol; j < bcol+3; j++ {
			if b[i][j].Number > 0 {
				used = addOnce(used, b[i][j].Number)
			}
		}
	}

	return used
}

// find free (available) numbers for a cell
func (b *Board) findFree(c Cell, used []int) []int {
	var free []int

	for i := 1; i < 10; i++ {
	out:
		for j := 0; j < len(used); j++ {
			// ilog("debug", "checking %d against used %d\n", i, used[j])
			if i == used[j] {
				// ilog("debug", "%d is equal to used %d, check next\n", i, used[j])
				break out
			}
			if j == len(used)-1 {
				// ilog("debug", "reached end of used numbers, add this %d to free\n", i)
				free = append(free, i)
				continue
			}
		}
	}
	return free
}

// check row for a number
func (b *Board) checkRow(num int, row int) interface{} {
	for col := 0; col < 9; col++ {
		if b[row][col].Number == num {
			return fmt.Sprintf("number %d found in cell [%d%d]", num, row, col)
		}
	}
	return nil
}

// check column for a number
func (b *Board) checkCol(n int, col int) interface{} {
	for row := 0; row < 9; row++ {
		if b[row][col].Number == n {
			return fmt.Sprintf("number %d found in cell [%d%d]", n, row, col)
		}
	}
	return nil
}

// box returns the 3x3 range (box) that a cell belongs in
// i.e the upper left cell's row and column
func box(row, col int) (int, int) {
	crd3 := row / 3
	ccd3 := col / 3

	srow := crd3 * 3
	scol := ccd3 * 3

	return srow, scol
}

// check box for a number
func (b *Board) checkBox(num int, row int, col int) interface{} {
	srow, scol := box(row, col)
	for row := srow; row < srow+3; row++ {
		for col := scol; col < scol+3; col++ {
			if b[row][col].Number == num {
				return fmt.Sprintf("number %d found in cell [%d%d]", num, row, col)
			}
		}
	}
	return nil
}

// check number on a cell
func(b *Board) checkNum(n int, r int, c int) interface{} {
	if found := b.checkRow(n, r); found != nil {
		return found
	}
	if found := b.checkCol(n, c); found != nil {
		return found
	}
	if found := b.checkBox(n, r, c); found != nil {
		return found
	}

	return nil
}

// add mark number for a cell
func (b *Board) addMark(row, col, n int) {
	for i := 0; i < len(b[row][col].marks); i++ {
		if b[row][col].marks[i] == n {
			return
		}
	}
	b[row][col].marks = addOnce(b[row][col].marks, n)
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
	if c.selected {
		color = cBgBlue
	}

	return "\033[0;" + color + number + "\033[0m"
}

// select a row
func (b *Board) selectRow(row int) {
	for i := 0; i < 9; i++ {
		b[row][i].selected = true
	}
}

// select a column
func (b *Board) selectColumn(column int) {
	for i := 0; i < 9; i++ {
		b[i][column].selected = true
	}
}

// select a 3x3 box
func (b *Board) selectBox(row int, col int) {
	brow, bcol := box(row, col)
	for i := brow; i < brow+3; i++ {
		for j := bcol; j < bcol+3; j++ {
			b[i][j].selected = true
		}
	}
}

// select row, columns and 3x3 box given a cell
func (b *Board) selectCells(row int, col int) {
	b.selectRow(row)
	b.selectColumn(col)
	b.selectBox(row, col)
}

// func printCell(c Cell) {
// 	ilog("debug", "printing cell: %#v\n", c)

// 	switch c.row % 3 {
// 	case 0:
// 		if c.col%3 == 0 {
// 			fmt.Printf("\033[0;%s\u250F\u2501\u2501\u2501\033[0m", c.color)
// 			fmt.Printf("\033[0;%s\u2503 %d \033[0m", c.color, c.Number)
// 		}
// 		if c.col%3 == 1 || c.col%3 == 2 {
// 			fmt.Printf("\033[0;%s\u252F\u2501\u2501\u2501\033[0m", c.color)
// 			fmt.Printf("\033[0;%s\u2502 %d \033[0m", c.color, c.Number)
// 		}
// 	case 1:
// 		if c.col%3 == 0 {
// 			fmt.Printf("\033[0;%s\u2520\u2500\u2500\u2500\033[0m", c.color)
// 			fmt.Printf("\033[0;%s\u2503 %d \033[0m", c.color, c.Number)
// 		}
// 		if c.col%3 == 1 || c.col%3 == 2 {
// 			fmt.Printf("\033[0;%s\u253C\u2501\u2501\u2501\033[0m", c.color)
// 			fmt.Printf("\033[0;%s\u2502 %d \033[0m", c.color, c.Number)
// 		}
// 	case 2:
// 		if c.col%3 == 0 {
// 			fmt.Printf("\033[0;%s\u2520\u2500\u2500\u2500\033[0m", c.color)
// 			fmt.Printf("\033[0;%s\u2503 %d \033[0m", c.color, c.Number)
// 		}
// 		if c.col%3 == 1 || c.col%3 == 2 {
// 			fmt.Printf("\033[0;%s\u253C\u2501\u2501\u2501\033[0m", c.color)
// 			fmt.Printf("\033[0;%s\u2502 %d \033[0m", c.color, c.Number)
// 		}
// 	}
// }

// prints the board with the cells contents if num not zero
func (b *Board) print() {
	fmt.Printf("\n\n")
	// START first row of boxes
	fmt.Printf("\t  \033[0;2m" + "  0   1   2   3   4   5   6   7   8\n" + "\033[0m")
	fmt.Printf("\t  \u250F\u2501\u2501\u2501\u252F\u2501\u2501\u2501\u252F\u2501\u2501\u2501\u2533\u2501\u2501\u2501\u252F\u2501\u2501\u2501\u252F\u2501\u2501\u2501\u2533\u2501\u2501\u2501\u252F\u2501\u2501\u2501\u252F\u2501\u2501\u2501\u2513\n")
	// first row of numbers
	b.printRow(0)
	fmt.Printf("\t  \u2520\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2542\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2542\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2528\n")
	// second row of numbers
	b.printRow(1)
	fmt.Printf("\t  \u2520\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2542\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2542\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2528\n")
	// third row of numbers
	b.printRow(2)
	fmt.Printf("\t  \u2523\u2501\u2501\u2501\u253F\u2501\u2501\u2501\u253F\u2501\u2501\u2501\u254B\u2501\u2501\u2501\u253F\u2501\u2501\u2501\u253F\u2501\u2501\u2501\u254B\u2501\u2501\u2501\u253F\u2501\u2501\u2501\u253F\u2501\u2501\u2501\u252B\n")
	// END first row of boxes

	// REPEAT
	// START second row of boxes (no border)
	// first row of numbers
	b.printRow(3)
	fmt.Printf("\t  \u2520\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2542\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2542\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2528\n")
	// second row of numbers
	b.printRow(4)
	fmt.Printf("\t  \u2520\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2542\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2542\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2528\n")
	// third row of numbers
	b.printRow(5)
	fmt.Printf("\t  \u2523\u2501\u2501\u2501\u253F\u2501\u2501\u2501\u253F\u2501\u2501\u2501\u254B\u2501\u2501\u2501\u253F\u2501\u2501\u2501\u253F\u2501\u2501\u2501\u254B\u2501\u2501\u2501\u253F\u2501\u2501\u2501\u253F\u2501\u2501\u2501\u252B\n")
	// END second row of boxes
	// REPEAT
	// START third row of boxes (no border)

	// first row of numbers
	b.printRow(6)
	fmt.Printf("\t  \u2520\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2542\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2542\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2528\n")
	// second row of numbers
	b.printRow(7)
	fmt.Printf("\t  \u2520\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2542\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2542\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u253C\u2500\u2500\u2500\u2528\n")
	// third row of numbers
	b.printRow(8)
	fmt.Printf("\t  \u2517\u2501\u2501\u2501\u2537\u2501\u2501\u2501\u2537\u2501\u2501\u2501\u253B\u2501\u2501\u2501\u2537\u2501\u2501\u2501\u2537\u2501\u2501\u2501\u253B\u2501\u2501\u2501\u2537\u2501\u2501\u2501\u2537\u2501\u2501\u2501\u251B\n")
	// END third row of boxes

	fmt.Printf("\n\n")
}

// print each row between cell borders separately
// so the cell's numbers are printed (with color)
// replace 2502 with 250A or 2506 for vertical lines
func (b *Board) printRow(row int) {
	// print row number row in gray color
	fmt.Printf("\t\033[0;2m%d\033[0m ", row)

	// this one line printed; broken into three for better readability
	fmt.Printf("\u2503 %s \u2502 %s \u2502 %s \u2503", b[row][0].Content(), b[row][1].Content(), b[row][2].Content())
	fmt.Printf(" %s \u2502 %s \u2502 %s \u2503", b[row][3].Content(), b[row][4].Content(), b[row][5].Content())
	fmt.Printf(" %s \u2502 %s \u2502 %s \u2503\n", b[row][6].Content(), b[row][7].Content(), b[row][8].Content())
}

// mapValues makes a map of numbers in cells
func (b *Board) mapValues() map[int][]Cell {
	m := make(map[int][]Cell)

	// iterate over all cells
	for row := 0; row < 9; row++ {
		for col := 0; col < 9; col++ {
			// current cell
			c := b[row][col]

			// initial number in this cell
			number := c.Number

			// ignore 0s
			if number == 0 {
				continue
			}

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

// print mapped values
func printMaps(vmap map[int][]Cell) {
	for i := 1; i < 10; i++ {
		fmt.Printf("\tnumber %d found in:", i)
		for _, v := range vmap[i] {
			fmt.Printf(" [%d%d]", v.row, v.col)
		}
		fmt.Print("\n")
	}
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
func (b *Board) clear() {
	for row := 0; row < 9; row++ {
		for col := 0; col < 9; col++ {
			b[row][col].invalid = false
			b[row][col].active = false
			b[row][col].selected = false
			b[row][col].candid = false
			b[row][col].solved = false
			b[row][col].blink = false
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

func (b *Board) play() {
	b.print()
	for {
		num := getNumber()
		ilog("debug", "got %d", num)
		err := b.save()
		if err != nil {
			ilog("error", "error saving: %s", err)
		}

		// TODO
		// check number
		// check in row
		// check in column
		// check in 9-cell box
		// cross hatch
		// set board's state
		b.print()
	}
}

func main() {
	debug = true
	b := board()

	// main loop
	fmt.Printf("\tOptions: (n)ew, (r)esume, e(x)it\n")
	input := getInput()
	for {
		switch input {
		case "n":
			// load puzzle from puzzle.json file
			err := b.load(puzzleFile)
			if err != nil {
				panic(err)
			}
			b.play()
			// // make a map of existing numbers in cells
			// mapv := b.mapValues()
			// ilog("info", "\tNew puzzle, difficulty: %s\n", difficulty(mapv))
			// b.print()
			// input = getInput()

		case "r":
			// load previously saved puzzle in state.json
			err := b.load(stateFile)
			if err != nil {
				panic(err)
			}
			b.play()

		case "x":
			return // exit program

		default:
			fmt.Printf("\tOptions: (n)ew, (r)esume, e(x)it\n")
			input = getInput()

		}
	}
}
