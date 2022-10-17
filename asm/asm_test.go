package asm

import (
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestParse_nop(t *testing.T) {
	expect := []byte{0x00}

	Convey("test pase nop", t, func() {
		actual, _ := Parse(strings.NewReader("nop"))
		So(actual, ShouldResemble, expect)
	})
}

func TestParse_halt(t *testing.T) {
	expect := []byte{0x10}

	Convey("test parse hald", t, func() {
		actual, _ := Parse(strings.NewReader("halt"))
		So(actual, ShouldResemble, expect)
	})
}

func TestParse_rrmovl(t *testing.T) {
	expect := []byte{0x20, 0x03}

	Convey("test parse rrmovl", t, func() {
		Convey("have space between two register names", func() {
			actual, _ := Parse(strings.NewReader("rrmovl %eax, %ebx"))
			So(actual, ShouldResemble, expect)
		})
		Convey("no space between two register names", func() {
			actual, _ := Parse(strings.NewReader("rrmovl %eax,%ebx"))
			So(actual, ShouldResemble, expect)
		})
	})
}

func TestParse_irmovl(t *testing.T) {
	expect := []byte{0x30, 0x83, 0x78, 0x56, 0x34, 0x12}

	Convey("test parse irmovl", t, func() {

		Convey("test hex immediate number", func() {
			actual, _ := Parse(strings.NewReader("irmovl $0x12345678, %ebx"))
			So(actual, ShouldResemble, expect)
		})
		Convey("test decimal immediate number", func() {
			actual, _ := Parse(strings.NewReader("irmovl $305419896, %ebx"))
			So(actual, ShouldResemble, expect)
		})
		Convey("test bin immediate number", func() {
			actual, _ := Parse(strings.NewReader("irmovl $0b00010010001101000101011001111000, %ebx"))
			So(actual, ShouldResemble, expect)
		})
		Convey("test oct immediate number", func() {
			actual, _ := Parse(strings.NewReader("irmovl $0o2215053170, %ebx"))
			So(actual, ShouldResemble, expect)
		})

		Convey("test nag immediate number", func() {
			expect := []byte{0x30, 0x83, 0x88, 0xa9, 0xcb, 0xed}
			actual, _ := Parse(strings.NewReader("irmovl $-0x12345678, %ebx"))
			So(actual, ShouldResemble, expect)
		})
	})
}

func TestParse_rmmvol(t *testing.T) {
	expect := []byte{0x40, 0x03, 0x78, 0x56, 0x34, 0x12}

	Convey("test parse rmmovl", t, func() {
		actual, _ := Parse(strings.NewReader("rmmvol %eax, 0x12345678(%ebx)"))
		So(actual, ShouldResemble, expect)
	})
}

func TestParse_mrmovl(t *testing.T) {
	expect := []byte{0x50, 0x30, 0x78, 0x56, 0x34, 0x12}

	Convey("test parse mrmovl", t, func() {
		actual, _ := Parse(strings.NewReader("mrmovl 0x12345678(%eax), %ebx"))
		So(actual, ShouldResemble, expect)
	})
}

func TestParse_addl(t *testing.T) {
	Convey("test parse addl", t, func() {
		expect := []byte{0x60, 0x30}
		actual, _ := Parse(strings.NewReader("addl %ebx, %eax"))
		So(actual, ShouldResemble, expect)
	})
}

func TestParse_subl(t *testing.T) {
	Convey("test parse subl", t, func() {
		expect := []byte{0x61, 0x13}
		actual, _ := Parse(strings.NewReader("subl %ecx, %ebx"))
		So(actual, ShouldResemble, expect)
	})
}

func TestParse_andl(t *testing.T) {
	Convey("test parse andl", t, func() {
		expect := []byte{0x62, 0x13}
		actual, _ := Parse(strings.NewReader("andl %ecx, %ebx"))
		So(actual, ShouldResemble, expect)
	})
}

func TestParse_xorl(t *testing.T) {
	Convey("test parse xorl", t, func() {
		expect := []byte{0x63, 0x13}
		actual, _ := Parse(strings.NewReader("xorl %ecx, %ebx"))
		So(actual, ShouldResemble, expect)
	})
}

func TestParse_jmp(t *testing.T) {
	Convey("test parse jmp", t, func() {
		Convey("test backward reference", func() {
			expect := []byte{0x70, 0x00, 0x00, 0x00, 0x00}
			actual, _ := Parse(strings.NewReader("loop: jmp loop"))
			So(actual, ShouldResemble, expect)
		})
		Convey("test forward reference", func() {
			expect := []byte{0x70, 0x05, 0x00, 0x00, 0x00}
			actual, _ := Parse(strings.NewReader("jmp end\nend:"))
			So(actual, ShouldResemble, expect)
		})
	})
}

func TestParse_jle(t *testing.T) {
	Convey("test parse jle", t, func() {
		expect := []byte{0x71, 0x00, 0x00, 0x00, 0x00}
		actual, _ := Parse(strings.NewReader("loop: jle loop"))
		So(actual, ShouldResemble, expect)
	})
}

func TestParse_jl(t *testing.T) {
	Convey("test parse jl", t, func() {
		expect := []byte{0x72, 0x00, 0x00, 0x00, 0x00}
		actual, _ := Parse(strings.NewReader("loop: jl loop"))
		So(actual, ShouldResemble, expect)
	})
}

func TestParse_je(t *testing.T) {
	Convey("test parse je", t, func() {
		expect := []byte{0x73, 0x00, 0x00, 0x00, 0x00}
		actual, _ := Parse(strings.NewReader("loop: je loop"))
		So(actual, ShouldResemble, expect)
	})
}

func TestParse_jne(t *testing.T) {
	Convey("test parse jne", t, func() {
		expect := []byte{0x74, 0x00, 0x00, 0x00, 0x00}
		actual, _ := Parse(strings.NewReader("loop: jne loop"))
		So(actual, ShouldResemble, expect)
	})
}

func TestParse_jge(t *testing.T) {
	Convey("test parse jge", t, func() {
		expect := []byte{0x75, 0x00, 0x00, 0x00, 0x00}
		actual, _ := Parse(strings.NewReader("loop: jge loop"))
		So(actual, ShouldResemble, expect)
	})
}

func TestParse_jg(t *testing.T) {
	Convey("test parse jg", t, func() {
		expect := []byte{0x76, 0x00, 0x00, 0x00, 0x00}
		actual, _ := Parse(strings.NewReader("loop: jg loop"))
		So(actual, ShouldResemble, expect)
	})
}

func TestParse_call(t *testing.T) {
	Convey("test parse call", t, func() {
		expect := []byte{0x80, 0x00, 0x00, 0x00, 0x00}
		actual, _ := Parse(strings.NewReader("loop: call loop"))
		So(actual, ShouldResemble, expect)
	})
}

func TestParse_ret(t *testing.T) {
	Convey("test parse ret", t, func() {
		expect := []byte{0x90}
		actual, _ := Parse(strings.NewReader("ret"))
		So(actual, ShouldResemble, expect)
	})
}

func TestParse_pushl(t *testing.T) {
	Convey("test parse pushl", t, func() {
		expect := []byte{0xa0, 0x38}
		actual, _ := Parse(strings.NewReader("pushl %ebx"))
		So(actual, ShouldResemble, expect)
	})
}

func TestParse_popl(t *testing.T) {
	Convey("test parse popl", t, func() {
		expect := []byte{0xb0, 0x38}
		actual, _ := Parse(strings.NewReader("popl %ebx"))
		So(actual, ShouldResemble, expect)
	})
}

func TestParse(t *testing.T) {
	src := "irmovl $0, %eax\n" +
		"irmovl $1, %ebx\n" +
		"irmovl $10, %ecx\n" +
		"loop:\n" +
		"addl %ebx, %eax\n" +
		"subl %ecx, %ebx\n" +
		"jne loop\n" +
		"halt"

	expect := []byte{
		0x30, 0x80, 0x00, 0x00, 0x00, 0x00,
		0x30, 0x83, 0x01, 0x00, 0x00, 0x00,
		0x30, 0x81, 0x0a, 0x00, 0x00, 0x00,
		0x60, 0x30,
		0x61, 0x13,
		0x74, 0x12, 0x00, 0x00, 0x00,
		0x10,
	}

	Convey("test parse src", t, func() {
		actual, _ := Parse(strings.NewReader(src))
		So(actual, ShouldResemble, expect)
	})
}
