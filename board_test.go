package main

import (
	"testing"
	// "fmt"
)

func TestCheckRow(t *testing.T) {
	// numbers to check against the current state of the board
	var tests = []struct {
		num int
		row int
		want string
	} {
		{1, 0, "number 1 found in cell[02]"},
		{5, 0, "number 5 found in cell[00]"},
		{4, 0, "number 4 not found in 0 row"},
		{4, 8, "number 4 not found in 8 row"},
		{3, 6, "number 3 not found in 6 row"},
		{5, 6, "number 5 not found in 6 row"},
	}

	var board [][]Cell
	if board, err := load(puzzleFile); err != nil {
		t.Errorf("error loading puzzle file: %s", err)
	}

	for _, test := range tests {
		got := checkRow(board, test.num, test.row)
		if got != test.want {
			t.Errorf("CheckRow(%d, %d) = %v; want %v", test.num, test.row, got, test.want)
		}
	}
}

func TestCheckCol(t *testing.T) {
	// numbers to check against the current state of the board
	var tests = []struct {
		num int
		col int
		want string
	} {
		{1, 0, "number 1 found in cell[70]"},
		{5, 0, "number 5 found in cell[00]"},
		{4, 0, "number 4 not found in 0 column"},
		{4, 8, "number 4 not found in cell[28]"},
		{3, 6, "number 3 not found in cell[46]"},
		{5, 6, "number 5 not found in 6 column"},
	}

	var board [][]Cell
	if board, err := load(puzzleFile); err != nil {
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
		num int
		row int
		col int
		want string
	} {
		{1, 0, 1, "number 1 found in cell[02]"},
		{5, 0, 2, "number 5 found in cell[00]"},
		{4, 3, 0, "number 4 not found in cell range [30]-[52]"},
		{4, 3, 8, "number 4 found in cell[57]"},
		{3, 3, 6, "number 3 found in cell[46]"},
		{5, 3, 6, "number 5 not found in cell range [36]-[58]"},
	}

	var board [][]Cell
	if board, err := load(puzzleFile); err != nil {
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
	var board [][]Cell
	if board, err := load(puzzleFile); err != nil {
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
	var board [][]Cell
	if board, err := load(puzzleFile); err != nil {
		t.Errorf("error loading state file: %s", err)
	}

	board[0][0].Number = 15

	// save puzzle state
	if err := save(board); err != nil {
		t.Logf("error saving puzzle: %s", err)
	}

	if board, err := load(stateFile); err != nil {
		t.Errorf("error loading state file: %s", err)
	}

	if board[0][0].Number != 15 {
		t.Error("the number was not saved")
	}
}

func TestDifficulty(t *testing.T) {
	var board [][]Cell
	if board, err := load(stateFile); err != nil {
		t.Logf("error loading state file: %s", err)
	}


}