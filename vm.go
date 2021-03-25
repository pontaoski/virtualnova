package main

import (
	"encoding/binary"
	"fmt"

	"github.com/jedib0t/go-pretty/v6/table"
)

type memory int

const (
	// one byte
	abyte     memory = 1
	akilobyte        = abyte * 1024
	amegabyte        = abyte * 1024 * 1024
)

const romsize = 2 * amegabyte
const ramsize = 512 * akilobyte

const screenWidth = 320
const screenHeight = 240

const screenSizePixels = screenWidth * screenHeight
const screenSizeBytes = screenSizePixels * 3

type vm struct {
	rom [romsize]byte
	ram [ramsize]byte

	registers [16]uint32

	pc address
}

func (v *vm) read(addr address) byte {
	if int(addr) < int(2*amegabyte) {
		return v.rom[addr]
	} else {
		return v.ram[int(addr)-int(2*amegabyte)]
	}
}

func (v *vm) write(addr address, va byte) {
	if int(addr) < int(2*amegabyte) {
		panic(fmt.Sprintf("tried to write to ROM: %d", addr))
	} else {
		v.rom[int(addr)-int(2*amegabyte)] = va
	}
}

func decode(inst packed, out *instruction) {
	out.opcode = opcode((inst[0] >> (8 * 0)) & 0xff)
	out.size = size((inst[0] >> (8 * 1)) & 0xff)
	out.source = address((inst[0] >> (8 * 2)) & 0xffff)
	out.dest = address(inst[1] & 0xffff)
}

func printb(v interface{}) {
	fmt.Printf("%b\n", v)
}

func (v *vm) tick() {
	v.pc += 10

	inst := instruction{}
	inst.opcode = opcode(v.read(v.pc))
	inst.size = size(v.read(v.pc + 1))
	inst.source = address(binary.BigEndian.Uint32([]byte{v.read(v.pc + 2), v.read(v.pc + 3), v.read(v.pc + 4), v.read(v.pc + 5)}))
	inst.dest = address(binary.BigEndian.Uint32([]byte{v.read(v.pc + 6), v.read(v.pc + 7), v.read(v.pc + 8), v.read(v.pc + 9)}))

	switch inst.opcode {
	case hcf:
		panic("halt and catch fire")
	case add:
		v.registers[2] = v.registers[0] + v.registers[1]
	case sub:
		v.registers[2] = v.registers[0] - v.registers[1]
	case div:
		v.registers[2] = v.registers[0] / v.registers[1]
	case mul:
		v.registers[2] = v.registers[0] * v.registers[1]
	case loadi:
		place := int(inst.dest & 0xff)

		switch inst.size {
		case byt:
			v.registers[place] = uint32(inst.source & 0xff)
		case word:
			v.registers[place] = uint32(inst.source & 0xffff)
		case longword:
			v.registers[place] = uint32(inst.source & 0xffffffff)
		default:
			panic("bad size")
		}
	case copyr:
		v.registers[inst.dest] = v.registers[inst.source]
	case move:
		src := inst.source
		dest := inst.dest

		switch inst.size {
		case byt:
			v.write(dest, v.read(src))
		case word:
			v.write(dest, v.read(src))
			v.write(dest+1, v.read(src+1))
		case longword:
			v.write(dest, v.read(src))
			v.write(dest+1, v.read(src+1))
			v.write(dest+2, v.read(src+2))
			v.write(dest+3, v.read(src+3))

		default:
			panic("bad size")
		}
	case store:
		from := inst.source & 0xff

		switch inst.size {
		case byt:
			v.write(inst.dest, byte(v.registers[from]&0xff))
		case word:
			v.write(inst.dest, byte(v.registers[from]&0xff))
			v.write(inst.dest+1, byte(v.registers[from]&0xff00))
		case longword:
			v.write(inst.dest, byte(v.registers[from]&0xff))
			v.write(inst.dest+1, byte(v.registers[from]&0xff00))
			v.write(inst.dest+2, byte(v.registers[from]&0xff0000))
			v.write(inst.dest+3, byte(v.registers[from]&0xff000000))
		}
	case load:
		to := inst.dest & 0xff
		switch inst.size {
		case byt:
			v.registers[to] = uint32(v.read(inst.source))
		case word:
			v.registers[to] = binary.BigEndian.Uint32([]byte{0x00, 0x00, v.read(inst.source), v.read(inst.source + 1)})
		case longword:
			v.registers[to] = binary.BigEndian.Uint32([]byte{v.read(inst.source), v.read(inst.source + 1), v.read(inst.source + 2), v.read(inst.source + 3)})
		}
	case swap:
		reg := inst.dest & 0xff
		p := make([]byte, 4)
		binary.BigEndian.PutUint32(p, v.registers[reg])
		p[0], p[1], p[2], p[3] = p[2], p[3], p[0], p[1]
		v.registers[reg] = binary.BigEndian.Uint32(p)
	case exchange:
		v.registers[int(inst.source)], v.registers[int(inst.dest)] = v.registers[int(inst.dest)], v.registers[int(inst.source)]
	case not:
		v.registers[int(inst.dest)] = v.registers[int(inst.source)] ^ 0
	case and:
		v.registers[int(inst.dest)] = v.registers[int(inst.source)] & v.registers[int(inst.dest)]
	case xor:
		v.registers[int(inst.dest)] = v.registers[int(inst.source)] ^ v.registers[int(inst.dest)]
	case jmp:
		switch inst.size {
		case byt:
			v.pc = address((v.registers[int(inst.source)]))
		case constantLongword:
			v.pc = address(inst.source)
		}
		v.pc -= 10
	case jmpEQ:
		reg := v.registers[int(inst.source)]

		if reg != 0 {
			return
		}

		switch inst.size {
		case byt:
			v.pc = address((v.registers[int(inst.dest)]))
		case constantLongword:
			v.pc = address(inst.dest)
		}
		v.pc -= 10
	case jmpNEQ:
		reg := v.registers[int(inst.source)]

		if reg == 0 {
			return
		}

		switch inst.size {
		case byt:
			v.pc = address((v.registers[int(inst.dest)]))
		case constantLongword:
			v.pc = address(inst.dest)
		}
		v.pc -= 10
	case mu:
		println("mu")
	default:
		panic("bad instruction: " + fmt.Sprintf("%b", inst.opcode))
	}
}

func hx(v interface{}) string {
	return fmt.Sprintf("%#b", v)
}

func (v *vm) dumpRegisters() {
	t := table.NewWriter()
	t.AppendHeader(table.Row{"R0", "R1", "R2", "R3", "R4", "R5", "R6", "R7"})
	t.AppendRow(table.Row{
		hx(v.registers[0]),
		hx(v.registers[1]),
		hx(v.registers[2]),
		hx(v.registers[3]),
		hx(v.registers[4]),
		hx(v.registers[5]),
		hx(v.registers[6]),
		hx(v.registers[7]),
	})
	println(t.Render())
}
