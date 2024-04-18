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

func (s Server) GetTelegramBadge(ctx context.Context, params oas.GetTelegramBadgeParams) (*oas.GetTelegramBadgeOKHeaders, error) {
	var members int
	{
		peer, err := s.resolver.ResolveDomain(ctx, params.GroupName)
		if err != nil {
			return nil, errors.Wrap(err, "resolve domain")
		}
		var inputChannel tg.InputChannel
		inputChannel.FillFrom(peer.(*tg.InputPeerChannel))
		fullChat, err := s.tg.API().ChannelsGetFullChannel(ctx, &inputChannel)
		if err != nil {
			return nil, errors.Wrap(err, "get chat")
		}
		s.lg.Info("Got chat",
			zap.Int("chats", len(fullChat.Chats)),
			zap.Int64("id", fullChat.FullChat.(*tg.ChatFull).ID),
			zap.String("about", fullChat.FullChat.(*tg.ChatFull).About),
		)
		for _, chat := range fullChat.Chats {
			switch c := chat.(type) {
			case *tg.Chat:
				members = c.ParticipantsCount
			case *tg.Channel:
				members = c.ParticipantsCount
			default:
				s.lg.Warn("unexpected chat type",
					zap.String("type", fmt.Sprintf("%T", chat)),
				)
			}
		}
	}
	var (
		title = params.Title.Or(params.GroupName)
		text  = strconv.Itoa(members)
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
		ETag:         oas.NewOptString(etag(params.GroupName, data)),
		Response: oas.GetTelegramBadgeOK{
			Data: bytes.NewReader(data),
		},
	}, nil
}
