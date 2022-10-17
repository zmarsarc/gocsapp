package asm

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"strings"
)

type labelTable map[string]*int

type encodeFunc func(op *opcode, labels labelTable) ([]byte, error)

type instFunc func(address int, labels labelTable, scanner *bufio.Scanner) (*opcode, error)

type opcode struct {
	Address    int
	Code       byte
	Size       int
	Ra, Rb     byte
	Var        int
	Label      string
	EncodeFunc encodeFunc
}

var instructionsMap = map[string]instFunc{
	"nop":    nop,
	"halt":   halt,
	"rrmovl": rrmovl,
	"irmovl": irmovl,
	"rmmvol": rmmvol,
	"mrmovl": mrmovl,
	"addl":   opl,
	"subl":   opl,
	"andl":   opl,
	"xorl":   opl,
	"jmp":    branch,
	"jle":    branch,
	"jl":     branch,
	"je":     branch,
	"jne":    branch,
	"jge":    branch,
	"jg":     branch,
	"call":   branch,
	"ret":    ret,
	"pushl":  stack,
	"popl":   stack,
}

func encodeSimpleOpCode(op *opcode, labels labelTable) ([]byte, error) {
	return []byte{op.Code}, nil
}

func encodeRegisterOpCode(op *opcode, labels labelTable) ([]byte, error) {
	return []byte{op.Code, (op.Ra << 4) | op.Rb}, nil
}

func encodeRegisterImmediateOpCode(op *opcode, labels labelTable) ([]byte, error) {
	result := append([]byte{}, op.Code, (op.Ra<<4)|op.Rb)
	numBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(numBytes, uint32(op.Var))
	result = append(result, numBytes...)
	return result, nil
}

func encodeBranchingOpCode(op *opcode, labels labelTable) ([]byte, error) {
	addr := labels[op.Label]
	if addr == nil {
		return nil, fmt.Errorf("unsolved reference: %s", op.Label)
	}

	numBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(numBytes, uint32(*addr))
	result := []byte{op.Code}
	result = append(result, numBytes...)
	return result, nil
}

func nop(address int, labels labelTable, scanner *bufio.Scanner) (*opcode, error) {
	return &opcode{Address: address, Code: 0x00, Size: 1, EncodeFunc: encodeSimpleOpCode}, nil
}

func halt(address int, labels labelTable, scanner *bufio.Scanner) (*opcode, error) {
	return &opcode{Address: address, Code: 0x10, Size: 1, EncodeFunc: encodeSimpleOpCode}, nil
}

func rrmovl(address int, labels labelTable, scanner *bufio.Scanner) (*opcode, error) {
	op := &opcode{Address: address, Code: 0x20, Size: 2, EncodeFunc: encodeRegisterOpCode}

	var err error
	if op.Ra, err = nextReadRegister(scanner); err != nil {
		return nil, err
	}
	if !nextReadToken(scanner, ",") {
		return nil, &MismatchError{Expect: "','", Actual: scanner.Text()}
	}
	if op.Rb, err = nextReadRegister(scanner); err != nil {
		return nil, err
	}

	return op, nil
}

func irmovl(addr int, labels labelTable, scanner *bufio.Scanner) (*opcode, error) {
	op := &opcode{
		Address:    addr,
		Code:       0x30,
		Ra:         registerCode["no_reg"],
		Size:       6,
		EncodeFunc: encodeRegisterImmediateOpCode,
	}

	var err error
	if op.Var, err = nextReadImmediateNumber(scanner); err != nil {
		return nil, err
	}
	if !nextReadToken(scanner, ",") {
		return nil, &MismatchError{Expect: "','", Actual: scanner.Text()}
	}
	if op.Rb, err = nextReadRegister(scanner); err != nil {
		return nil, err
	}

	return op, nil
}

