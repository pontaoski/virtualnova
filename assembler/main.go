package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"text/scanner"
)

func expect(s *scanner.Scanner, r rune) {
	i := s.Scan()
	if r != i {
		panic(fmt.Sprintf("%s: expected %q got %q", s.Pos(), r, i))
	}
}

func reg(s *scanner.Scanner) byte {
	i := s.Scan()
	if i != scanner.Ident {
		panic(fmt.Sprintf("%s: expected ident got %s", s.Pos(), s.TokenText()))
	}

	data := map[string]byte{
		"reg0":  0x00,
		"reg1":  0x01,
		"reg2":  0x02,
		"reg3":  0x03,
		"reg4":  0x04,
		"reg5":  0x05,
		"reg6":  0x06,
		"reg7":  0x07,
		"reg8":  0x08,
		"reg9":  0x09,
		"reg10": 0x0a,
		"reg11": 0x0b,
		"reg12": 0x0c,
		"reg13": 0x0d,
		"reg14": 0x0e,
		"reg15": 0x0f,
	}
	return data[s.TokenText()]
}

func size(s *scanner.Scanner) byte {
	i := s.Scan()
	if i != scanner.Ident {
		panic(fmt.Sprintf("%s: expected ident got %s", s.Pos(), s.TokenText()))
	}

	data := map[string]byte{
		"byte":     0x00,
		"word":     0x01,
		"longword": 0x02,
	}
	return data[s.TokenText()]
}

func expectIdent(s *scanner.Scanner) string {
	i := s.Scan()
	if i != scanner.Ident {
		panic(fmt.Sprintf("%s: expected ident got %s", s.Pos(), s.TokenText()))
	}

	return s.TokenText()
}

func expectS(s *scanner.Scanner, st string) {
	id := expectIdent(s)
	if id != st {
		panic(fmt.Sprintf("%s: expected %s got %s", s.Pos(), st, id))
	}
}

func expectInt(s *scanner.Scanner) uint32 {
	i := s.Scan()
	if i != scanner.Int {
		panic(fmt.Sprintf("%s: expected int got %s", s.Pos(), s.TokenText()))
	}

	data, err := strconv.ParseUint(s.TokenText(), 10, 32)
	if err != nil {
		panic(err)
	}

	return uint32(data)
}

