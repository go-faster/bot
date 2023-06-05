package api

import (
	"context"
	"time"

	"github.com/go-faster/errors"

	"github.com/go-faster/bot/internal/ent/gitcommit"
	"github.com/go-faster/bot/internal/oas"
)

func (s *Server) Status(ctx context.Context) (*oas.Status, error) {
	// Last week.
	until := time.Now().AddDate(0, 0, -7)
	totalCommits, err := s.db.GitCommit.Query().
		Where(gitcommit.DateGTE(until)).
		Count(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "count commits")
	}
	return &oas.Status{
		Message: "Weekly stats:",
		Stat: oas.Statistics{
			TotalCommits: totalCommits,
		},
	}, nil
}
