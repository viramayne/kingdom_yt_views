package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

type Video struct {
	Id          int
	VideoID     string
	Title       string
	PublishedAt time.Time
}

type Stats struct {
	ID       int
	VideoID  string
	Views    int
	Likes    int
	Dislikes int
	AtTime   time.Time
}

const (
	host     string = "localhost"
	port     int    = 5432
	user     string = "postgres"
	password string = "toortoor"
	dbName   string = "kingdom"

	sqlTimeFormat = "YYYY-MM-DD HH24:MI:SS" // Формат времени запроса в БД для sql
)

func ConnectToDB() (*sql.DB, error) {
	psqlInfo := fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s password=%s sslmode=disable",
		host, port, dbName, user, password)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	// defer db.Close()

	if err = db.Ping(); err != nil {
		log.Println(err)
		return nil, err
	}

	return db, nil
}

func (yt *YTStat) WriteToDBVideos(item *ItemSearch) error {
	tx, err := yt.Db.Begin()
	if err != nil {
		log.Printf("Can not begin transaction: %s", err)
		return err
	}

	queryStr := "INSERT INTO videos(videoId, title, publishedTime) " +
		"VALUES ($1,$2,$3);"
	if _, err = tx.Exec(queryStr, item.Id.VideoID, item.Snippet.Title,
		item.Snippet.PublishedTime); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		log.Printf("Can not begin transaction: %s", err)
		return err
	}
	tx.Commit()
	return nil
}

// Получить все видео id
func (yt *YTStat) GetVideoIds() (ids string, err error) {
	request := "SELECT * FROM videos ORDER BY id;"

	rows, err := yt.Db.Query(request)
	if err != nil {
		return
	}

	for rows.Next() {
		var item Video
		err = rows.Scan(&item.Id, &item.VideoID, &item.Title,
			&item.PublishedAt)
		if err != nil {
			return
		}

		ids += item.VideoID + ","
	}
	return ids[:len(ids)-1], nil
}

// Выбрать видео по дате публикации
func (yt *YTStat) ReadFromDBVideos(date string) (items []Video, err error) {
	request := "SELECT * FROM videos " +
		"WHERE date_trunc('day', publishedtime) = to_timestamp($1, $2) " +
		"ORDER BY id;"

	rows, err := yt.Db.Query(request, date, sqlTimeFormat)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var item Video
		err = rows.Scan(&item.Id, &item.VideoID, &item.Title,
			&item.PublishedAt)
		if err != nil {
			return
		}
		items = append(items, item)
	}
	return
}

func (yt *YTStat) WriteToDBStatistics(resp *Resp) error {
	tx, err := yt.Db.Begin()
	if err != nil {
		return err
	}
	queryStr := "INSERT INTO statistics(videoid, views, " +
		"likes, dislikes, attime) " +
		"VALUES ($1,$2,$3,$4,$5);"
	for _, item := range resp.Items {
		_, err := tx.Exec(queryStr, item.Id, item.Statistics.Views,
			item.Statistics.Likes, item.Statistics.Dislikes, time.Now())
		if err != nil {
			return err
		}
	}
	tx.Commit()
	return nil
}

func (yt *YTStat) GetVideosForDateFromBD(date string) ([]string, error) {
	items, err := yt.ReadFromDBVideos(date)
	if err != nil || len(items) == 0 {
		return []string{}, err
	}
	ids := make([]string, len(items))
	for i, vid := range items {
		ids[i] = vid.VideoID
	}
	return ids, nil
}

func (yt *YTStat) GetVideoTitleFromDB(id string) (string, error) {
	var title string
	request := "SELECT title from videos WHERE videoid = $1"
	rows, err := yt.Db.Query(request, id)
	if err != nil {
		return title, err
	}
	if rows.Next() {
		err = rows.Scan(&title)
		if err != nil {
			return title, err
		}
	}
	return title, nil
}

// Чтение данных статистики по указанным видео id
func (yt *YTStat) ReadFromDBStatistics(videoIds []string) ([]Stats, error) {
	result := make([]Stats, 0)
	for _, videoId := range videoIds {
		request := "SELECT * FROM statistics " +
			"WHERE videoid =$1 AND date_trunc('second',attime) = date_trunc('second',(SELECT MAX(attime) FROM statistics));"
		rows, err := yt.Db.Query(request, videoId)
		if err != nil {
			return nil, err
		}
		var items []Stats
		for rows.Next() {
			var item Stats
			err := rows.Scan(&item.ID, &item.VideoID, &item.Views, &item.Likes, &item.Dislikes, &item.AtTime)
			if err != nil {
				return nil, err
			}
			items = append(items, item)
		}
		result = append(result, items[len(items)-1])
	}
	return result, nil
}
