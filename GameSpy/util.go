package GameSpy

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"math/rand"
	"net"
	"strings"
	"time"

	log "github.com/ReviveNetwork/GoRevive/Log"
)

const gamespyLetters = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ]["
const (
	gamespyLettersIdxBits = 6                            // 6 bits to represent a letter index
	gamespyLettersIdxMask = 1<<gamespyLettersIdxBits - 1 // All 1-bits, as many as letterIdxBits
	gamespyLettersIdxMax  = 63 / gamespyLettersIdxBits   // # of letter indices fitting in 63 bits
)

// CrcLookup table "for some sort of hashing thing what was this"
var CrcLookup = []rune{
	0x0000, 0xC0C1, 0xC181, 0x0140, 0xC301, 0x03C0, 0x0280, 0xC241,
	0xC601, 0x06C0, 0x0780, 0xC741, 0x0500, 0xC5C1, 0xC481, 0x0440,
	0xCC01, 0x0CC0, 0x0D80, 0xCD41, 0x0F00, 0xCFC1, 0xCE81, 0x0E40,
	0x0A00, 0xCAC1, 0xCB81, 0x0B40, 0xC901, 0x09C0, 0x0880, 0xC841,
	0xD801, 0x18C0, 0x1980, 0xD941, 0x1B00, 0xDBC1, 0xDA81, 0x1A40,
	0x1E00, 0xDEC1, 0xDF81, 0x1F40, 0xDD01, 0x1DC0, 0x1C80, 0xDC41,
	0x1400, 0xD4C1, 0xD581, 0x1540, 0xD701, 0x17C0, 0x1680, 0xD641,
	0xD201, 0x12C0, 0x1380, 0xD341, 0x1100, 0xD1C1, 0xD081, 0x1040,
	0xF001, 0x30C0, 0x3180, 0xF141, 0x3300, 0xF3C1, 0xF281, 0x3240,
	0x3600, 0xF6C1, 0xF781, 0x3740, 0xF501, 0x35C0, 0x3480, 0xF441,
	0x3C00, 0xFCC1, 0xFD81, 0x3D40, 0xFF01, 0x3FC0, 0x3E80, 0xFE41,
	0xFA01, 0x3AC0, 0x3B80, 0xFB41, 0x3900, 0xF9C1, 0xF881, 0x3840,
	0x2800, 0xE8C1, 0xE981, 0x2940, 0xEB01, 0x2BC0, 0x2A80, 0xEA41,
	0xEE01, 0x2EC0, 0x2F80, 0xEF41, 0x2D00, 0xEDC1, 0xEC81, 0x2C40,
	0xE401, 0x24C0, 0x2580, 0xE541, 0x2700, 0xE7C1, 0xE681, 0x2640,
	0x2200, 0xE2C1, 0xE381, 0x2340, 0xE101, 0x21C0, 0x2080, 0xE041,
	0xA001, 0x60C0, 0x6180, 0xA141, 0x6300, 0xA3C1, 0xA281, 0x6240,
	0x6600, 0xA6C1, 0xA781, 0x6740, 0xA501, 0x65C0, 0x6480, 0xA441,
	0x6C00, 0xACC1, 0xAD81, 0x6D40, 0xAF01, 0x6FC0, 0x6E80, 0xAE41,
	0xAA01, 0x6AC0, 0x6B80, 0xAB41, 0x6900, 0xA9C1, 0xA881, 0x6840,
	0x7800, 0xB8C1, 0xB981, 0x7940, 0xBB01, 0x7BC0, 0x7A80, 0xBA41,
	0xBE01, 0x7EC0, 0x7F80, 0xBF41, 0x7D00, 0xBDC1, 0xBC81, 0x7C40,
	0xB401, 0x74C0, 0x7580, 0xB541, 0x7700, 0xB7C1, 0xB681, 0x7640,
	0x7200, 0xB2C1, 0xB381, 0x7340, 0xB101, 0x71C0, 0x7080, 0xB041,
	0x5000, 0x90C1, 0x9181, 0x5140, 0x9301, 0x53C0, 0x5280, 0x9241,
	0x9601, 0x56C0, 0x5780, 0x9741, 0x5500, 0x95C1, 0x9481, 0x5440,
	0x9C01, 0x5CC0, 0x5D80, 0x9D41, 0x5F00, 0x9FC1, 0x9E81, 0x5E40,
	0x5A00, 0x9AC1, 0x9B81, 0x5B40, 0x9901, 0x59C0, 0x5880, 0x9841,
	0x8801, 0x48C0, 0x4980, 0x8941, 0x4B00, 0x8BC1, 0x8A81, 0x4A40,
	0x4E00, 0x8EC1, 0x8F81, 0x4F40, 0x8D01, 0x4DC0, 0x4C80, 0x8C41,
	0x4400, 0x84C1, 0x8581, 0x4540, 0x8701, 0x47C0, 0x4680, 0x8641,
	0x8201, 0x42C0, 0x4380, 0x8341, 0x4100, 0x81C1, 0x8081, 0x4040,
}

var randSrc = rand.NewSource(time.Now().UnixNano())

// Command struct
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

// ProcessCommand turns gamespy's command string to the
// command struct
func ProcessCommand(msg string) (*Command, error) {
	outCommand := new(Command)
	outCommand.Message = make(map[string]string)
	data := strings.Split(msg, "\\")

	// TODO:
	// Should maybe return an emtpy Command struct instead
	if len(data) < 1 {
		log.Errorln("Command message invalid")
		return nil, errors.New("Command message invalid")
	}

	// TODO:
	// Check if that makes any sense? Kinda just translated from the js-code
	//		if (data.length < 2) { return out; }
	if len(data) == 1 {
		outCommand.Message["__query"] = data[0]
		outCommand.Query = data[0]
		return outCommand, nil
	}

	outCommand.Query = data[1]
	outCommand.Message["__query"] = data[1]
	for i := 1; i < len(data)-1; i = i + 2 {
		outCommand.Message[strings.ToLower(data[i])] = data[i+1]
	}

	return outCommand, nil
}

// DecodePassword decodes gamespy's base64 string used for passwords
// to a cleantext string
func DecodePassword(pass string) (string, error) {
	pass = strings.Replace(pass, "_", "=", -1)
	pass = strings.Replace(pass, "[", "+", -1)
	pass = strings.Replace(pass, "]", "/", -1)
	decodedPass, err := base64.StdEncoding.DecodeString(pass)
	return string(decodedPass), err
}

// BF2RandomUnsafe is a not thread-safe version of BF2Random
// For thread-safety you should use BF2Random with your own seed
func BF2RandomUnsafe(randomLen int) string {
	return BF2Random(randomLen, randSrc)
}

// BF2Random generates a random string with valid BF2 random chars
// https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-golang/31832326
func BF2Random(randomLen int, source rand.Source) string {
	b := make([]byte, randomLen)
	// A source.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := randomLen-1, source.Int63(), gamespyLettersIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = source.Int63(), gamespyLettersIdxMax
		}
		if idx := int(cache & gamespyLettersIdxMask); idx < len(gamespyLetters) {
			b[i] = gamespyLetters[idx]
			i--
		}
		cache >>= gamespyLettersIdxBits
		remain--
	}

	return string(b)
}

func ProcessFESL(data string) map[string]string {
	out := make(map[string]string)
	dataMap := strings.Split(data, "\n")

	for i := 0; i < len(dataMap); i += 1 {
		objectMap := strings.Split(data, "=")
		if len(objectMap) != 2 {
			continue
		}

		out[objectMap[0]] = objectMap[1]
	}

	return out
}

func SerializeFESL(data map[string]string) string {
	var out string
	for key, value := range data {
		out += key + "=" + value + "\n"
	}
	return out
}

func Inet_ntoa(ipnr int64) net.IP {
	var bytes [4]byte
	bytes[0] = byte(ipnr & 0xFF)
	bytes[1] = byte((ipnr >> 8) & 0xFF)
	bytes[2] = byte((ipnr >> 16) & 0xFF)
	bytes[3] = byte((ipnr >> 24) & 0xFF)

	return net.IPv4(bytes[0], bytes[1], bytes[2], bytes[3])
}
