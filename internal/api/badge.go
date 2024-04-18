package api

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-faster/errors"

	"github.com/go-faster/bot/internal/oas"
)

func toBadgeStr(v string) string {
	v = strings.ReplaceAll(v, " ", "_")
	return v
}

func generateBadgePath(name, text, style string) string {
	return "/badge/" + strings.Join([]string{
		toBadgeStr(name),
		toBadgeStr(text),
		style,
	}, "-")
}

func (s Server) GetTelegramBadge(ctx context.Context, params oas.GetTelegramBadgeParams) (*oas.GetTelegramBadgeOKHeaders, error) {
	members := map[string]int{
		"gotd_en":   237,
		"gotd_ru":   234,
		"gotd_zhcn": 15,
	}[params.GroupName]
	_ = s.tg // TODO(ernado): fetch actual data.
	title := params.Title.Or(params.GroupName)
	text := strconv.Itoa(members)
	u := &url.URL{
		Scheme: "https",
		Host:   "img.shields.io",
		Path:   generateBadgePath(title, text, "179cde"),
	}
	q := u.Query()
	q.Set("logo", "telegram")
	u.RawQuery = q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), http.NoBody)
	if err != nil {
		return nil, errors.Wrap(err, "create request")
	}
	res, err := s.ht.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "send request")
	}
	defer func() { _ = res.Body.Close() }()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "read body")
	}

	return &oas.GetTelegramBadgeOKHeaders{
		CacheControl: oas.NewOptString("no-cache"),
		Response: oas.GetTelegramBadgeOK{
			Data: bytes.NewReader(data),
		},
	}, nil
}
