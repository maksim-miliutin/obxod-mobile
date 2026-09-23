package plan

import (
	"bytes"
	"encoding/binary"
	"errors"
	"testing"

	"obxod/internal/clienthello"
	"obxod/internal/rules"
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

	padding := binary.BigEndian.AppendUint16(nil, 0x0015)
	padding = binary.BigEndian.AppendUint16(padding, 4)
	padding = append(padding, 0, 0, 0, 0)

	extension = append(extension, padding...)

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

func hostStart(hello []byte, host string) int {
	return bytes.Index(hello, []byte(host))
}

func splitAt(t *testing.T, p Plan) int {
	t.Helper()
	if len(p.Segments) != 2 {
		t.Fatalf("segments = %d, want 2", len(p.Segments))
	}

	return len(p.Segments[0].Bytes)
}

func TestCutStillSplitsInsideTheHost(t *testing.T) {
	host := "gateway.discord.gg"
	hello := helloWith(host)

	p, err := Cut(hello)
	if err != nil {
		t.Fatalf("Cut: %v", err)
	}

	at := splitAt(t, p)
	start := hostStart(hello, host)
	if at <= start || at >= start+len(host) {
		t.Errorf("split at %d, want inside host [%d, %d)", at, start, start+len(host))
	}
}

func TestFromRuleCutAfterSplitsPastTheHost(t *testing.T) {
	host := "gateway.discord.gg"
	hello := helloWith(host)

	p, err := FromRule(hello, rules.Rule{Cut: "after"})
	if err != nil {
		t.Fatalf("FromRule: %v", err)
	}

	at := splitAt(t, p)
	if at != hostStart(hello, host)+len(host) {
		t.Errorf("after: split at %d, want %d", at, hostStart(hello, host)+len(host))
	}
}

func TestFromRuleCutStartSplitsNearTheStart(t *testing.T) {
	hello := helloWith("gateway.discord.gg")

	p, err := FromRule(hello, rules.Rule{Cut: "start"})
	if err != nil {
		t.Fatalf("FromRule: %v", err)
	}

	if at := splitAt(t, p); at != 2 {
		t.Errorf("start: split at %d, want 2", at)
	}
}

func TestFromRuleKeepsAllBytes(t *testing.T) {
	hello := helloWith("gateway.discord.gg")

	for _, mode := range []string{"name", "after", "start"} {
		p, err := FromRule(hello, rules.Rule{Cut: mode})
		if err != nil {
			t.Fatalf("FromRule %q: %v", mode, err)
		}

		joined := append(append([]byte{}, p.Segments[0].Bytes...), p.Segments[1].Bytes...)
		if !bytes.Equal(joined, hello) {
			t.Errorf("%q: joined segments differ from the hello", mode)
		}
	}
}

func TestFromRuleWithoutCutIsNoSocketWay(t *testing.T) {
	hello := helloWith("example.com")

	if _, err := FromRule(hello, rules.Rule{}); !errors.Is(err, ErrNoSocketWay) {
		t.Errorf("empty rule: err = %v, want ErrNoSocketWay", err)
	}
}

func TestFromRuleOnGarbageReturnsError(t *testing.T) {
	if _, err := FromRule([]byte{0x00, 0x01}, rules.Rule{Cut: "name"}); err == nil {
		t.Fatal("garbage: want error, got nil")
	}
}

func parseName(t *testing.T, hello []byte) string {
	t.Helper()
	parsed, err := clienthello.Parse(hello)
	if err != nil {
		t.Fatalf("parse decoy: %v", err)
	}
	found, err := parsed.ServerName()
	if err != nil {
		t.Fatalf("name of decoy: %v", err)
	}

	return found.Host
}

func TestFakePrependsADoomedDecoy(t *testing.T) {
	host := "gateway.discord.gg"
	hello := helloWith(host)

	p, err := FromRule(hello, rules.Rule{Decoy: "auto"})
	if err != nil {
		t.Fatalf("fake: %v", err)
	}
	if len(p.Segments) != 2 {
		t.Fatalf("segments = %d, want 2", len(p.Segments))
	}
	if p.Segments[0].TTL != decoyTTL {
		t.Errorf("decoy TTL = %d, want %d", p.Segments[0].TTL, decoyTTL)
	}
	if parseName(t, p.Segments[0].Bytes) == host {
		t.Error("decoy carries the real host")
	}
	if p.Segments[1].TTL != 0 {
		t.Errorf("real TTL = %d, want 0", p.Segments[1].TTL)
	}
	if !bytes.Equal(p.Segments[1].Bytes, hello) {
		t.Error("real segment differs from hello")
	}
}

func TestFakeThenCutGivesDecoyAndTwoReal(t *testing.T) {
	hello := helloWith("gateway.discord.gg")

	p, err := FromRule(hello, rules.Rule{Decoy: "auto", Cut: "name"})
	if err != nil {
		t.Fatalf("fake+cut: %v", err)
	}
	if len(p.Segments) != 3 {
		t.Fatalf("segments = %d, want 3", len(p.Segments))
	}

	joined := append(append([]byte{}, p.Segments[1].Bytes...), p.Segments[2].Bytes...)
	if !bytes.Equal(joined, hello) {
		t.Error("the two real segments do not rejoin the hello")
	}
}

func TestFakeUsesAGivenDecoyName(t *testing.T) {
	hello := helloWith("gateway.discord.gg")

	p, err := FromRule(hello, rules.Rule{Decoy: "ya.ru"})
	if err != nil {
		t.Fatalf("fake: %v", err)
	}
	if got := parseName(t, p.Segments[0].Bytes); got != "ya.ru" {
		t.Errorf("decoy name = %q, want ya.ru", got)
	}
}

func TestFakeHonoursAGivenTTL(t *testing.T) {
	hello := helloWith("gateway.discord.gg")

	p, err := FromRule(hello, rules.Rule{Decoy: "auto", TTL: 4})
	if err != nil {
		t.Fatalf("fake: %v", err)
	}
	if p.Segments[0].TTL != 4 {
		t.Errorf("decoy TTL = %d, want 4", p.Segments[0].TTL)
	}
}

func TestDecoyNameKeepsTheLength(t *testing.T) {
	for _, host := range []string{"gateway.discord.gg", "a.co", "example.com", "vk.com"} {
		if got := decoyName(host); len(got) != len(host) {
			t.Errorf("decoyName(%q) = %q, len %d, want %d", host, got, len(got), len(host))
		}
	}
}

func TestDisorderDoomsTheFirstSegment(t *testing.T) {
	hello := helloWith("gateway.discord.gg")

	p, err := FromRule(hello, rules.Rule{Cut: "name", Disorder: true})
	if err != nil {
		t.Fatalf("disorder: %v", err)
	}
	if len(p.Segments) != 2 {
		t.Fatalf("segments = %d, want 2", len(p.Segments))
	}
	if p.Segments[0].TTL != 1 {
		t.Errorf("first TTL = %d, want 1", p.Segments[0].TTL)
	}
	if p.Segments[1].TTL != 0 {
		t.Errorf("second TTL = %d, want 0", p.Segments[1].TTL)
	}
	joined := append(append([]byte{}, p.Segments[0].Bytes...), p.Segments[1].Bytes...)
	if !bytes.Equal(joined, hello) {
		t.Error("disorder changed the bytes")
	}
}

func TestCutWithoutDisorderIsNormalTTL(t *testing.T) {
	hello := helloWith("gateway.discord.gg")

	p, err := FromRule(hello, rules.Rule{Cut: "name"})
	if err != nil {
		t.Fatalf("cut: %v", err)
	}
	if p.Segments[0].TTL != 0 {
		t.Errorf("first TTL = %d, want 0 (no disorder)", p.Segments[0].TTL)
	}
}

func TestRawWaysAreRejected(t *testing.T) {
	hello := helloWith("gateway.discord.gg")

	raw := []rules.Rule{
		{BadSeq: 100000},
		{BadSum: true},
		{Overlap: 1},
		{Signed: true},
		{HostFake: "mail.ru"},
		{Recorded: true},
		{Stale: 1 << 30},
		{BadAck: -66000},
	}
	for _, rule := range raw {
		if _, err := FromRule(hello, rule); !errors.Is(err, ErrNotOnSocket) {
			t.Errorf("%+v: err = %v, want ErrNotOnSocket", rule, err)
		}
	}
}

func TestRawWayIsRejectedEvenBesideACut(t *testing.T) {
	hello := helloWith("gateway.discord.gg")

	if _, err := FromRule(hello, rules.Rule{Cut: "name", BadSeq: 100000}); !errors.Is(err, ErrNotOnSocket) {
		t.Errorf("cut+badseq: err = %v, want ErrNotOnSocket", err)
	}
}

func TestSocketRuleStillPlans(t *testing.T) {
	hello := helloWith("gateway.discord.gg")

	if _, err := FromRule(hello, rules.Rule{Decoy: "auto", TTL: 4, Cut: "name"}); err != nil {
		t.Errorf("socket rule: %v", err)
	}
}
