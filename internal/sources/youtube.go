package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/GatosTheDog/versous/internal/store"
)

const videoSearchUrl = "https://www.googleapis.com/youtube/v3/search"
const commentSearchUrl = "https://www.googleapis.com/youtube/v3/commentThreads"

type Youtube struct {
	client      *http.Client
	apiKey      string
	maxComments int
	maxVideos   int
}

type YoutubeSearchResponse struct {
	Items []struct {
		ID struct {
			VideoId string `json:"videoId"`
		} `json:"id"`
	} `json:"items"`
}

type YoutubeCommentsResponse struct {
	Items []struct {
		Snippet struct {
			TopLevelComment struct {
				Snippet struct {
					TextDisplay string `json:"textDisplay"`
					AuthorName  string `json:"authorDisplayName"`
				} `json:"snippet"`
			} `json:"topLevelComment"`
		} `json:"snippet"`
	} `json:"items"`
}

func NewYoutube(maxComments, maxVideos int) *Youtube {
	return &Youtube{
		client:      &http.Client{},
		apiKey:      os.Getenv("YOUTUBE_API_KEY"),
		maxComments: maxComments,
		maxVideos:   maxVideos,
	}
}

func (yt *Youtube) Name() string { return "youtube" }

func (yt *Youtube) Fetch(ctx context.Context, product string) ([]store.Comment, error) {
	parsedVideoUrl, err := url.Parse(videoSearchUrl)
	if err != nil {
		return nil, fmt.Errorf("videoSearchUrl parse: %w", err)
	}

	values := parsedVideoUrl.Query()
	values.Add("part", "id")
	values.Add("type", "video")
	values.Add("maxResults", fmt.Sprintf("%d", yt.maxVideos))
	values.Add("q", product)
	values.Add("key", yt.apiKey)

	parsedVideoUrl.RawQuery = values.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsedVideoUrl.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	resp, err := yt.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch youtube: %w", err)
	}
	defer resp.Body.Close()

	var resultVideos YoutubeSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&resultVideos); err != nil {
		return nil, fmt.Errorf("decode youtube response: %w", err)
	}

	comments := make([]store.Comment, 0)
	for _, video := range resultVideos.Items {
		var resultComments YoutubeCommentsResponse
		parsedCommentUrl, err := url.Parse(commentSearchUrl)
		if err != nil {
			return nil, fmt.Errorf("commentSearchUrl parse: %w", err)
		}

		values := parsedCommentUrl.Query()
		values.Add("part", "snippet")
		values.Add("maxResults", fmt.Sprintf("%d", yt.maxComments))
		values.Add("videoId", video.ID.VideoId)
		values.Add("key", yt.apiKey)

		parsedCommentUrl.RawQuery = values.Encode()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsedCommentUrl.String(), nil)
		if err != nil {
			return nil, fmt.Errorf("build request: %w", err)
		}

		resp, err := yt.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("fetch youtube: %w", err)
		}

		if err := json.NewDecoder(resp.Body).Decode(&resultComments); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("decode youtube response: %w", err)
		}

		for i, item := range resultComments.Items {
			body := stripHTML(item.Snippet.TopLevelComment.Snippet.TextDisplay)
			if len(body) < 50 {
				continue
			}
			comments = append(comments, store.Comment{
				ID:     fmt.Sprintf("yt:%s:%d", video.ID.VideoId, i),
				Source: yt.Name(),
				Body:   body,
				Url:    "https://youtube.com/watch?v=" + video.ID.VideoId,
			})
		}
		resp.Body.Close()
	}
	return comments, nil
}
