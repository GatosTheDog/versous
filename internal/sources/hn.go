package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html"

	"github.com/GatosTheDog/versous/internal/store"
)

const hnSearchURL = "https://hn.algolia.com/api/v1/search"

type HN struct {
	client *http.Client
	limit  int
}

type hnResponse struct {
	Hits []hnHit `json:"hits"`
}

type hnHit struct {
	ObjectID    string `json:"objectID"`
	CommentText string `json:"comment_text"`
	Author      string `json:"author"`
	StoryURL    string `json:"story_url"`
	CreatedAt   string `json:"created_at"`
}

func NewHN(limit int) *HN {
	return &HN{
		client: &http.Client{},
		limit:  limit,
	}
}

func (h *HN) Name() string { return "hn" }

func (h *HN) Fetch(ctx context.Context, product string) ([]store.Comment, error) {
	parsedUrl, err := url.Parse(hnSearchURL)
	if err != nil {
		return nil, fmt.Errorf("url parse: %w", err)
	}

	values := parsedUrl.Query()
	values.Add("query", product)
	values.Add("tags", "comment")
	values.Add("hitsPerPage", fmt.Sprintf("%d", h.limit))

	parsedUrl.RawQuery = values.Encode()

	fmt.Println(parsedUrl.String())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsedUrl.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch hn: %w", err)
	}
	defer resp.Body.Close()

	var result hnResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode hn response: %w", err)
	}

	comments := make([]store.Comment, 0, len(result.Hits))
	for _, hit := range result.Hits {
		if hit.CommentText == "" {
			continue
		}
		comments = append(comments, store.Comment{
			ID:      "hn:" + hit.ObjectID,
			Product: product,
			Source:  h.Name(),
			Body:    stripHTML(hit.CommentText),
			Url:     "https://news.ycombinator.com/item?id=" + hit.ObjectID,
		})
	}

	return comments, nil
}

func stripHTML(s string) string {
	doc, err := html.Parse(strings.NewReader(s))
	if err != nil {
		return s
	}

	var b strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			b.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return strings.TrimSpace(b.String())
}
