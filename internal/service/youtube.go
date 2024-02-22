package service

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type Youtube struct {
	service *youtube.Service
}

func NewYoutube(apiKey string) *Youtube {
	ctx := context.Background()
	service, err := youtube.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		panic("Could not create youtube service: " + err.Error())
	}

	return &Youtube{service: service}
}

func (yt *Youtube) Service() *youtube.Service {
	return yt.service
}

func (yt *Youtube) FetchChannel(channelId string) *youtube.Channel {
	channels, err := yt.Service().Channels.List([]string{"statistics", "contentDetails", "snippet"}).Id(channelId).Do()
	if err != nil {
		fmt.Println("Failed to call API: " + err.Error())
	}

	return channels.Items[0]
}

func (yt *Youtube) FetchVideosUntilDate(playlistId string, date time.Time) []youtube.Video {
	videos := []youtube.Video{}
	yt.fetchVideosUntilDate(playlistId, date, "", &videos)
	return videos
}

func (yt *Youtube) fetchVideosUntilDate(playlistId string, date time.Time, pageToken string, videos *[]youtube.Video) {
	call := yt.service.PlaylistItems.List([]string{"snippet"}).PlaylistId(playlistId)
	if pageToken != "" {
		call.PageToken(pageToken)
	}

	uploads, err := call.Do()
	if err != nil {
		fmt.Println("Failed to get uploads of channel: " + err.Error())
		return
	}

	for _, item := range uploads.Items {
		publish_date, _ := time.Parse(time.RFC3339, item.Snippet.PublishedAt)
		if publish_date.Before(date) {
			return
		}

		videoResponse, err := yt.service.Videos.List([]string{"statistics", "snippet", "contentDetails", "topicDetails", "status"}).
			Id(item.Snippet.ResourceId.VideoId).Do()
		if err != nil {
			fmt.Println("Could not fetch video: " + err.Error())
			continue
		}

		*videos = append(*videos, *videoResponse.Items[0])
	}

	yt.fetchVideosUntilDate(playlistId, date, uploads.NextPageToken, videos)
}
