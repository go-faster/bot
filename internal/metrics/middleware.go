package metrics

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"

	"github.com/go-faster/errors"
	"github.com/gotd/td/fileid"
	"go.uber.org/zap"

	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/tg"

	"github.com/go-faster/bot/internal/botapi"
	"github.com/go-faster/bot/internal/dispatch"
)

type Middleware struct {
	next       dispatch.MessageHandler
	downloader *downloader.Downloader
	client     *botapi.Client
	metrics    Metrics

	logger *zap.Logger
}

// NewMiddleware creates new metrics middleware
func NewMiddleware(
	next dispatch.MessageHandler,
	d *downloader.Downloader,
	metrics Metrics,
	opts MiddlewareOptions,
) Middleware {
	opts.setDefaults()
	return Middleware{
		next:       next,
		downloader: d,
		client:     opts.BotAPI,
		metrics:    metrics,
		logger:     opts.Logger,
	}
}

func maxSize(sizes []tg.PhotoSizeClass) string {
	var (
		maxSize string
		maxH    int
	)

	for _, size := range sizes {
		if s, ok := size.(interface {
			GetH() int
			GetType() string
		}); ok && s.GetH() > maxH {
			maxH = s.GetH()
			maxSize = s.GetType()
		}
	}

	return maxSize
}

func (m Middleware) downloadMedia(ctx context.Context, rpc *tg.Client, loc tg.InputFileLocationClass) error {
	h := sha256.New()
	w := &metricWriter{
		Increase: m.metrics.MediaBytes.Add,
	}

	if _, err := m.downloader.Download(rpc, loc).
		Stream(ctx, io.MultiWriter(h, w)); err != nil {
		return errors.Wrap(err, "stream")
	}

	m.logger.Info("Downloaded media",
		zap.Int64("bytes", w.Bytes),
		zap.String("sha256", fmt.Sprintf("%x", h.Sum(nil))),
	)

	return nil
}

func (m Middleware) handleMedia(ctx context.Context, rpc *tg.Client, msg *tg.Message) error {
	log := m.logger.With(zap.Int("msg_id", msg.ID), zap.Stringer("peer_id", msg.PeerID))

	switch media := msg.Media.(type) {
	case *tg.MessageMediaDocument:
		doc, ok := media.Document.AsNotEmpty()
		if !ok {
			return nil
		}
		if err := m.downloadMedia(ctx, rpc, &tg.InputDocumentFileLocation{
			ID:            doc.ID,
			AccessHash:    doc.AccessHash,
			FileReference: doc.FileReference,
		}); err != nil {
			return errors.Wrap(err, "download")
		}

		if err := m.checkOurFileID(ctx, fileid.FromDocument(doc)); err != nil {
			log.Warn("Test document FileID", zap.Error(err))
		}

	case *tg.MessageMediaPhoto:
		p, ok := media.Photo.AsNotEmpty()
		if !ok {
			return nil
		}
		size := maxSize(p.Sizes)
		if err := m.downloadMedia(ctx, rpc, &tg.InputPhotoFileLocation{
			ID:            p.ID,
			AccessHash:    p.AccessHash,
			FileReference: p.FileReference,
			ThumbSize:     size,
		}); err != nil {
			return errors.Wrap(err, "download")
		}

		thumbType := 'x'
		if len(size) >= 1 {
			thumbType = rune(size[0])
		}

		if err := m.checkOurFileID(ctx, fileid.FromPhoto(p, thumbType)); err != nil {
			log.Warn("Test photo FileID", zap.Error(err))
		}
	default:
		// Do not try to get file_id from messages without attachments.
		return nil
	}

	loc, fileID, err := m.tryGetFileID(ctx, msg.ID)
	if err != nil {
		log.Warn("Parse file_id",
			zap.String("file_id", fileID),
			zap.Error(err),
		)
	} else {
		if _, err := m.downloader.Download(rpc, loc).Stream(ctx, io.Discard); err != nil {
			log.Warn("Download file_id",
				zap.String("file_id", fileID),
				zap.Error(err),
			)
		}
		log.Info("Successfully downloaded file_id", zap.String("file_id", fileID))
	}

	return nil
}

// OnMessage implements dispatch.MessageHandler.
func (m Middleware) OnMessage(ctx context.Context, e dispatch.MessageEvent) error {
	m.metrics.Messages.Inc()

	if err := m.next.OnMessage(ctx, e); err != nil {
		return err
	}

	if err := m.handleMedia(ctx, e.RPC(), e.Message); err != nil {
		return errors.Wrap(err, "handle media")
	}

	m.metrics.Responses.Inc()
	return nil
}
