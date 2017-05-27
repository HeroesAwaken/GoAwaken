package GameSpy

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"strings"
)

type Command struct {
	Message map[string]string
	Query   string
}

// Hash returns the MD5 hash as a hex-string
func Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

// ShortHash returns a MD5 hash of "str" reduced to 12 chars.
func ShortHash(str string) string {
	hash := Hash(str)
	return hash[0:12]
}

func ProcessCommand(msg string) (*Command, error) {
	outCommand := new(Command)
	outCommand.Message = make(map[string]string)
	data := strings.Split(msg, "\\")

	// TODO:
	// Should maybe return an emtpy Command struct instead
	if len(data) < 1 {
		return nil, errors.New("Command message invalid")
	}

	// TODO:
	// Check if that makes any sense? Kinda just translated from the js-code
	//		if (data.length < 2) { return out; }
	if len(data) == 2 {
		outCommand.Message["__query"] = data[0]
		outCommand.Query = data[1]
		return outCommand, nil
	}

	outCommand.Query = data[1]
	outCommand.Message["__query"] = data[1]
	for i := 1; i < len(data)-1; i = i + 2 {
		outCommand.Message[strings.ToLower(data[i])] = data[i+1]
	}

	return outCommand, nil
}
