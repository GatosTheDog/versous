package sources

import (
	"context"

	"github.com/GatosTheDog/versous/internal/store"
)

type CommentSource interface {
	Fetch(ctx context.Context, product string) ([]store.Comment, error)
	Name() string
}
