package javaio

import "testing"
import "bufio"
import "bytes"

///////////////////////////////////////
// Parser entry
///////////////////////////////////////

func TestParseServerboundPacketUncompressed(t *testing.T) {
	t.Error("todo")
}

func TestParseServerboundPacketCompressed(t *testing.T) {
	t.Error("todo")
}

///////////////////////////////////////
// Packets for handshake state
///////////////////////////////////////

func TestParseHandshake(t *testing.T) {
	t.Error("todo")
}

///////////////////////////////////////
// Packets for status state
///////////////////////////////////////

func TestParseStatusRequest(t *testing.T) {
	t.Error("todo")
}

func TestParsePing(t *testing.T) {
	t.Error("todo")
}

///////////////////////////////////////
// Packets for login state
///////////////////////////////////////

///////////////////////////////////////
// Packets for play state
///////////////////////////////////////

///////////////////////////////////////
// Basic types
///////////////////////////////////////

func TestParseVarInt(t *testing.T) {
	iomap := []struct {
		input []byte
		output int
	} {
		{[]byte {0x00, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF},           0},
		{[]byte {0x01, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF},           1},
		{[]byte {0x02, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF},           2},
		{[]byte {0x7f, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF},         127},
		{[]byte {0x80, 0x01, 0xCC, 0xDD, 0xEE, 0xFF},         128},
		{[]byte {0xff, 0x01, 0xCC, 0xDD, 0xEE, 0xFF},         255},
		{[]byte {0xff, 0xff, 0xff, 0xff, 0x07, 0xFF},  2147483647},
		{[]byte {0xff, 0xff, 0xff, 0xff, 0x0f, 0xFF}, -         1},
		{[]byte {0x80, 0x80, 0x80, 0x80, 0x08, 0xFF}, -2147483647},
	}

	for i, mapping := range iomap {
		output, err := ParseVarInt(bufio.NewReader(bytes.NewReader(mapping.input)))

		if err != nil {
			t.Error(err)
			continue
		}

		if output != mapping.output {
			t.Errorf("Output incorrect for mapping %d", i)
		}
	}

	iemap := [][]byte {
		{0x80                              }, // Abrupt ending
		{0xff, 0xff, 0xff, 0xff            }, // Abrupt ending
		{0xff, 0xff, 0xff, 0xff, 0xff, 0x0f}, // Exceeds max length
	}

	for i, mappingInput := range iemap {
		_, err := ParseVarInt(bufio.NewReader(bytes.NewReader(mappingInput)))

		if _, ok := err.(*MalformedPacketError); !ok {
			t.Errorf("Expected MalformedPacketError for mapping %d but instead got: %v", i, err)
		}
	}
}

func TestLong(t *testing.T) {
	t.Error("todo")
}

func TestUnsignedShort(t *testing.T) {
	t.Error("todo")
}

func TestParseString(t *testing.T) {
	t.Error("todo")
}