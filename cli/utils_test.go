package cli

import (
	"testing"
)

func TestFindLastCommand(t *testing.T) {
	_, cmd := findLastCommand([]string{
		"A",
		"B",
		"\"CCCC CC\"",
		"D",
		"-aa",
		"-bb",
		"E",
	}, true)
	if cmd != "E" {
		t.Errorf("cmd is not E, it's %v", cmd)
		t.Fail()
	}
}

func TestFindLastCommand2(t *testing.T) {
	_, cmd := findLastCommand([]string{
		"A",
		"B",
		"\"CCCC CC\"",
		"D",
		"-aa",
		"-bb",
	}, true)
	if cmd != "D" {
		t.Errorf("cmd is not D, it's %v", cmd)
		t.Fail()
	}
}

func TestFindLastCommand3(t *testing.T) {
	_, cmd := findLastCommand([]string{
		"A",
	}, true)
	if cmd != "A" {
		t.Errorf("cmd is not A, it's %v", cmd)
		t.Fail()
	}
}

func TestFindLastCommand4(t *testing.T) {
	_, cmd := findLastCommand([]string{
		"A",
	}, false)
	if cmd != "A" {
		t.Errorf("cmd is not A, it's %v", cmd)
		t.Fail()
	}
}

func TestFindAllCommands(t *testing.T) {
	cmds := findAllCommands([]string{
		"A",
		"B",
		"-shit=\"CCCC CC\"",
		"D",
		"-aa",
		"-bb",
	})
	if cmds[0] != "A" ||
	cmds[1] != "B" ||
	cmds[2] != "D" {
		t.Errorf("boom: %v", cmds)
		t.Fail()
	}
}
