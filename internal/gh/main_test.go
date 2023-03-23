package gh

import (
	"os"
	"testing"

	"github.com/go-faster/bot/internal/gold"
)

func TestMain(m *testing.M) {
	// Explicitly registering flags for golden files.
	gold.Init()

	os.Exit(m.Run())
}
