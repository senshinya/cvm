package common

import (
	"fmt"
	"strconv"
	"strings"
)

func IsLetter_(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || b == '_'
}

func IsDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

func IsOctDigit(b byte) bool {
	return b >= '0' && b <= '7'
}

func IsHexDigit(b byte) bool {
	return IsDigit(b) || b == 'A' || b == 'a' || b == 'B' || b == 'b' ||
		b == 'C' || b == 'c' || b == 'D' || b == 'd' || b == 'E' || b == 'e' || b == 'F' || b == 'f'
}

var charToAscii = map[byte]int64{
	'0': 0,
	'1': 1,
	'2': 2,
	'3': 3,
	'4': 4,
	'5': 5,
	'6': 6,
	'7': 7,
	'8': 8,
	'9': 9,
	'a': 10,
	'A': 10,
	'b': 11,
	'B': 11,
	'c': 12,
	'C': 12,
	'd': 13,
	'D': 13,
	'e': 14,
	'E': 14,
	'f': 15,
	'F': 15,
}

func HexOrOctCharToNum(char byte) int64 {
	return charToAscii[char]
}

func CheckAndUnquoteCharacterInString(bytes string) (byte, error) {
	if !strings.HasPrefix(bytes, "\\") {
		return bytes[0], nil
	}
	if IsSimpleEscapeSuffix(bytes[1]) {
		unquote, err := strconv.Unquote("\"" + bytes + "\"")
		if err != nil {
			return 0, err
		}
		return unquote[0], nil
	}
	var ascii int64
	if strings.HasPrefix(bytes, "\\x") {
		// hex escape
		for i := 2; i < len(bytes); i++ {
			ascii *= 16
			ascii += HexOrOctCharToNum(bytes[i])
		}
	} else {
		// oct escape
		for i := 1; i < len(bytes); i++ {
			ascii *= 8
			ascii += HexOrOctCharToNum(bytes[i])
		}
	}
	if ascii > 255 || ascii < 0 {
		return 0, fmt.Errorf("%s out of range", bytes)
	}
	return byte(ascii), nil
}

func CheckAndUnquoteCharacterLiteral(bytes string) (byte, error) {
	if !strings.HasPrefix(bytes, "\\") {
		return bytes[0], nil
	}
	if IsSimpleEscapeSuffix(bytes[1]) {
		if bytes[1] == '\'' {
			bytes = "'"
		}
		unquote, err := strconv.Unquote("\"" + bytes + "\"")
		if err != nil {
			return 0, err
		}
		return unquote[0], nil
	}
	var ascii int64
	if strings.HasPrefix(bytes, "\\x") {
		// hex escape
		for i := 2; i < len(bytes); i++ {
			ascii *= 16
			ascii += HexOrOctCharToNum(bytes[i])
		}
	} else {
		// oct escape
		for i := 1; i < len(bytes); i++ {
			ascii *= 8
			ascii += HexOrOctCharToNum(bytes[i])
		}
	}
	if ascii > 255 || ascii < 0 {
		return 0, fmt.Errorf("%s out of range", bytes)
	}
	return byte(ascii), nil
}

func IsSimpleEscapeSuffix(b byte) bool {
	suffixMap := map[byte]struct{}{
		'\'': {}, '"': {}, '?': {}, '\\': {}, 'a': {}, 'b': {}, 'f': {}, 'n': {}, 'r': {}, 't': {}, 'v': {},
	}
	_, ok := suffixMap[b]
	return ok
}
