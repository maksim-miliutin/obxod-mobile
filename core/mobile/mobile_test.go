package mobile

import (
	"encoding/binary"
	"errors"
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

func TestServerNameReadsHost(t *testing.T) {
	name, err := ServerName(helloWith("gateway.discord.gg"))
	if err != nil {
		t.Fatalf("ServerName: %v", err)
	}
	if name != "gateway.discord.gg" {
		t.Errorf("name = %q, want gateway.discord.gg", name)
	}
}

func TestServerNameOnGarbageReturnsError(t *testing.T) {
	_, err := ServerName([]byte{0x00, 0x01, 0x02})
	if err == nil {
		t.Fatal("ServerName on garbage: want error, got nil")
	}
}

func TestGuardTurnsPanicIntoError(t *testing.T) {
	_, err := guard(func() (int, error) {
		panic("boom")
	})
	if !errors.Is(err, ErrPanicked) {
		t.Errorf("err = %v, want ErrPanicked", err)
	}
}