func rmmvol(addr int, labels labelTable, scanner *bufio.Scanner) (*opcode, error) {
	op := &opcode{
		Address:    addr,
		Code:       0x40,
		Size:       6,
		EncodeFunc: encodeRegisterImmediateOpCode,
	}

	var err error
	if op.Ra, err = nextReadRegister(scanner); err != nil {
		return nil, err
	}
	if !nextReadToken(scanner, ",") {
		return nil, &MismatchError{Expect: "','", Actual: scanner.Text()}
	}
	if op.Var, err = nextReadNumber(scanner); err != nil {
		return nil, err
	}
	if !nextReadToken(scanner, "(") {
		return nil, &MismatchError{Expect: "'('", Actual: scanner.Text()}
	}
	if op.Rb, err = nextReadRegister(scanner); err != nil {
		return nil, err
	}
	if !nextReadToken(scanner, ")") {
		return nil, &MismatchError{Expect: "')'", Actual: scanner.Text()}
	}

	return op, nil
}

func mrmovl(addr int, labels labelTable, scanner *bufio.Scanner) (*opcode, error) {
	op := &opcode{
		Address:    addr,
		Code:       0x50,
		Size:       6,
		EncodeFunc: encodeRegisterImmediateOpCode,
	}

	var err error
	if op.Var, err = nextReadNumber(scanner); err != nil {
		return nil, err
	}
	if !nextReadToken(scanner, "(") {
		return nil, &MismatchError{Expect: "'('", Actual: scanner.Text()}
	}
	if op.Rb, err = nextReadRegister(scanner); err != nil {
		return nil, err
	}
	if !nextReadToken(scanner, ")") {
		return nil, &MismatchError{Expect: "')'", Actual: scanner.Text()}
	}
	if !nextReadToken(scanner, ",") {
		return nil, &MismatchError{Expect: "','", Actual: scanner.Text()}
	}
	if op.Ra, err = nextReadRegister(scanner); err != nil {
		return nil, err
	}

	return op, nil
}

func opl(addr int, labels labelTable, scanner *bufio.Scanner) (*opcode, error) {
	op := &opcode{
		Address:    addr,
		Size:       2,
		EncodeFunc: encodeRegisterOpCode,
	}

	switch strings.ToLower(scanner.Text()) {
	case "addl":
		op.Code = 0x60
	case "subl":
		op.Code = 0x61
	case "andl":
		op.Code = 0x62
	case "xorl":
		op.Code = 0x63
	}

	var err error
	if op.Ra, err = nextReadRegister(scanner); err != nil {
		return nil, err
	}
	if !nextReadToken(scanner, ",") {
		return nil, &MismatchError{Expect: "','", Actual: scanner.Text()}
	}
	if op.Rb, err = nextReadRegister(scanner); err != nil {
		return nil, err
	}

	return op, nil
}

func branch(addr int, labels labelTable, scanner *bufio.Scanner) (*opcode, error) {
	op := &opcode{
		Address:    addr,
		Size:       5,
		EncodeFunc: encodeBranchingOpCode,
	}

	switch strings.ToLower(scanner.Text()) {
	case "jmp":
		op.Code = 0x70
	case "jle":
		op.Code = 0x71
	case "jl":
		op.Code = 0x72
	case "je":
		op.Code = 0x73
	case "jne":
		op.Code = 0x74
	case "jge":
		op.Code = 0x75
	case "jg":
		op.Code = 0x76
	case "call":
		op.Code = 0x80
	}

	if !nextReadRegExp(scanner, regLabel) {
		return nil, &MismatchError{Expect: "label", Actual: scanner.Text()}
	}
	op.Label = scanner.Text()

	if _, ok := labels[op.Label]; !ok {
		labels[op.Label] = nil
	}

	return op, nil
}

func ret(addr int, labels labelTable, scanner *bufio.Scanner) (*opcode, error) {
	return &opcode{
		Address:    addr,
		Code:       0x90,
		Size:       1,
		EncodeFunc: encodeSimpleOpCode,
	}, nil
}

func stack(addr int, labels labelTable, scanner *bufio.Scanner) (*opcode, error) {
	op := &opcode{
		Address:    addr,
		Size:       2,
		Rb:         registerCode["no_reg"],
		EncodeFunc: encodeRegisterOpCode,
	}

	switch strings.ToLower(scanner.Text()) {
	case "pushl":
		op.Code = 0xa0
	case "popl":
		op.Code = 0xb0
	}

	var err error
	if op.Ra, err = nextReadRegister(scanner); err != nil {
		return nil, err
	}

	return op, nil
}
