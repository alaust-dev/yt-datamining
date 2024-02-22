package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type DislikeApi struct {
	client *http.Client
}

type VideoStats struct {
	Dislikes int32 `json:"dislikes"`
}

func NewDislikeApi() *DislikeApi {
	return &DislikeApi{client: &http.Client{}}
}

func (api *DislikeApi) GetDislikes(videoId string) int32 {
	url := "https://returnyoutubedislikeapi.com/Votes?videoId=" + videoId

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return 0
	}

	resp, err := api.client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return 0
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return 0
	}

	var stats VideoStats
	if err := json.Unmarshal(body, &stats); err != nil {
		fmt.Println("Error parsing JSON:", err)
		return 0
	}

	return stats.Dislikes
}
