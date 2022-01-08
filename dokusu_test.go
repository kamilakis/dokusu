package main

import (
	// "fmt"
	"testing"
)

func TestRandRow(t *testing.T) {
	var board [9][9]Cell
	ints := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	ints = shuffle(ints)

	// populate a row
	for col := 0; col < 9; col++ {
		board[0][col].Number = ints[col]
	}
	print(board)

	// check row
	for col := 0; col < 9; col++ {
		got := board[0][col].Number
		if got > 9 || got < 1 { // this never happens
			t.Fail()
			t.Logf("not a valid number: %d", got)
			board[0][col].invalid = true
		}
	}
	print(board)
}

func TestRandColumn(t *testing.T) {
	var board [9][9]Cell
	ints := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	ints = shuffle(ints)

	// populate column 5
	column := 5
	for row := 0; row < 9; row++ {
		board[row][column].Number = ints[row]
	}
	print(board)
}

func TestGenBox(t *testing.T) {
	var board [9][9]Cell
	ints := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	ints = shuffle(ints)

	// populate first box
	cell := Cell{row: 0, col: 0}
	board = genBox(board, cell)
	// populate second box
	cell = Cell{row: 3, col: 3}
	board = genBox(board, cell)
	// populate third box
	cell = Cell{row: 6, col: 6}
	board = genBox(board, cell)

	print(board)
}

func TestFindUsed(t *testing.T) {
	var board [9][9]Cell
	ints := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	ints = shuffle(ints)

	// populate first box
	c := Cell{row: 0, col: 0}
	board = genBox(board, c)
	// populate second box
	c = Cell{row: 3, col: 3}
	board = genBox(board, c)
	// populate third box
	c = Cell{row: 6, col: 6}
	board = genBox(board, c)

	// find used numbers (not available) for this cell
	c = Cell{row: 5, col: 2}
	used := findUsed(board, c)
	print(board)
	t.Logf("used numbers (not available) for [%d%d], %#+v", c.row, c.col, used)

	// for i := 0; i < len(used); i++ {
	// 	for j := 0; j < len(used); j++ {
	// 		if
	// 	}
	// }
}

func TestFindFree(t *testing.T) {
	var board [9][9]Cell
	ints := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	ints = shuffle(ints)

	// populate first box
	c := Cell{row: 0, col: 0}
	board = genBox(board, c)
	// populate second box
	c = Cell{row: 3, col: 3}
	board = genBox(board, c)
	// populate third box
	c = Cell{row: 6, col: 6}
	board = genBox(board, c)

	// cells to check for free (available) numbers
	var tests = []struct {
		row  int
		col  int
		want int
	}{
		{0, 0, 0},
		{1, 0, 0},
		{2, 0, 0},
		{0, 1, 0},
		{1, 2, 0},
		{2, 3, 9},
		{3, 3, 0},
		{3, 4, 0},
		{3, 5, 0},
		{3, 6, 9},
		{4, 7, 9},
		{5, 0, 9},
		{6, 1, 9},
		{7, 1, 9},
		{8, 1, 9},
		{8, 7, 0},
		{8, 8, 0},
		{5, 5, 0},
		{8, 6, 0},
	}

	for _, test := range tests {
		t.Logf("testing cell [%d%d]", test.row, test.col)
		used := findUsed(board, Cell{row: test.row, col: test.col})
		free := findFree(board, Cell{row: test.row, col: test.col}, used)
		if len(used)+len(free) != test.want {
			t.Errorf("used+free equals to %d", len(used)+len(free))
		}
		for i := 0; i < len(used); i++ {
			for j := 0; j < len(free); j++ {
				if used[i] == free[j] {
					t.Errorf("used %d found in free", used[i])
				}
			}
		}
	}
}

func TestCheckRow(t *testing.T) {
	// numbers to check against the current state of the board
	var tests = []struct {
		num  int
		row  int
		want interface{}
	}{
		{1, 0, "number 1 found in cell [02]"},
		{5, 0, "number 5 found in cell [00]"},
		{4, 0, nil},
		{4, 8, nil},
		{3, 6, nil},
		{5, 6, nil},
	}

	board, err := load(puzzleFile)
	if err != nil {
		t.Errorf("error loading puzzle file: %s", err)
	}

	for _, test := range tests {
		got := checkRow(board, test.num, test.row)
		if got != test.want {
			t.Errorf("CheckRow(%d, %d) = %#+v; want %#+v", test.num, test.row, got, test.want)
		}
	}
}

func TestCheckCol(t *testing.T) {
	// numbers to check against the current state of the board
	var tests = []struct {
		num  int
		col  int
		want interface{}
	}{
		{1, 0, "number 1 found in cell [70]"},
		{5, 0, "number 5 found in cell [00]"},
		{4, 0, nil},
		{4, 8, "number 4 found in cell [28]"},
		{3, 6, "number 3 found in cell [46]"},
		{5, 6, nil},
	}

	board, err := load(puzzleFile)
	if err != nil {
		t.Errorf("error loading puzzle file: %s", err)
	}

	for _, test := range tests {
		got := checkCol(board, test.num, test.col)
		if got != test.want {
			t.Errorf("checkCol(%d, %d) = %v; want %v", test.num, test.col, got, test.want)
		}
	}
}

func TestCheckBox(t *testing.T) {
	// numbers to check against the current state of the board
	var tests = []struct {
		num  int
		row  int
		col  int
		want interface{}
	}{
		{1, 0, 1, "number 1 found in cell [02]"},
		{5, 0, 2, "number 5 found in cell [00]"},
		{4, 3, 0, nil},
		{4, 3, 8, "number 4 found in cell [57]"},
		{3, 3, 6, "number 3 found in cell [46]"},
		{5, 3, 6, nil},
	}

	board, err := load(puzzleFile)
	if err != nil {
		t.Errorf("error loading puzzle file: %s", err)
	}

	for _, test := range tests {
		got := checkBox(board, test.num, test.row, test.col)
		if got != test.want {
			t.Errorf("checkBox(%d, %d, %d) = %v; want %v", test.num, test.row, test.col, got, test.want)
		}
	}
}

// TODO
func TestMapValues(t *testing.T) {
	board, err := load(puzzleFile)
	if err != nil {
		t.Errorf("error loading puzzle file: %s", err)
	}
	vmap := mapValues(board)

	// set a cell's number; replace the first 0 found with an invalid number
	c := vmap[0][0]
	c.Number = 10

	vmap = mapValues(board)
	// TODO: test here
}

func TestSaveState(t *testing.T) {
	board, err := load(puzzleFile)
	if err != nil {
		t.Errorf("error loading state file: %s", err)
	}

	board[0][0].Number = 15

	// save puzzle state
	if err := save(board); err != nil {
		t.Logf("error saving puzzle: %s", err)
	}

	board, err = load(stateFile)
	if err != nil {
		t.Errorf("error loading state file: %s", err)
	}

	if board[0][0].Number != 15 {
		t.Error("the number was not saved")
	}
}
