package api

import (
	"context"

	"github.com/go-faster/errors"

	"github.com/go-faster/bot/internal/oas"
)

func (s *Server) Status(ctx context.Context) (*oas.Status, error) {
	totalCommits, err := s.db.GitCommit.Query().Count(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "count commits")
	}
	return &oas.Status{
		Message: "All systems operational",
		Stat: oas.Statistics{
			TotalCommits: totalCommits,
		},
	}, nil
}
