package main

import (
	"context"

	"github.com/impactj90/lofi-tracker/cmd/internal/afk"
)

func main() {
	context := context.Background()
	afk.Run(context)
}