func parse(s *scanner.Scanner, buf *bytes.Buffer, replacements map[int]string) {
	fromDest := func(opcode byte) {
		src := reg(s)
		expect(s, '-')
		expect(s, '>')
		dest := reg(s)
		buf.Write([]byte{opcode, 0x00 /**/, 0x00, 0x00, 0x00, src /**/, 0x00, 0x00, 0x00, dest})
	}
	r := s.Scan()
	switch r {
	case scanner.Ident:
		switch s.TokenText() {
		case "hcf":
			buf.Write([]byte{0x00, 0x00 /**/, 0x00, 0x00, 0x00, 0x00 /**/, 0x00, 0x00, 0x00, 0x00})
			return
		case "mu":
			buf.Write([]byte{0xff, 0x00 /**/, 0x00, 0x00, 0x00, 0x00 /**/, 0x00, 0x00, 0x00, 0x00})
			return
		case "add":
			buf.Write([]byte{0x01, 0x00 /**/, 0x00, 0x00, 0x00, 0x00 /**/, 0x00, 0x00, 0x00, 0x00})
			return
		case "sub":
			buf.Write([]byte{0x02, 0x00 /**/, 0x00, 0x00, 0x00, 0x00 /**/, 0x00, 0x00, 0x00, 0x00})
			return
		case "div":
			buf.Write([]byte{0x03, 0x00 /**/, 0x00, 0x00, 0x00, 0x00 /**/, 0x00, 0x00, 0x00, 0x00})
			return
		case "mul":
			buf.Write([]byte{0x04, 0x00 /**/, 0x00, 0x00, 0x00, 0x00 /**/, 0x00, 0x00, 0x00, 0x00})
			return
		case "move":
			sz := size(s)
			src := reg(s)
			expect(s, '-')
			expect(s, '>')
			dest := reg(s)
			buf.Write([]byte{0x06, sz /**/, 0x00, 0x00, 0x00, src /**/, 0x00, 0x00, 0x00, dest})
			return
		case "load":
			sz := size(s)
			dest := reg(s)
			expect(s, '<')
			expect(s, '-')

			from := expectInt(s)

			b := make([]byte, 4)
			binary.BigEndian.PutUint32(b, uint32(from))

			buf.Write([]byte{0x07, sz /**/, b[0], b[1], b[2], b[3] /**/, 0x00, 0x00, 0x00, dest})
			return
		case "store":
			sz := size(s)
			src := reg(s)
			expect(s, '-')
			expect(s, '>')

			dest := expectInt(s)

			b := make([]byte, 4)
			binary.BigEndian.PutUint32(b, uint32(dest))

			buf.Write([]byte{0x08, sz /**/, 0x00, 0x00, 0x00, src /**/, b[0], b[1], b[2], b[3]})
			return
		case "swap":
			buf.Write([]byte{0x09, 0x00 /**/, 0x00, 0x00, 0x00, 0x00 /**/, 0x00, 0x00, 0x00, reg(s)})
			return
		case "exchange":
			r1 := reg(s)
			expect(s, '<')
			expect(s, '-')
			expect(s, '>')
			r2 := reg(s)

			buf.Write([]byte{0x0a, 0x00 /**/, 0x00, 0x00, 0x00, r1 /**/, 0x00, 0x00, 0x00, r2})
			return
		case "not":
			fromDest(0x0b)
			return
		case "and":
			fromDest(0x0c)
			return
		case "or":
			fromDest(0x0d)
			return
		case "xor":
			fromDest(0x0e)
			return
		case "copy":
			fromDest(0x13)
		case "jump":
			jumpTo := expectIdent(s)

			if jumpTo == "to" {
				jumpTo = expectIdent(s)

				expectS(s, "if")

				src := reg(s)

				expectS(s, "is")

				cond := expectIdent(s)

				// write our opcode
				switch cond {
				case "not":
					expectS(s, "equal")
					expectS(s, "to")

					buf.Write([]byte{0x11, 0x05})
				case "equal":
					expectS(s, "to")

					buf.Write([]byte{0x10, 0x05})
				}

				expectS(s, "zero")

				// write our src
				buf.Write([]byte{0x00, 0x00, 0x00, src})

				// note the index
				replacements[buf.Len()] = jumpTo

				// add padding
				buf.Write([]byte{0x00, 0x00, 0x00, 0x00})

				return
			}

			// write our opcode
			buf.Write([]byte{0x0f, 0x05})

			// note the index
			replacements[buf.Len()] = jumpTo

			buf.Write([]byte{0x00, 0x00, 0x00, 0x00 /**/, 0x00, 0x00, 0x00, 0x00})
			return
		default:
			panic("unhandled ident: '" + s.TokenText() + "' at " + s.Pos().String())
		}
	case scanner.Int:
		data, err := strconv.ParseUint(s.TokenText(), 10, 32)
		if err != nil {
			panic(err)
		}
		expect(s, '-')
		expect(s, '>')

		b := make([]byte, 4)
		binary.BigEndian.PutUint32(b, uint32(data))

		rg := reg(s)
		sz := size(s)

		buf.Write([]byte{0x05, sz /**/, b[0], b[1], b[2], b[3] /**/, 0x00, 0x00, 0x00, rg})
		return
	case '|':
		parse(s, buf, replacements)
		parse(s, buf, replacements)
	default:
		panic(fmt.Sprintf("unhandled: %s at %s", string(r), s.Pos()))
	}
}

func parseBlocks(s *scanner.Scanner) []byte {
	positions := map[string]int{}
	replacements := map[int]string{}

	b := new(bytes.Buffer)

outermost:
	for {
		r := s.Scan()
		switch r {
		case scanner.Ident:
			t := s.TokenText()
			expect(s, ':')
			expect(s, '{')

			positions[t] = b.Len()

			parse(s, b, replacements)

			expect(s, '}')
		default:
			if r != scanner.EOF {
				panic(r)
			}
			break outermost
		}
	}

	data := b.Bytes()

	for idx, sym := range replacements {
		pos := positions[sym]

		b := make([]byte, 4)
		binary.BigEndian.PutUint32(b, uint32(pos))

		data[idx] = b[0]
		data[idx+1] = b[1]
		data[idx+2] = b[2]
		data[idx+3] = b[3]
	}

	return data
}

func main() {
	data, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	s := scanner.Scanner{}
	s.Init(bytes.NewReader(data))

	out := parseBlocks(&s)

	ioutil.WriteFile(os.Args[2], out, os.ModePerm)
}
