package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/alaust-dev/yt-datamining/internal"
	"github.com/alaust-dev/yt-datamining/internal/service"
	"github.com/go-co-op/gocron"
	"github.com/joho/godotenv"
)

var creatorChannelMap = []string{
	"UCYJ61XIK64sp6ZFFS8sctxw",
	"UCOzqhzAIw4JTJIMkDTfLnkg",
	"UCrgoxldvOW1Lj6MRvbktdig",
	"UCKbbjKo0BSaNB99FcY9bSPQ",
	"UCGdw90ycxxvubIvW4W8Zzzw",
	"UC7ASgzHJm6d0-yh_jGyEhbw",
	"UCz7eGR_UPsgq6rlDNXSWFug",
	"UC78qa96bVJpd6xrW-2FTRmw",
	"UCZHpIFMfoJJ_1QxNGLJTzyA",
	"UCfa7jJFYnn3P5LdJXsFkrjw",
	"UCL5-tPmf_sttES7ZcYJRp5A",
	"UCLoWcRy-ZjA-Erh0p_VDLjQ",
	"UCDmbhGe7-wC1a55l5ZYAZJw",
	"UC7E_mZfYy4IEsnNZFczopOg",
	"UCIckY7J5AHFsnUSulUq6E7g",
	"UC8E8eD7mOcnMazJT4laKbFQ",
	"UCxdaFKaMlRHtVkBeGygsYTw",
	"UCTGJRTPwtW_f_3ADDD9Sk1w",
	"UCYXZkXt3qNRyK5hebGgSqjA",
	"UCc6ZNbnzaab4G_FBMMjsVKg",
	"UCxivtyDib-xnd_fH3a3nQFg",
	"UCmUlXO8XjjtRtfrAygGmGTw",
	"UC7M_QoFOXa9MEVbfKSIeDSA",
}
var published_after = time.Date(2024, 2, 1, 0, 0, 0, 0, time.Now().Local().Location())

var database internal.Database
var yt service.Youtube
var dislikeApi service.DislikeApi

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Could not load .env file. Ignoring...")
	}

	api_key := os.Getenv("API_KEY")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	int_port, err := strconv.Atoi(port)
	if err != nil {
		panic("Could not parse Port to int: " + err.Error())
	}

	database = *internal.NewDatabase(host, int_port, user, password, dbname)
	yt = *service.NewYoutube(api_key)
	dislikeApi = *service.NewDislikeApi()

	// 2.2b cron scheduled for every 12 hours to fetch new data
	s := gocron.NewScheduler(time.Local)
	s.Every(12).Hour().Do(runJob)
	s.StartBlocking()
}

func runJob() {
	for _, channelId := range creatorChannelMap {
		// 2.1 Fetch chanel interaction data (channel views, video_count, etc.)
		channel := yt.FetchChannel(channelId)

		// 2.5 Saving channel interaction data to database
		database.InsertChannelInteraction(channel)

		fmt.Println("Processing Channel: " + channel.Snippet.Title)

		// 2.1 Fetch all viedeos in the timespan of the project (published_after - now)
		videos := yt.FetchVideosUntilDate(channel.ContentDetails.RelatedPlaylists.Uploads, published_after)
		for _, video := range videos {
			fmt.Println("Processing Video: " + video.Snippet.Title)

			// 2.1 Fetch dislikes from seperate api (return your dislike api)
			dislikes := dislikeApi.GetDislikes(video.Id)

			// 2.5 saving new video if it does not exist
			if !database.VideoWithIdExists(video.Id) {
				database.CreateVideo(channelId, &video)
			}

			// 2.5 Save fetched video data to db
			database.InsertViedoInteraction(&video, dislikes)
			database.InsertVideoCategory(&video)
			database.InsertVideoTags(&video)
		}
	}

	fmt.Println("Finished Job!")
}
