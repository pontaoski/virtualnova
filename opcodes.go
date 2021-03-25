package main

import "fmt"

// the system is big-endian

// instructions: 00 00   00 00 00 00   00 00 00 00
//               |  |    | src         | dst
//               |  |
//               |  | size
//               |
//               | opcode

type opcode uint8

const (
	// halt and catch fire
	hcf opcode = 0x00

	// adds r0 + r1 -> r2
	add opcode = 0x01

	// subtracts r0 - r1 -> r2
	sub opcode = 0x02

	// divides r0 / r1 -> r2
	div opcode = 0x03

	// multiplies r0 * r1 -> r3
	mul opcode = 0x04

	// loads the value of src into the register named by dst
	loadi opcode = 0x05

	// copies src->dst
	move opcode = 0x06

	// reads address into register
	load opcode = 0x07

	// stores register into address
	store opcode = 0x08

	// swaps the halves of the register
	swap opcode = 0x09

	// exchanges the two provided registers with eachother
	exchange opcode = 0x0a

	// ~src(reg) -> dest(reg)
	not opcode = 0x0b

	// src(reg) & dest(reg) -> dest(reg)
	and opcode = 0x0c

	// src(reg) | dest(reg) -> dest(reg)
	or opcode = 0x0d

	// src(reg) ^ dest(reg) -> dest(reg)
	xor opcode = 0x0e

	// src(reg) -> pc
	jmp opcode = 0x0f

	// dest(reg) -> pc if src(reg) == 0
	jmpEQ = 0x10

	// dest(reg) -> pc if src(reg) != 0
	jmpNEQ = 0x11

	// dest(reg) -> stdout
	dout = 0x12

	// src(reg) -> dest(reg)
	copyr opcode = 0x13

	// mu
	mu opcode = 0xff

	// todo: convenience bit operations
)

type size uint8

const (
	// 8 bits
	byt size = iota
	// 16 bits
	word
	// 32 bits
	longword

	// 8 bits; constant last
	constantByt

	// 16 bits; constant last
	constantWord

	// 32 bits; constant last
	constantLongword
)

type address uint32

type instruction struct {
	opcode opcode
	size   size
	source address
	dest   address
}

type packed [2]uint64

func (p packed) String() string {
	return fmt.Sprintf("0b%064b%064b", p[0], p[1])
}
