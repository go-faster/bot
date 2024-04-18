package api

import (
	"bytes"
	"context"
	"fmt"
	"hash/crc32"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-faster/errors"
	"github.com/gotd/td/tg"
	"go.uber.org/zap"

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

func etag(name string, data []byte) string {
	crc := crc32.ChecksumIEEE(data)
	return fmt.Sprintf(`W/"%s-%d-%08X"`, name, len(data), crc)
}

func (s Server) download(ctx context.Context, u *url.URL) (*oas.SVGHeaders, error) {
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

	return &oas.SVGHeaders{
		CacheControl: oas.NewOptString("no-cache"),
		ETag:         oas.NewOptString(etag("svg", data)),
		Response: oas.SVG{
			Data: bytes.NewReader(data),
		},
	}, nil
}

func (s Server) fetchChannel(ctx context.Context, name string) (*tg.ChannelFull, error) {
	peer, err := s.resolver.ResolveDomain(ctx, name)
	if err != nil {
		return nil, errors.Wrap(err, "resolve domain")
	}
	var inputChannel tg.InputChannel
	inputChannel.FillFrom(peer.(*tg.InputPeerChannel))
	full, err := s.tg.API().ChannelsGetFullChannel(ctx, &inputChannel)
	if err != nil {
		return nil, errors.Wrap(err, "get chat")
	}
	v := full.FullChat.(*tg.ChannelFull)
	s.lg.Info("Got chat",
		zap.String("name", name),
		zap.Int64("id", v.ID),
		zap.Int("participants", v.ParticipantsCount),
		zap.Int("online", v.OnlineCount),
	)
	return v, nil
}

func (s Server) GetTelegramOnlineBadge(ctx context.Context, params oas.GetTelegramOnlineBadgeParams) (*oas.SVGHeaders, error) {
	var count int
	for _, name := range params.Groups {
		full, err := s.fetchChannel(ctx, name)
		if err != nil {
			return nil, errors.Wrap(err, "get chat")
		}
		count += full.OnlineCount
	}
	var (
		text = strconv.Itoa(count)
		u    = &url.URL{
			Scheme: "https",
			Host:   "img.shields.io",
			Path:   generateBadgePath("online", text, "green"),
		}
	)
	{
		q := u.Query()
		q.Set("logo", "telegram")
		u.RawQuery = q.Encode()
	}
	return s.download(ctx, u)
}

func (s Server) GetTelegramBadge(ctx context.Context, params oas.GetTelegramBadgeParams) (*oas.SVGHeaders, error) {
	channel, err := s.fetchChannel(ctx, params.GroupName)
	if err != nil {
		return nil, errors.Wrap(err, "get chat")
	}
	var (
		title = params.Title.Or(params.GroupName)
		text  = strconv.Itoa(channel.ParticipantsCount)
		u     = &url.URL{
			Scheme: "https",
			Host:   "img.shields.io",
			Path:   generateBadgePath(title, text, "179cde"),
		}
	)
	{
		q := u.Query()
		q.Set("logo", "telegram")
		u.RawQuery = q.Encode()
	}
	return s.download(ctx, u)
}
