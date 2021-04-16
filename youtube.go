package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

// RespSearch struct to get needed info from yt api
type RespSearch struct {
	Items []ItemSearch `json:"items"`
}

type ItemSearch struct {
	Id         ID         `json:"id"`
	Statistics Statistics `json:"statistics"`
	Snippet    Snippets   `json:"snippet"`
}
type ID struct {
	VideoID string `json:"videoId"`
}

// Resp struct to get needed info from yt api
type Resp struct {
	Items []Item `json:"items"`
}

type Item struct {
	Id         string     `json:"id"`
	Statistics Statistics `json:"statistics"`
	Snippet    Snippets   `json:"snippet"`
}

type Snippets struct {
	Title         string    `json:"title"`
	PublishedTime time.Time `json:"publishedAt"`
}

type Statistics struct {
	Views    string `json:"viewCount"`
	Likes    string `json:"likeCount"`
	Dislikes string `json:"dislikeCount"`
}

type YTStat struct {
	Api       string
	ChannelID string
	Videos    map[string]*[]string
}

var (
	channelsURL string = "https://youtube.googleapis.com/youtube/v3/channels"
	videosURL   string = "https://youtube.googleapis.com/youtube/v3/videos"
	searchURL   string = "https://youtube.googleapis.com/youtube/v3/search"
)

func NewYTStatistics() (*YTStat, error) {
	// Запросим Channel ID из URL канала
	channelID, err := getChannelID()
	if err != nil {
		return nil, err
	}
	return &YTStat{Api: YT_Key,
		ChannelID: channelID,
	}, nil
}

// Определяет id канала по channelName из url
func getChannelID() (string, error) {
	idInd := strings.LastIndex(YT_Channel, "/") + 1

	request, err := http.NewRequest("GET", channelsURL, nil)
	if err != nil {
		return "", err
	}

	query := request.URL.Query()
	query.Add("key", YT_Key)
	query.Add("forUsername", YT_Channel[idInd:])
	query.Add("part", "snippet")
	request.URL.RawQuery = query.Encode()

	client := &http.Client{}
	r, err := client.Do(request)
	if err != nil {

		return "", err
	}
	defer r.Body.Close()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		return "", err
	}
	var resp Resp

	err = json.Unmarshal(body, &resp)
	if err != nil {
		log.Println(err)
		return "", err
	}
	if len(resp.Items) == 0 {
		return "", errors.New("no result for channel ID")
	}

	if err != nil {
		return "", err
	}
	return resp.Items[0].Id, nil
}

// UpdateVideoIDList Обновляет список видео для получения статистики
func (yt *YTStat) UpdateVideoIDList() error {
	videos, err := yt.updateListOfVideos()
	if err != nil {
		return err
	}
	yt.Videos = *videos
	return nil
}

// GetVideoStatisticsForDate Отправляет запрос на получение статистики о видео
// date format "2021-04-01" - date of performance
func (yt *YTStat) GetVideoStatisticsForDate(date string) (*Resp, error) {
	if yt.Videos[date] == nil || len(*yt.Videos[date]) == 0 {
		return nil, errors.New("no video to make request")
	}
	resp, err := yt.getStatistics(date)
	if err != nil {
		return nil, err
	}
	// Сортируем по убыванию числа просмотров видео
	resp.sortRespByViews()

	return resp, nil
}

// Сортировка полученной информации по количеству просмотров видео в убывающем порядке
func (resp *Resp) sortRespByViews() {
	sort.Slice(resp.Items, func(i, j int) (less bool) {
		iv, _ := strconv.ParseInt(resp.Items[i].Statistics.Views, 10, 64)
		jv, _ := strconv.ParseInt(resp.Items[j].Statistics.Views, 10, 64)
		return iv > jv
	})
}

// соединение id видео для запроса инфо
func (yt *YTStat) formVideoIdsString(vid_ids *[]string) string {
	var ids string
	for _, id := range *vid_ids {
		ids += id + ","
	}
	if len(ids) == 0 {
		return ""
	}
	return ids[:len(ids)-1]
}

