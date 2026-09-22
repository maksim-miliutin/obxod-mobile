package plan

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func helloWith(host string) []byte {
	name := []byte(host)

	list := binary.BigEndian.AppendUint16(nil, uint16(len(name)+3))
	list = append(list, 0x00)
	list = binary.BigEndian.AppendUint16(list, uint16(len(name)))
	list = append(list, name...)

	extension := binary.BigEndian.AppendUint16(nil, 0x0000)
	extension = binary.BigEndian.AppendUint16(extension, uint16(len(list)))
	extension = append(extension, list...)

	block := binary.BigEndian.AppendUint16(nil, uint16(len(extension)))
	block = append(block, extension...)

	body := []byte{0x03, 0x03}
	body = append(body, make([]byte, 32)...)
	body = append(body, 0x00)
	body = append(body, 0x00, 0x02, 0x13, 0x01)
	body = append(body, 0x01, 0x00)
	body = append(body, block...)

	handshake := []byte{0x01, byte(len(body) >> 16), byte(len(body) >> 8), byte(len(body))}
	handshake = append(handshake, body...)

	record := []byte{0x16, 0x03, 0x01, 0x00, 0x00}
	binary.BigEndian.PutUint16(record[3:5], uint16(len(handshake)))

	return append(record, handshake...)
}

func hostOffset(hello []byte, host string) int {
	return bytes.Index(hello, []byte(host))
}

func TestCutSplitsInTwoAndKeepsBytes(t *testing.T) {
	hello := helloWith("gateway.discord.gg")

	p, err := Cut(hello)
	if err != nil {
		t.Fatalf("Cut: %v", err)
	}
	if len(p.Segments) != 2 {
		t.Fatalf("segments = %d, want 2", len(p.Segments))
	}

	joined := append(append([]byte{}, p.Segments[0].Bytes...), p.Segments[1].Bytes...)
	if !bytes.Equal(joined, hello) {
		t.Error("joined segments differ from the original hello")
	}
}

func TestCutBreaksInsideTheHost(t *testing.T) {
	host := "gateway.discord.gg"
	hello := helloWith(host)

	p, err := Cut(hello)
	if err != nil {
		t.Fatalf("Cut: %v", err)
	}

	at := len(p.Segments[0].Bytes)
	start := hostOffset(hello, host)
	if at <= start || at >= start+len(host) {
		t.Errorf("split at %d, want strictly inside host [%d, %d)", at, start, start+len(host))
	}
}

func TestCutOnGarbageReturnsError(t *testing.T) {
	if _, err := Cut([]byte{0x00, 0x01, 0x02}); err == nil {
		t.Fatal("Cut on garbage: want error, got nil")
	}
}
