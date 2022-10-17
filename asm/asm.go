package asm

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

type MismatchError struct {
	Expect string
	Actual string
}

func (e *MismatchError) Error() string {
	return fmt.Sprintf("expect a %s, but found '%s'", e.Expect, e.Actual)
}

var regRegister = regexp.MustCompile("^(eax|ecx|edx|ebx|esi|edi|esp|ebp)$")
var regNumber = regexp.MustCompile(`^(0x|0b|0o)?[a-fA-F0-9]+$`)
var regLabel = regexp.MustCompile(`^[_a-zA-Z][_a-zA-Z0-9]*$`)

// Register name to code map.
var registerCode = map[string]byte{
	"eax":    0,
	"ecx":    1,
	"edx":    2,
	"ebx":    3,
	"esp":    4,
	"ebp":    5,
	"esi":    6,
	"edi":    7,
	"no_reg": 8,
}

func Parse(src io.Reader) ([]byte, error) {

	var err error
	opCodes := make([]*opcode, 0)
	labelTable := make(labelTable)
	address := 0

	scanner := bufio.NewScanner(src)
	scanner.Split(splitFunc)

	for scanner.Scan() {
		var op *opcode

		// Find instruction, if exists, parse instruciton.
		if handleFunc, ok := instructionsMap[strings.ToLower(scanner.Text())]; ok {
			if op, err = handleFunc(address, labelTable, scanner); err != nil {
				return nil, err
			}
			opCodes = append(opCodes, op)
			address += op.Size
			continue
		}

		// Try define a label.
		if regLabel.Match([]byte(scanner.Text())) {
			label := scanner.Text()
			if !nextReadToken(scanner, ":") {
				return nil, &MismatchError{Expect: "label define", Actual: scanner.Text()}
			}
			addr := address
			labelTable[label] = &addr
			continue
		}

		return nil, fmt.Errorf("unexpect token '%s'", scanner.Text())
	}

	var code, result []byte
	for _, op := range opCodes {
		if code, err = op.EncodeFunc(op, labelTable); err != nil {
			return nil, err
		}
		result = append(result, code...)
	}

	return result, nil
}

func splitFunc(data []byte, atEOF bool) (advance int, token []byte, err error) {
	// Skip leading spaces.
	start := 0
	for start < len(data) {
		r, width := utf8.DecodeRune(data[start:])
		if !unicode.IsSpace(r) {
			break
		}
		start += width
	}

	// If end of stream, read more
	if start >= len(data) {
		return 0, nil, nil
	}

	// Lookup one char token like ',', '\n' etc.
	r, width := utf8.DecodeRune(data[start:])
	switch r {
	case ',', '%', '$', '(', ')', '-', ':':
		return start + width, data[start : start+width], nil
	}

	// Scan a word untill read a non-letter char or non-digit char or EOF.
	end := start
	for end < len(data) {
		r, width = utf8.DecodeRune(data[end:])
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			end += width
			continue
		}
		break
	}
	if end < len(data) {
		return end, data[start:end], nil
	}
	if atEOF {
		return end, data[start:], nil
	}

	// Read more char and retry.
	return 0, nil, nil
}

func nextReadToken(scanner *bufio.Scanner, token string) bool {
	return scanner.Scan() && scanner.Text() == token
}

func nextReadRegExp(scanner *bufio.Scanner, reg *regexp.Regexp) bool {
	return scanner.Scan() && reg.Match(scanner.Bytes())
}

func nextReadRegister(scanner *bufio.Scanner) (byte, error) {
	if !nextReadToken(scanner, "%") {
		return 0, fmt.Errorf("expect a register name, but found '%s'", scanner.Text())
	}
	if !nextReadRegExp(scanner, regRegister) {
		return 0, fmt.Errorf("expect a register name, but found '%s'", scanner.Text())
	}
	r := registerCode[scanner.Text()]
	return r, nil
}

func nextReadNumber(scanner *bufio.Scanner) (int, error) {
	if !scanner.Scan() {
		return 0, &MismatchError{Expect: "immediate number", Actual: scanner.Text()}
	}

	isNeg := false
	if scanner.Text() == "-" {
		isNeg = true
		if !scanner.Scan() {
			return 0, &MismatchError{Expect: "immediate number", Actual: scanner.Text()}
		}
	}

	if !regNumber.Match(scanner.Bytes()) {
		return 0, &MismatchError{Expect: "immediate number", Actual: scanner.Text()}
	}
	numStr := scanner.Text()
	if isNeg {
		numStr = "-" + numStr
	}
	num, err := strconv.ParseInt(numStr, 0, 32)
	if err != nil {
		return 0, &MismatchError{Expect: "immediate number", Actual: scanner.Text()}
	}
	return int(num), nil
}

func nextReadImmediateNumber(scanner *bufio.Scanner) (int, error) {
	if !nextReadToken(scanner, "$") {
		return 0, &MismatchError{Expect: "immediate number", Actual: scanner.Text()}
	}
	return nextReadNumber(scanner)
}
