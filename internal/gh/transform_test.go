package gh

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-faster/jx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"

	"github.com/go-faster/sdk/gold"
)

func TestTransform(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("_testdata", "event.json"))
	require.NoErrorf(t, err, "read event.json")

	var (
		e = jx.GetEncoder()
		d = jx.GetDecoder()
		h = &Webhook{
			tracer: trace.NewNoopTracerProvider().Tracer("test"),
		}
	)

	e.Reset()
	e.SetIdent(2)
	d.ResetBytes(data)
	v, err := h.Transform(context.Background(), d, e)
	require.NoErrorf(t, err, "transform")
	assert.Equal(t, &Event{Type: "IssuesEvent", RepoName: "ernado/oss-estimator", RepoID: 610784405}, v)

	gold.Str(t, gold.NormalizeNewlines(e.String()), "event_wh.json")
}
