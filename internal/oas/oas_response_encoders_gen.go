// Code generated by ogen, DO NOT EDIT.

package oas

import (
	"io"
	"net/http"

	"github.com/go-faster/errors"
	"github.com/go-faster/jx"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/ogen-go/ogen/conv"
	ht "github.com/ogen-go/ogen/http"
	"github.com/ogen-go/ogen/uri"
)

func encodeGetTelegramBadgeResponse(response *SVGHeaders, w http.ResponseWriter, span trace.Span) error {
	w.Header().Set("Content-Type", "image/svg+xml")
	// Encoding response headers.
	{
		h := uri.NewHeaderEncoder(w.Header())
		// Encode "Cache-Control" header.
		{
			cfg := uri.HeaderParameterEncodingConfig{
				Name:    "Cache-Control",
				Explode: false,
			}
			if err := h.EncodeParam(cfg, func(e uri.Encoder) error {
				if val, ok := response.CacheControl.Get(); ok {
					return e.EncodeValue(conv.StringToString(val))
				}
				return nil
			}); err != nil {
				return errors.Wrap(err, "encode Cache-Control header")
			}
		}
		// Encode "ETag" header.
		{
			cfg := uri.HeaderParameterEncodingConfig{
				Name:    "ETag",
				Explode: false,
			}
			if err := h.EncodeParam(cfg, func(e uri.Encoder) error {
				if val, ok := response.ETag.Get(); ok {
					return e.EncodeValue(conv.StringToString(val))
				}
				return nil
			}); err != nil {
				return errors.Wrap(err, "encode ETag header")
			}
		}
	}
	w.WriteHeader(200)
	span.SetStatus(codes.Ok, http.StatusText(200))

	writer := w
	if closer, ok := response.Response.Data.(io.Closer); ok {
		defer closer.Close()
	}
	if _, err := io.Copy(writer, response.Response); err != nil {
		return errors.Wrap(err, "write")
	}

	return nil
}

func encodeGetTelegramOnlineBadgeResponse(response *SVGHeaders, w http.ResponseWriter, span trace.Span) error {
	w.Header().Set("Content-Type", "image/svg+xml")
	// Encoding response headers.
	{
		h := uri.NewHeaderEncoder(w.Header())
		// Encode "Cache-Control" header.
		{
			cfg := uri.HeaderParameterEncodingConfig{
				Name:    "Cache-Control",
				Explode: false,
			}
			if err := h.EncodeParam(cfg, func(e uri.Encoder) error {
				if val, ok := response.CacheControl.Get(); ok {
					return e.EncodeValue(conv.StringToString(val))
				}
				return nil
			}); err != nil {
				return errors.Wrap(err, "encode Cache-Control header")
			}
		}
		// Encode "ETag" header.
		{
			cfg := uri.HeaderParameterEncodingConfig{
				Name:    "ETag",
				Explode: false,
			}
			if err := h.EncodeParam(cfg, func(e uri.Encoder) error {
				if val, ok := response.ETag.Get(); ok {
					return e.EncodeValue(conv.StringToString(val))
				}
				return nil
			}); err != nil {
				return errors.Wrap(err, "encode ETag header")
			}
		}
	}
	w.WriteHeader(200)
	span.SetStatus(codes.Ok, http.StatusText(200))

	writer := w
	if closer, ok := response.Response.Data.(io.Closer); ok {
		defer closer.Close()
	}
	if _, err := io.Copy(writer, response.Response); err != nil {
		return errors.Wrap(err, "write")
	}

	return nil
}

func encodeGithubStatusResponse(response *GithubStatusOK, w http.ResponseWriter, span trace.Span) error {
	w.WriteHeader(200)
	span.SetStatus(codes.Ok, http.StatusText(200))

	return nil
}

func encodeStatusResponse(response *Status, w http.ResponseWriter, span trace.Span) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	span.SetStatus(codes.Ok, http.StatusText(200))

	e := new(jx.Encoder)
	response.Encode(e)
	if _, err := e.WriteTo(w); err != nil {
		return errors.Wrap(err, "write")
	}

	return nil
}

func encodeErrorResponse(response *ErrorStatusCode, w http.ResponseWriter, span trace.Span) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	code := response.StatusCode
	if code == 0 {
		// Set default status code.
		code = http.StatusOK
	}
	w.WriteHeader(code)
	if st := http.StatusText(code); code >= http.StatusBadRequest {
		span.SetStatus(codes.Error, st)
	} else {
		span.SetStatus(codes.Ok, st)
	}

	e := new(jx.Encoder)
	response.Response.Encode(e)
	if _, err := e.WriteTo(w); err != nil {
		return errors.Wrap(err, "write")
	}

	if code >= http.StatusInternalServerError {
		return errors.Wrapf(ht.ErrInternalServerErrorResponse, "code: %d, message: %s", code, http.StatusText(code))
	}
	return nil

}
