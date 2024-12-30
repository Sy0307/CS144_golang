package reassembler

import (
	"testing"
)

func TestReassembler_Basic(t *testing.T) {
	r := NewReassembler(1000)
	segs := []TCPSegment{
		{Seq: 1000, Data: []byte("Hello "), Length: 6},
		{Seq: 1006, Data: []byte("World"), Length: 5},
		{Seq: 1011, Data: []byte("!"), Length: 1},
	}

	r.AddSegment(segs[2])
	r.AddSegment(segs[0])
	r.AddSegment(segs[1])
	r.MergeSegments()

	got := string(r.GetByteStream())
	want := "Hello World!"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestReassembler_Overlap(t *testing.T) {
	r := NewReassembler(1000)
	// Overlapping segments
	segs := []TCPSegment{
		{Seq: 1000, Data: []byte("Hello"), Length: 5},
		{Seq: 1003, Data: []byte("loWorld"), Length: 7}, // Overlap on "lo"
	}

	for _, s := range segs {
		r.AddSegment(s)
	}
	r.MergeSegments()

	got := string(r.GetByteStream())
	want := "HelloWorld"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestReassembler_Gap(t *testing.T) {
	r := NewReassembler(1000)
	// Missing segment in the middle
	segs := []TCPSegment{
		{Seq: 1000, Data: []byte("Hel"), Length: 3},
		{Seq: 1006, Data: []byte("World"), Length: 5}, // gap between 1003~1006
	}

	for _, s := range segs {
		r.AddSegment(s)
	}
	r.MergeSegments()

	got := string(r.GetByteStream())
	want := "Hel"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	// Fill gap
	r.AddSegment(TCPSegment{Seq: 1003, Data: []byte("lo "), Length: 3})
	r.MergeSegments()

	got2 := string(r.GetByteStream())
	want2 := "Hello World"
	if got2 != want2 {
		t.Errorf("got %q, want %q", got2, want2)
	}
}
