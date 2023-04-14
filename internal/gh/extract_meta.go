package gh

import "github.com/go-faster/jx"

type eventMeta struct {
	Organization string // go-faster
	Repository   string // bot
}

func extractEventMeta(raw []byte) (*eventMeta, error) {
	d := jx.GetDecoder()
	defer jx.PutDecoder(d)

	d.ResetBytes(raw)

	var m eventMeta
	if err := d.ObjBytes(func(d *jx.Decoder, key []byte) error {
		return d.Skip()
	}); err != nil {
		return nil, err
	}

	return &m, nil
}
