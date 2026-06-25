package sources

import (
	"context"
	"os"
	"testing"
)

func TestYoutubeFetch(t *testing.T) {
	if os.Getenv("YOUTUBE_API_KEY") == "" {
		t.Skip("missing env vars")
	}

	ctx := context.Background()
	yt := NewYoutube(5)

	comments, err := yt.Fetch(ctx, "iphone 16 battery")
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
