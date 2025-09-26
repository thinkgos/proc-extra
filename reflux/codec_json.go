package reflux

import "encoding/json"

var _ Codec = (*CodecJSON)(nil)

type CodecJSON struct{}

// Marshal implements Codec.
func (CodecJSON) Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

// Unmarshal implements Codec.
func (CodecJSON) Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}
