package hash_id

import (
	"fmt"

	"github.com/sqids/sqids-go"
	"github.com/thinkgos/proc/luhn"
)

type HashId struct {
	luhn luhn.LuhnModN
	sqid *sqids.Sqids
}

func NewHashId(alphabet string) (*HashId, error) {
	sqid, err := sqids.New(sqids.Options{
		Alphabet: alphabet,
	})
	if err != nil {
		return nil, err
	}
	luhn, err := luhn.New(alphabet)
	if err != nil {
		return nil, err
	}
	return &HashId{
		luhn: luhn,
		sqid: sqid,
	}, nil
}

func (h *HashId) Encode(v uint64) (string, error) {
	id, err := h.sqid.Encode([]uint64{v})
	if err != nil {
		return "", err
	}
	return h.luhn.Encode(id)
}
func (h *HashId) Decode(v string) (uint64, error) {
	r, err := h.luhn.Decode(v)
	if err != nil {
		return 0, fmt.Errorf("invalid id: %s", v)
	}
	numbers := h.sqid.Decode(r)
	if len(numbers) != 1 {
		return 0, fmt.Errorf("invalid id: %s", v)
	}
	return numbers[0], nil
}
