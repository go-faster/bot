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

func (s Server) GetTelegramGoTDBadge(ctx context.Context) (oas.GetTelegramGoTDBadgeOK, error) {
	members := 236 + 234 + 15
	_ = s.tg // TODO(ernado): fetch actual data.
	u := &url.URL{
		Scheme: "https",
		Host:   "img.shields.io",
		Path:   generateBadgePath(strconv.Itoa(members), "members", "179cde"),
	}
	q := u.Query()
	q.Set("logo", "telegram")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), http.NoBody)
	if err != nil {
		return oas.GetTelegramGoTDBadgeOK{}, errors.Wrap(err, "create request")
	}
	res, err := s.ht.Do(req)
	if err != nil {
		return oas.GetTelegramGoTDBadgeOK{}, errors.Wrap(err, "send request")
	}
	defer func() { _ = res.Body.Close() }()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return oas.GetTelegramGoTDBadgeOK{}, errors.Wrap(err, "read body")
	}
	return oas.GetTelegramGoTDBadgeOK{Data: bytes.NewReader(data)}, nil
}
