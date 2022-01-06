package main

import (
	"fmt"
	"testing"
)

func TestCheck

func TestMapValues(t *testing.T) {
	if err := load(puzzleFile); err != nil {
		t.Errorf("error loading puzzle file: %s", err)
	}
	if err := mapValues(); err != nil {
		t.Logf("error mapping values: %s", err)
	}

	// set a cell's number; replace the first 0 found with an invalid number
	// in this case it's cell 03s
	c := ValuesMap[0][0]
	c.Number = 10

	if err := mapValues(); err != nil {
		if fmt.Sprintf("%s", err) == "error mapping values: invalid number 10 in cell 03" {
			t.Logf("error mapping values: %s", err)
		}
	}
}

func TestLoad(t *testing.T) {
	if err := load(puzzleFile); err != nil {
		t.Errorf("error loading state file: %s", err)
	}
}

func TestSave(t *testing.T) {
	if err := load(puzzleFile); err != nil {
		t.Errorf("error loading state file: %s", err)
	}

	Board[0][0].Number = 15

	// save puzzle state
	if err := save(); err != nil {
		t.Logf("error saving puzzle: %s", err)
	}

	if err := load(stateFile); err != nil {
		t.Errorf("error loading state file: %s", err)
	}

	if Board[0][0].Number != 15 {
		t.Error("the number was not saved")
	}
}

func TestDifficulty(t *testing.T) {
	if err := load(stateFile); err != nil {
		t.Logf("error loading state file: %s", err)
	}

}
