package webhook

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/simesaba80/toybox-back/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestDiscordWorkNotifier_NotifyCreated(t *testing.T) {
	workID := uuid.New()
	description := strings.Repeat("あ", maxDescriptionRune+1)
	work := &entity.Work{
		ID:           workID,
		Title:        "新しい作品",
		Description:  description,
		ThumbnailURL: "https://example.com/thumbnail.png",
		User: &entity.User{
			Name:      "author-name",
			AvatarURL: "https://example.com/avatar.png",
		},
	}

	var received discordWebhookPayload
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		require.NoError(t, json.NewDecoder(r.Body).Decode(&received))
		return &http.Response{
			StatusCode: http.StatusNoContent,
			Body:       io.NopCloser(strings.NewReader("")),
			Header:     make(http.Header),
		}, nil
	})}

	notifier := newDiscordWorkNotifier(client, "https://discord.example.com/webhook", "https://toybox.example.com/")
	err := notifier.NotifyCreated(context.Background(), work)

	require.NoError(t, err)
	require.Len(t, received.Embeds, 1)
	assert.Equal(t, work.User.Name, received.Username)
	assert.Equal(t, work.User.AvatarURL, received.AvatarURL)
	assert.Equal(t, work.Title, received.Embeds[0].Title)
	assert.Equal(t, "https://toybox.example.com/works/"+workID.String(), received.Embeds[0].URL)
	assert.Equal(t, strings.Repeat("あ", maxDescriptionRune)+"…", received.Embeds[0].Description)
	assert.Equal(t, toyboxLogoURL, received.Embeds[0].Thumbnail.URL)
	assert.Equal(t, discordEmbedColor, received.Embeds[0].Color)
	assert.Equal(t, work.ThumbnailURL, received.Embeds[0].Image.URL)
}

func TestDiscordWorkNotifier_NotifyCreatedReturnsErrorForFailedRequest(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusBadRequest,
			Body:       io.NopCloser(strings.NewReader("")),
			Header:     make(http.Header),
		}, nil
	})}

	notifier := newDiscordWorkNotifier(client, "https://discord.example.com/webhook", "https://toybox.example.com")
	err := notifier.NotifyCreated(context.Background(), &entity.Work{User: &entity.User{}})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "status 400")
}

func TestDiscordWorkNotifier_NotifyCreatedReturnsErrorWithoutConfiguration(t *testing.T) {
	notifier := newDiscordWorkNotifier(http.DefaultClient, "", "https://toybox.example.com")

	err := notifier.NotifyCreated(context.Background(), &entity.Work{User: &entity.User{}})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "DISCORD_WEBHOOK_URL")
}
