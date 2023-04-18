package gh

import (
	"net/url"
	"strings"

	"github.com/go-faster/errors"
	"github.com/go-faster/jx"
)

type Event struct {
	Type     string
	RepoName string
	RepoID   int64
}

func htmlURL(s string) string {
	u, err := url.Parse(s)
	if err != nil {
		return s
	}
	u.Host = "github.com"
	secondSlash := strings.Index(u.Path[1:], "/")
	if secondSlash == -1 {
		return s
	}
	u.Path = u.Path[secondSlash+1:]
	return u.String()
}

func Transform(d *jx.Decoder, e *jx.Encoder) (*Event, error) {
	var (
		repoID       int64
		fullRepoName string
		repoURL      string
		evType       string
	)
	e.ObjStart()
	if err := d.ObjBytes(func(d *jx.Decoder, key []byte) error {
		var err error
		switch string(key) {
		case "actor":
			e.FieldStart("sender")
			e.ObjStart()
			if err := d.ObjBytes(func(d *jx.Decoder, key []byte) error {
				switch string(key) {
				case "url":
					s, err := d.Str()
					if err != nil {
						return errors.Wrap(err, "url")
					}
					e.Field("url", func(e *jx.Encoder) {
						e.Str(s)
					})
					e.Field("html_url", func(e *jx.Encoder) {
						e.Str(htmlURL(s))
					})
					return nil
				default:
					v, err := d.Raw()
					if err != nil {
						return errors.Wrap(err, "actor")
					}
					e.Field(string(key), func(e *jx.Encoder) {
						e.Raw(v)
					})
					return nil
				}
			}); err != nil {
				return err
			}
			e.ObjEnd()
			return nil
		case "payload":
			return d.ObjBytes(func(d *jx.Decoder, key []byte) error {
				v, err := d.Raw()
				if err != nil {
					return errors.Wrap(err, "payload")
				}
				e.Field(string(key), func(e *jx.Encoder) {
					e.Raw(v)
				})
				return nil
			})
		case "type":
			if evType, err = d.Str(); err != nil {
				return errors.Wrap(err, "type")
			}
			return nil
		case "repo":
			return d.ObjBytes(func(d *jx.Decoder, key []byte) error {
				switch string(key) {
				case "id":
					if repoID, err = d.Int64(); err != nil {
						return errors.Wrap(err, "id")
					}
					return nil
				case "name":
					if fullRepoName, err = d.Str(); err != nil {
						return errors.Wrap(err, "name")
					}
					return nil
				case "url":
					if repoURL, err = d.Str(); err != nil {
						return errors.Wrap(err, "url")
					}
					return nil
				default:
					return d.Skip()
				}
			})
		default:
			return d.Skip()
		}
	}); err != nil {
		return nil, errors.Wrap(err, "decode")
	}
	e.Field("repository", func(e *jx.Encoder) {
		e.Obj(func(e *jx.Encoder) {
			e.Field("id", func(e *jx.Encoder) {
				e.Int64(repoID)
			})
			e.Field("name", func(e *jx.Encoder) {
				// Strip first part of repo name
				_, name, ok := strings.Cut(fullRepoName, "/")
				if !ok {
					name = fullRepoName
				}
				e.Str(name)
			})
			e.Field("full_name", func(e *jx.Encoder) {
				e.Str(fullRepoName)
			})
			e.Field("url", func(e *jx.Encoder) {
				e.Str(repoURL)
			})
			e.Field("html_url", func(e *jx.Encoder) {
				e.Str(htmlURL(repoURL))
			})
		})
	})
	e.ObjEnd()
	d.ResetBytes(e.Bytes())

	return &Event{
		Type:     evType,
		RepoName: fullRepoName,
		RepoID:   repoID,
	}, nil
}
