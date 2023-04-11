package main

import (
	"context"
	"log"

	"github.com/go-faster/bot/internal/ent"

	"entgo.io/ent/dialect/sql/schema"
)

func main() {
	client, err := ent.Open("postgresql", "root:pass@tcp(localhost:3306)/test")
	if err != nil {
		log.Fatalf("failed connecting to mysql: %v", err)
	}
	defer client.Close()
	ctx := context.Background()
	// Run migration.
	err = client.Schema.Create(ctx, schema.WithAtlas(true))
	if err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}
}
