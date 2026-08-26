package hash_id

import (
	"math/rand/v2"
	"testing"
)

func Test_HashId(t *testing.T) {
	h, err := NewHashId("0123456789abcdefghijklmnopqrstuvwxyz")
	if err != nil {
		t.Fatal(err)
	}
	wantId := rand.Uint64()
	encoded, err := h.Encode(wantId)
	if err != nil {
		t.Fatal(err)
	}
	gotId, err := h.Decode(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if gotId != wantId {
		t.Fatalf("got value(%v) does not match want id(%v)", gotId, wantId)
	}
}
