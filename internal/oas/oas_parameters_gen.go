// Code generated by ogen, DO NOT EDIT.

package oas

import (
	"net/http"
	"net/url"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/conv"
	"github.com/ogen-go/ogen/middleware"
	"github.com/ogen-go/ogen/ogenerrors"
	"github.com/ogen-go/ogen/uri"
	"github.com/ogen-go/ogen/validate"
)

// GetTelegramBadgeParams is parameters of getTelegramBadge operation.
type GetTelegramBadgeParams struct {
	Title     OptString
	GroupName string
}

func unpackGetTelegramBadgeParams(packed middleware.Parameters) (params GetTelegramBadgeParams) {
	{
		key := middleware.ParameterKey{
			Name: "title",
			In:   "query",
		}
		if v, ok := packed[key]; ok {
			params.Title = v.(OptString)
		}
	}
	{
		key := middleware.ParameterKey{
			Name: "group_name",
			In:   "path",
		}
		params.GroupName = packed[key].(string)
	}
	return params
}

func decodeGetTelegramBadgeParams(args [1]string, argsEscaped bool, r *http.Request) (params GetTelegramBadgeParams, _ error) {
	q := uri.NewQueryDecoder(r.URL.Query())
	// Decode query: title.
	if err := func() error {
		cfg := uri.QueryParameterDecodingConfig{
			Name:    "title",
			Style:   uri.QueryStyleForm,
			Explode: true,
		}

		if err := q.HasParam(cfg); err == nil {
			if err := q.DecodeParam(cfg, func(d uri.Decoder) error {
				var paramsDotTitleVal string
				if err := func() error {
					val, err := d.DecodeValue()
					if err != nil {
						return err
					}

					c, err := conv.ToString(val)
					if err != nil {
						return err
					}

					paramsDotTitleVal = c
					return nil
				}(); err != nil {
					return err
				}
				params.Title.SetTo(paramsDotTitleVal)
				return nil
			}); err != nil {
				return err
			}
		}
		return nil
	}(); err != nil {
		return params, &ogenerrors.DecodeParamError{
			Name: "title",
			In:   "query",
			Err:  err,
		}
	}
	// Decode path: group_name.
	if err := func() error {
		param := args[0]
		if argsEscaped {
			unescaped, err := url.PathUnescape(args[0])
			if err != nil {
				return errors.Wrap(err, "unescape path")
			}
			param = unescaped
		}
		if len(param) > 0 {
			d := uri.NewPathDecoder(uri.PathDecoderConfig{
				Param:   "group_name",
				Value:   param,
				Style:   uri.PathStyleSimple,
				Explode: false,
			})

			if err := func() error {
				val, err := d.DecodeValue()
				if err != nil {
					return err
				}

				c, err := conv.ToString(val)
				if err != nil {
					return err
				}

				params.GroupName = c
				return nil
			}(); err != nil {
				return err
			}
		} else {
			return validate.ErrFieldRequired
		}
		return nil
	}(); err != nil {
		return params, &ogenerrors.DecodeParamError{
			Name: "group_name",
			In:   "path",
			Err:  err,
		}
	}
	return params, nil
}