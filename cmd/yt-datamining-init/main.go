package main

import (
	"context"
	"database/sql"
	"fmt"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"

	_ "github.com/lib/pq"
)

var creatorChannelMap = map[string]int{
	"UCYJ61XIK64sp6ZFFS8sctxw": 1,
	"UCOzqhzAIw4JTJIMkDTfLnkg": 2,
	"UCrgoxldvOW1Lj6MRvbktdig": 2,
	"UCKbbjKo0BSaNB99FcY9bSPQ": 3,
	"UC7ASgzHJm6d0-yh_jGyEhbw": 3,
	"UCGdw90ycxxvubIvW4W8Zzzw": 3,
	"UCz7eGR_UPsgq6rlDNXSWFug": 3,
	"UCLoWcRy-ZjA-Erh0p_VDLjQ": 4,
	"UCZHpIFMfoJJ_1QxNGLJTzyA": 4,
	"UCfa7jJFYnn3P5LdJXsFkrjw": 4,
	"UC78qa96bVJpd6xrW-2FTRmw": 4,
	"UCL5-tPmf_sttES7ZcYJRp5A": 4,
	"UCDmbhGe7-wC1a55l5ZYAZJw": 5,
	"UC8E8eD7mOcnMazJT4laKbFQ": 5,
	"UCxdaFKaMlRHtVkBeGygsYTw": 5,
	"UCIckY7J5AHFsnUSulUq6E7g": 5,
	"UC7E_mZfYy4IEsnNZFczopOg": 5,
	"UCYXZkXt3qNRyK5hebGgSqjA": 6,
	"UCTGJRTPwtW_f_3ADDD9Sk1w": 6,
	"UCc6ZNbnzaab4G_FBMMjsVKg": 6,
	"UCmUlXO8XjjtRtfrAygGmGTw": 7,
	"UCxivtyDib-xnd_fH3a3nQFg": 7,
	"UC7M_QoFOXa9MEVbfKSIeDSA": 7,
}

const (
	api_key = ""

	host     = ""
	port     = 5432
	user     = ""
	password = ""
	dbname   = ""
)

// This programm just do the init part of saving channel meta data like channel name for once

func main() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic("Could not connect to db: " + err.Error())
	}

	ctx := context.Background()
	service, err := youtube.NewService(ctx, option.WithAPIKey(api_key))
	if err != nil {
		panic("Could not get youtube service: " + err.Error())
	}
	defer db.Close()

	for channelId, creatorId := range creatorChannelMap {
		results, err := service.Channels.List([]string{"snippet", "statistics"}).Id(channelId).Do()
		channel := results.Items[0]
		if err != nil {
			panic(fmt.Sprintf("Failed to fetch channel (%s): %s", channelId, err.Error()))
		}

		_, err = db.Exec("INSERT INTO channel (id, name, description, creator_id) VALUES ($1, $2, $3, $4)",
			channel.Id, channel.Snippet.Title, channel.Snippet.Description, creatorId)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		_, err = db.Exec("INSERT INTO channel_interaction (video_count, view_count, subscriber_count, channel_id) VALUES ($1, $2, $3, $4)",
			channel.Statistics.VideoCount, channel.Statistics.ViewCount, channel.Statistics.SubscriberCount, channel.Id)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
	}
}
