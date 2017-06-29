package GameSpy_test

import (
	"reflect"
	"testing"

	"github.com/HeroesAwaken/GoAwaken/GameSpy"
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

func TestDecodePassword(t *testing.T) {
	decodePassword, err := GameSpy.DecodePassword("U3VwZXJEdXBlclNlY3JldFBhc3N3b3Jk")
	if err != nil {
		t.Errorf("TestDecodePassword was incorrect, got error: %s", err)
	}
	if decodePassword != "SuperDuperSecretPassword" {
		t.Errorf("TestDecodePassword was incorrect, got: %s, want: %s.", decodePassword, "SuperDuperSecretPassword")
	}
}

func TestBF2RandomUnsafe(t *testing.T) {
	rand1 := GameSpy.BF2RandomUnsafe(5)
	rand2 := GameSpy.BF2RandomUnsafe(5)
	if rand1 == rand2 {
		t.Errorf("TestBF2RandomUnsafe was incorrect, got same value twice: %s, %s.", rand1, rand2)
	}

	if len(rand1) != 5 || len(rand2) != 5 {
		t.Errorf("TestBF2RandomUnsafe was incorrect, got wrong lenght: %s, %s.", rand1, rand2)
	}

}
