package service

import (
	"bytes"
	"testing"
)

func TestBitWriterAndReaderParity(t *testing.T) {
	var buf bytes.Buffer
	bw := NewBitWriter(&buf)

	inputs := []struct {
		val   uint32
		nBits uint8
	}{
		{5, 3},
		{23, 5},
		{511, 10},
		{1, 1},
		{12, 4},
	}

	for _, in := range inputs {
		if err := bw.WriteBits(in.val, in.nBits); err != nil {
			t.Fatalf("failed to write bits: %v", err)
		}
	}
	if err := bw.Flush(); err != nil {
		t.Fatalf("failed to flush: %v", err)
	}

	br := NewBitReader(&buf)
	for idx, in := range inputs {
		got, err := br.ReadBits(in.nBits)
		if err != nil {
			t.Fatalf("[%d] read error: %v", idx, err)
		}
		if got != in.val {
			t.Errorf("[%d] mismatch: got %d, expected %d", idx, got, in.val)
		}
	}
}
