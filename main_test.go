package comb

import (
	"testing"

	"github.com/google/uuid"
)

func putBytesSafe(b []byte, n int, v uint64) {
	_ = b[5] // early bounds check to guarantee safety of writes below
	b[0] = byte(v >> 40)
	b[1] = byte(v >> 32)
	b[2] = byte(v >> 24)
	b[3] = byte(v >> 16)
	b[4] = byte(v >> 8)
	b[5] = byte(v)
}

func fillUint64Safe(b []byte) uint64 {
	_ = b[7] // bounds check hint to compiler; see golang.org/issue/14808
	return uint64(b[7]) | uint64(b[6])<<8 | uint64(b[5])<<16 | uint64(b[4])<<24 |
		uint64(b[3])<<32 | uint64(b[2])<<40 | uint64(b[1])<<48 | uint64(b[0])<<56
}

func TestUint64ToBytes(t *testing.T) {
	id := uuid.Nil
	uint64ToBytes(id[10:], 6, 0xffffffffffff)
	str := "00000000-0000-0000-0000-ffffffffffff"
	if id.String() != str {
		t.Errorf("want %q got %q", str, id.String())
	}
	id = uuid.Nil
	putBytesSafe(id[10:], 6, 0xffffffffffff)
	if id.String() != str {
		t.Errorf("want %q got %q", str, id.String())
	}
}

func TestBytesToUint64(t *testing.T) {
	id, err := uuid.Parse("00000000-0000-0000-0000-ffffffffffff")
	if err != nil {
		t.Error("did not expect an error:", err)
	}
	byt := make([]byte, 8)
	copy(byt[2:], id[10:])
	i := fillUint64Safe(byt)
	j := uint64(0xffffffffffff)
	if i != j {
		t.Errorf("want %d got %d", j, i)
	}

	i = bytesToUint64(id[10:], 6)
	if i != j {
		t.Errorf("want %b got %b", j, i)
	}
}
