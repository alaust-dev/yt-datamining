package internal

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/lib/pq"
	"google.golang.org/api/youtube/v3"
)

type Database struct {
	db *sql.DB
}

func NewDatabase(host string, port int, user string, password string, database string) *Database {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, database)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic("Could not connect to db: " + err.Error())
	}

	return &Database{db: db}
}

func (database *Database) InsertChannelInteraction(channel *youtube.Channel) {
	_, err := database.db.Exec("INSERT INTO channel_interaction (video_count, view_count, subscriber_count, channel_id) VALUES ($1, $2, $3, $4)",
		channel.Statistics.VideoCount, channel.Statistics.ViewCount, channel.Statistics.SubscriberCount, channel.Id)
	if err != nil {
		fmt.Printf("Could not insert channel interaction (channel_id: %s) to database: %s\n", channel.Id, err.Error())
	}
}

func (database *Database) CreateVideo(channelId string, video *youtube.Video) {
	_, err := database.db.Exec("INSERT INTO video (id, title, description, fsk_rating, yt_age_restricted, made_for_kids, self_declared_made_for_kids, duration, channel_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)",
		video.Id,
		video.Snippet.Title,
		video.Snippet.Description,
		video.ContentDetails.ContentRating.FskRating,
		video.ContentDetails.ContentRating.YtRating == "ytAgeRestricted",
		video.Status.MadeForKids,
		video.Status.SelfDeclaredMadeForKids,
		video.ContentDetails.Duration,
		channelId)

	if err != nil {
		fmt.Println("Could not create video in db: " + err.Error())
	}
}

func (database *Database) InsertViedoInteraction(video *youtube.Video, dislikes int32) {
	database.db.Exec("INSERT INTO video_interaction (comments, views, likes, video_id, dislikes) VALUES ($1, $2, $3, $4, $5)",
		video.Statistics.CommentCount,
		video.Statistics.ViewCount,
		video.Statistics.LikeCount,
		video.Id,
		dislikes)
}

func (database *Database) InsertVideoCategory(video *youtube.Video) {
	if video.TopicDetails == nil {
		fmt.Println("Empty topic details, skipping for video: " + video.Snippet.Title)
		return
	}

	for _, category := range video.TopicDetails.TopicCategories {
		row := database.db.QueryRow("SELECT id FROM topic_categories WHERE wiki_link = $1", category)

		var categoryId int64
		err := row.Scan(&categoryId)

		if errors.Is(err, sql.ErrNoRows) {
			r, err := database.db.Exec("INSERT INTO topic_categories (wiki_link) VALUES ($1)", category)
			if err != nil {
				fmt.Println("Could not create video_category in db: " + err.Error())
				continue
			}
			categoryId, _ = r.LastInsertId()
		}

		_, err = database.db.Exec("INSERT INTO video_topics (video_id, topic_id) VALUES ($1, $2)", video.Id, categoryId)
	}
}

func (database *Database) InsertVideoTags(video *youtube.Video) {
	for _, tag := range video.Snippet.Tags {
		row := database.db.QueryRow("SELECT id FROM tags WHERE name = $1", tag)

		var tagId int64
		err := row.Scan(&tagId)

		if errors.Is(err, sql.ErrNoRows) {
			r, err := database.db.Exec("INSERT INTO tags (name) VALUES ($1)", tag)
			if err != nil {
				fmt.Println("Could not create tag in db: " + err.Error())
				continue
			}
			tagId, _ = r.LastInsertId()
		}

		_, err = database.db.Exec("INSERT INTO video_tags (video_id, tag_id) VALUES ($1, $2)", video.Id, tagId)
	}
}

func (database *Database) VideoWithIdExists(videoId string) bool {
	row := database.db.QueryRow("SELECT id FROM video WHERE id = $1", videoId)

	var exists string
	err := row.Scan(&exists)

	if errors.Is(err, sql.ErrNoRows) {
		return false
	} else if err != nil {
		fmt.Println("Could not check if video is new in database: " + err.Error())
		return false
	}

	return true
}
