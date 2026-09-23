package mobile

import (
	"bytes"
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

func TestPlanForWayCutGivesTwoSegments(t *testing.T) {
	hello := helloWith("gateway.discord.gg")

	p, err := PlanForWay(hello, "cut")
	if err != nil {
		t.Fatalf("PlanForWay: %v", err)
	}
	if p.Count() != 2 {
		t.Fatalf("Count = %d, want 2", p.Count())
	}

	joined := append(append([]byte{}, p.Bytes(0)...), p.Bytes(1)...)
	if !bytes.Equal(joined, hello) {
		t.Error("joined segments differ from the hello")
	}
	if p.TTL(0) != 0 {
		t.Errorf("TTL(0) = %d, want 0", p.TTL(0))
	}
}

func TestPlanForWayUnknownWayReturnsError(t *testing.T) {
	_, err := PlanForWay(helloWith("example.com"), "nope")
	if !errors.Is(err, ErrUnknownWay) {
		t.Errorf("err = %v, want ErrUnknownWay", err)
	}
}

func TestPlanIndexOutOfRangeIsSafe(t *testing.T) {
	p, err := PlanForWay(helloWith("example.com"), "cut")
	if err != nil {
		t.Fatalf("PlanForWay: %v", err)
	}
	if p.Bytes(-1) != nil || p.Bytes(99) != nil {
		t.Error("out-of-range Bytes should be nil")
	}
	if p.TTL(99) != 0 {
		t.Error("out-of-range TTL should be 0")
	}
}

func TestPlanKeepsItsOwnBytes(t *testing.T) {
	hello := helloWith("example.com")

	p, err := PlanForWay(hello, "cut")
	if err != nil {
		t.Fatalf("PlanForWay: %v", err)
	}

	before := append([]byte{}, p.Bytes(0)...)
	for i := range hello {
		hello[i] = 0
	}
	if !bytes.Equal(p.Bytes(0), before) {
		t.Error("plan bytes changed after the input was trashed")
	}
}

func TestPlanForHostAppliesTheRule(t *testing.T) {
	hello := helloWith("discord.com")

	p, err := PlanForHost(hello, "discord.com", "discord.com=cut:name")
	if err != nil {
		t.Fatalf("PlanForHost: %v", err)
	}
	if p.Count() != 2 {
		t.Errorf("segments = %d, want 2 (a cut)", p.Count())
	}
}

func TestPlanForHostFakeThenCut(t *testing.T) {
	hello := helloWith("discord.com")

	p, err := PlanForHost(hello, "discord.com", "discord.com=decoy:auto,cut:name")
	if err != nil {
		t.Fatalf("PlanForHost: %v", err)
	}
	if p.Count() != 3 {
		t.Errorf("segments = %d, want 3 (decoy + two)", p.Count())
	}
}

func TestPlanForHostWithNoRuleSaysSo(t *testing.T) {
	hello := helloWith("discord.com")

	_, err := PlanForHost(hello, "discord.com", "other.com=cut:name")
	if !errors.Is(err, ErrNoRule) {
		t.Errorf("err = %v, want ErrNoRule", err)
	}
}

func TestPlanForHostReportsBadRules(t *testing.T) {
	hello := helloWith("discord.com")

	if _, err := PlanForHost(hello, "discord.com", "discord.com=cut"); err == nil {
		t.Fatal("bad rule text: want error, got nil")
	}
}

func TestPlanForStrategyCuts(t *testing.T) {
	hello := helloWith("discord.com")

	p, err := PlanForStrategy(hello, "cut:name")
	if err != nil {
		t.Fatalf("PlanForStrategy: %v", err)
	}
	if p.Count() != 2 {
		t.Errorf("segments = %d, want 2", p.Count())
	}
}

func TestPlanForStrategyFakeThenCut(t *testing.T) {
	hello := helloWith("discord.com")

	p, err := PlanForStrategy(hello, "decoy:auto,ttl:4,cut:name")
	if err != nil {
		t.Fatalf("PlanForStrategy: %v", err)
	}
	if p.Count() != 3 {
		t.Errorf("segments = %d, want 3", p.Count())
	}
	if p.TTL(0) != 4 {
		t.Errorf("decoy TTL = %d, want 4", p.TTL(0))
	}
}

func TestPlanForStrategyRejectsGarbage(t *testing.T) {
	hello := helloWith("discord.com")

	if _, err := PlanForStrategy(hello, "cut"); err == nil {
		t.Fatal("bad strategy: want error, got nil")
	}
}

func TestStrategiesAreOffered(t *testing.T) {
	ways := Strategies()
	if ways.Count() == 0 {
		t.Fatal("no strategies offered")
	}
	if ways.At(0) == "" {
		t.Error("first strategy is empty")
	}
	if ways.At(-1) != "" || ways.At(9999) != "" {
		t.Error("out-of-range At should be empty")
	}
}

func TestEveryOfferedStrategyPlans(t *testing.T) {
	hello := helloWith("gateway.discord.gg")
	ways := Strategies()
	for i := 0; i < ways.Count(); i++ {
		if _, err := PlanForStrategy(hello, ways.At(i)); err != nil {
			t.Errorf("offered strategy %q does not plan: %v", ways.At(i), err)
		}
	}
}

func TestSweepStartsEmpty(t *testing.T) {
	s := NewSweep()
	if s.Found() || s.Best() != "" {
		t.Error("a fresh sweep should have no winner")
	}
}

func TestSweepKeepsTheStrategyWithMostBytes(t *testing.T) {
	s := NewSweep()
	s.Record("cut:name", 40)
	s.Record("decoy:auto,ttl:4", 900)
	s.Record("cut:start", 120)

	if !s.Found() {
		t.Fatal("Found should be true after a positive result")
	}
	if s.Best() != "decoy:auto,ttl:4" {
		t.Errorf("Best = %q, want decoy:auto,ttl:4", s.Best())
	}
}

func TestSweepIgnoresEmptyResults(t *testing.T) {
	s := NewSweep()
	s.Record("cut:name", 0)

	if s.Found() {
		t.Error("a zero-byte result is not a win")
	}
}