// getStatistics получение статистики о видео
func (yt *YTStat) getStatistics(date string) (*Resp, error) {
	var resp Resp
	// GET https://youtube.googleapis.com/youtube/v3/videos?part=snippet%2CcontentDetails%2Cstatistics
	// &id=Ks-_Mh1QhMc&key=[YOUR_API_KEY] HTTP/1.1

	request, err := http.NewRequest("GET", videosURL, nil)
	if err != nil {
		log.Println(err)
		return &resp, err
	}

	vidIdsStr := yt.formVideoIdsString(yt.Videos[date])

	query := request.URL.Query()
	query.Add("key", yt.Api)
	query.Add("id", vidIdsStr) // id видео через запятую, если несколько
	query.Add("part", "snippet,statistics")
	request.URL.RawQuery = query.Encode()

	client := &http.Client{}
	r, err := client.Do(request)
	if err != nil {
		log.Println(err)
		return &resp, err
	}
	defer r.Body.Close()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		return &resp, err
	}

	err = json.Unmarshal(body, &resp)
	if err != nil {
		log.Println(err)
		return &resp, err
	}
	if len(resp.Items) == 0 {
		return &resp, errors.New("no result for videos statistics")
	}

	return &resp, nil
}

func (yt *YTStat) updateListOfVideos() (*map[string]*[]string, error) {
	dateIntro := time.Date(2021, 04, 01, 0, 0, 0, 0, time.Local)
	date1Round1Day := time.Date(2021, 04, 8, 0, 0, 0, 0, time.Local)
	date1Round2Day := time.Date(2021, 04, 15, 0, 0, 0, 0, time.Local)

	videos := make(map[string]*[]string)
	yt.Videos = videos
	err := yt.getDataByPublishedDay(&dateIntro)
	if err != nil {
		return nil, err
	}
	err = yt.getDataByPublishedDay(&date1Round1Day)
	if err != nil {
		return nil, err
	}
	err = yt.getDataByPublishedDay(&date1Round2Day)
	if err != nil {
		return nil, err
	}

	return &videos, nil
}

// получаем список видео за указанную дату публикации
func (yt *YTStat) getDataByPublishedDay(publishedAfter *time.Time) error {
	var resp RespSearch

	request, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		log.Println(err)
		return err
	}
	today := time.Now()
	query := request.URL.Query()
	query.Add("channelId", yt.ChannelID)
	query.Add("part", "snippet")
	query.Add("order", "date")
	query.Add("publishedAfter", publishedAfter.Format("2006-01-02T15:04:05Z"))
	// log.Println("publishedAfter: " + publishedAfter.Format("2006-01-02T15:04:05Z"))
	if publishedAfter.Day() < today.Day() {
		publishedBefore := publishedAfter.Add(24 * time.Hour)
		query.Add("publishedBefore", publishedBefore.Format("2006-01-02T15:04:05Z"))
		// log.Println("publishedBefore: " + publishedBefore.Format("2006-01-02T15:04:05Z"))
	}
	query.Add("maxResults", "25")
	query.Add("q", "[풀버전]")
	query.Add("key", yt.Api)

	request.URL.RawQuery = query.Encode()

	client := &http.Client{}
	r, err := client.Do(request)
	if err != nil {
		log.Println(err)
		return err
	}
	defer r.Body.Close()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		return err
	}

	err = json.Unmarshal(body, &resp)
	if err != nil {
		log.Println(err)
		return err
	}
	if len(resp.Items) == 0 {
		return errors.New("no results")
	}

	videosForDay := make([]string, 0, 6)
	date := publishedAfter.Format("2006-01-02")

	for _, item := range resp.Items {
		if strings.HasPrefix(item.Snippet.Title, "[풀버전]") {
			videosForDay = append(videosForDay, item.Id.VideoID)
		}
	}

	yt.Videos[date] = &videosForDay
	return nil
}

func (yt *YTStat) FillMsgForIntro() string {
	return yt.formMsgForDate("2021-04-01")
}

func (yt *YTStat) FillMsgForFirstRound() string {
	return yt.formMsgForDate("2021-04-08") + "\n" +
		yt.formMsgForDate("2021-04-15")
}

func (yt *YTStat) FillMsgForSecondRound() string {
	return yt.formMsgForDate("2021-04-22")
}

func (yt *YTStat) formMsgForDate(date string) string {
	var text string
	if yt.Videos[date] == nil {
		return "\nNo videos for perfomances on channel.\n<i><b>Perhaps round not aired yet!</b></i>"
	}
	resp, err := yt.getStatistics(date)
	if err != nil {
		return err.Error()
	}
	resp.sortRespByViews()

	if resp != nil {
		for i, v := range resp.Items {
			text += fmt.Sprintf("\n%2d:%15v|%12v|%15v|\t<a href=\"http://y2u.be/%s\">%s</a>\n",
				i+1, v.Statistics.Views, v.Statistics.Likes, v.Statistics.Dislikes,
				v.Id, v.Snippet.Title)
		}
	}
	return text
}
