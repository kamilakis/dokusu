package main

import (
	// "fmt"
	"testing"
)

func TestPrintCell(t *testing.T) {
	debug = false
	b := Board{}
	// for i := 0; i < 9; i++ {
	// 	t.Logf("%d mod 3 equals: %#+v", i, i%3)
	// }
	b.gen3boxes()
	b.selectCells(5, 8)
	b.print()
}

func TestRandRow(t *testing.T) {
	b := Board{}
	ints := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	ints = shuffle(ints)

	// populate a row
	for col := 0; col < 9; col++ {
		b[0][col].Number = ints[col]
	}
	b.print()

	// check row
	for col := 0; col < 9; col++ {
		got := b[0][col].Number
		if got > 9 || got < 1 { // this never happens
			t.Fail()
			t.Logf("not a valid number: %d", got)
			b[0][col].invalid = true
		}
	}
	b.print()
}

func TestRandColumn(t *testing.T) {
	b := Board{}
	ints := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	ints = shuffle(ints)

	// populate column 5
	column := 5
	for row := 0; row < 9; row++ {
		b[row][column].Number = ints[row]
	}
	b.print()
}

func TestGenBox(t *testing.T) {
	b := Board{}
	ints := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	ints = shuffle(ints)

	// populate first box
	cell := Cell{row: 0, col: 0}
	b.genBox(cell)
	// populate second box
	cell = Cell{row: 3, col: 3}
	b.genBox(cell)
	// populate third box
	cell = Cell{row: 6, col: 6}
	b.genBox(cell)

	b.print()
}

func TestFillBox(t *testing.T) {
	debug = true
	b := Board{}
	b.gen3boxes()
	// b.print()

	// try filling second starting from cell [03]
	c := Cell{row: 0, col: 3}
	b.fillBox(c, 0, []int{})
	b.print()

	if b[c.row+2][c.col+2].Number > 0 {
		c := Cell{row: 0, col: 6}
		b.fillBox(c, 0, []int{})
		b.print()
	}

	if b[c.row+2][c.col+2].Number > 0 {
		c := Cell{row: 3, col: 6}
		b.fillBox(c, 0, []int{})
		b.print()
	}

	if b[c.row+2][c.col+2].Number > 0 {
		c := Cell{row: 3, col: 0}
		b.fillBox(c, 0, []int{})
		b.print()
	}
}

func TestComplete(t *testing.T) {
	debug = true
	b := Board{}
	b.gen3boxes()

	// try all numbers from 1 to 9
	// mark cells with possible values
	for row := 0; row < 9; row++ {
		for col := 0; col < 9; col++ {
			for n := 1; n < 10; n++ {
				if  b.checkNum(n, row, col) == nil {
					// b[row][col].selected = true
					b[row][col].marks = addOnce(b[row][col].marks, n)
				}
			}
		}
	}

	// show marks
	for row := 0; row < 9; row++ {
		for col := 0; col < 9; col++ {
			t.Logf("marks for [%d%d]: %v", row, col, b[row][col].marks)
		}
	}

	// vmap := mapValues(board)
	// printMaps(vmap)

	// having mark numbers for all empty cells
	// try all possible values for any cell, i.e. brute force
	for row := 0; row < 9; row++ {
		for col := 0; col < 9; col++ {
			if len(b[row][col].marks) > 0 {
				c := b[row][col]
				t.Logf("marks for [%d%d]: %v", row, col, c.marks)
				if check := b.checkNum(c.marks[0], row, col); check != nil {
					t.Logf("cannot set %d for [%d%d]: %v", c.marks[0], row, col, check)
					b.setValue(row, col, c.marks[1])
				} else  {
					b.setValue(row, col, c.marks[0])
				}
			}
		}
	}

	b.print()
}

