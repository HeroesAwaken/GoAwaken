package GameSpy_test

import (
	"reflect"
	"testing"

	"github.com/ReviveNetwork/GoRevive/GameSpy"
)

func TestShortHash(t *testing.T) {
	shortHash := GameSpy.ShortHash("ABC")
	if shortHash != "902fbdd2b1df" {
		t.Errorf("ShortHash was incorrect, got: %s, want: %s.", shortHash, "902fbdd2b1df")
	}

	shortHash = GameSpy.ShortHash("abc")
	if shortHash != "900150983cd2" {
		t.Errorf("ShortHash was incorrect, got: %s, want: %s.", shortHash, "900150983cd2")
	}
}

func TestProcessCommand(t *testing.T) {
	testMessage1 := "\\lc\\2\\sesskey\\1\\proof\\2\\userid\\3\\final\\"
	testResult1 := &GameSpy.Command{
		Query:   "lc",
		Message: map[string]string{},
	}
	testResult1.Message["__query"] = "lc"
	testResult1.Message["lc"] = "2"
	testResult1.Message["sesskey"] = "1"
	testResult1.Message["proof"] = "2"
	testResult1.Message["userid"] = "3"
	testResult1.Message["final"] = ""

	processCommand, _ := GameSpy.ProcessCommand(testMessage1)
	if !reflect.DeepEqual(processCommand, testResult1) {
		t.Errorf("ProcessCommand was incorrect, got: %v, want: %v.", processCommand, testResult1)
	}
}
