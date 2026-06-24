package sources

import (
	"context"
	"testing"
)

func TestHNFetch(t *testing.T) {
	ctx := context.Background()
	hn := NewHN(5)

	comments, err := hn.Fetch(ctx, "iphone 16")
	if err != nil {
		t.Fatal(err)
	}

	if len(comments) == 0 {
		t.Fatal("expected comments")
	}

	body := comments[0].Body
	if len(body) > 100 {
		body = body[:100]
	}
	t.Logf("got %d comments, first: %s", len(comments), body)

}