func TestGenCell(t *testing.T) {
	b := Board{}
	b.gen3boxes()
	b.print()
	// proceed one cell at a time
	for row := 0; row < 9; row++ {
		for col := 0; col < 9; col++ {
			c := b[row][col]
			if b[row][col].Number == 0 {
				used := b.findUsed(c)
				free := b.findFree(c, used)
				t.Logf("free numbers for [%d%d]: %#v", row, col, free)
				// set cell's number with the first free
				b[row][col].Number = free[0]
				b.print()
				t.Logf("set number %d for [%d%d]\n", free[0], row, col)
				if col == 8 {
					t.Logf("---------- row %d complete\n", row)
				}
				continue
			}
			t.Logf("number %d already set in [%d%d]\n", b[row][col].Number, row, col)
			if col == 8 {
				t.Logf("+++ column %d complete\n", col)
			}
		}
		if row == 8 {
			t.Logf("---------- row %d complete\n", row)
		}
	}
	b.print()
}

func TestFindUsed(t *testing.T) {
	b := Board{}
	b.gen3boxes()

	// find used numbers (not available) for this cell
	c := b[5][2]
	used := b.findUsed(c)
	b.print()
	t.Logf("used numbers (not available) for [%d%d], %#+v", c.row, c.col, used)
}

func TestFindFree(t *testing.T) {
	b := Board{}
	b.gen3boxes()

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
		c := b[test.row][test.col]
		used := b.findUsed(c)
		free := b.findFree(c, used)
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
	b := Board{}
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

	err := b.load(puzzleFile)
	if err != nil {
		t.Errorf("error loading puzzle file: %s", err)
	}

	for _, test := range tests {
		got := b.checkRow(test.num, test.row)
		if got != test.want {
			t.Errorf("CheckRow(%d, %d) = %#+v; want %#+v", test.num, test.row, got, test.want)
		}
	}
}

func TestCheckCol(t *testing.T) {
	b := Board{}
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

	err := b.load(puzzleFile)
	if err != nil {
		t.Errorf("error loading puzzle file: %s", err)
	}

	for _, test := range tests {
		got := b.checkCol(test.num, test.col)
		if got != test.want {
			t.Errorf("checkCol(%d, %d) = %v; want %v", test.num, test.col, got, test.want)
		}
	}
}

func TestCheckBox(t *testing.T) {
	b := Board{}
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

	err := b.load(puzzleFile)
	if err != nil {
		t.Errorf("error loading puzzle file: %s", err)
	}

	for _, test := range tests {
		got := b.checkBox(test.num, test.row, test.col)
		if got != test.want {
			t.Errorf("checkBox(%d, %d, %d) = %v; want %v", test.num, test.row, test.col, got, test.want)
		}
	}
}

func TestMapValues(t *testing.T) {
	b := Board{}
	var tests = []struct {
		row int
		col int
		set int
	}{
		{5, 3, 7},
		{0, 8, 5},
		{4, 4, 3},
		{2, 5, 2},
		{1, 0, 6},
		{5, 3, 8},
		{5, 1, 1},
		{8, 1, 4},
		{6, 3, 2},
	}
	for _, test := range tests {
		t.Logf("=== start test %#v", test)
		// set cell's number
		b[test.row][test.col].Number = test.set
		// make new values map of board
		vmap := b.mapValues()
		t.Logf("values map for %d: %#+v\n", test.set, vmap[test.set])
		// expect to find the cell set on the vmap
		for i := 0; i < len(vmap[test.set]); i++ {
			if vmap[test.set][i].Number == test.set &&
				vmap[test.set][i].row == test.row &&
				vmap[test.set][i].col == test.col {
				t.Logf("number %d for [%d%d] found in %#v\n", test.set, test.row, test.col, vmap[test.set][i])
				break
			}
			// reached the end of the vmap for set number
			if i == len(vmap[test.set])-1 {
				t.Errorf("%d not found for cell [%d%d]", test.set, test.row, test.col)
			}
		}
		t.Logf("--- end test %#v", test)
	}
}

func TestSaveState(t *testing.T) {
	b := Board{}
	err := b.load(puzzleFile)
	if err != nil {
		t.Errorf("error loading state file: %s", err)
	}

	b[0][0].Number = 15

	// save puzzle state
	if err := b.save(); err != nil {
		t.Logf("error saving puzzle: %s", err)
	}

	err = b.load(stateFile)
	if err != nil {
		t.Errorf("error loading state file: %s", err)
	}

	if b[0][0].Number != 15 {
		t.Error("the number was not saved")
	}
}
