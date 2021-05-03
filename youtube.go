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
	query.Add("part", "id")
	query.Add("forUsername", YT_Channel[idInd:])
	query.Add("key", YT_Key)
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
	idsStr := yt.formVideoIdsString(yt.Videos[date])

	resp, err := yt.getStatistics(idsStr)
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

// getStatistics получение статистики о спике видео за определенную дату
// в формате "yyyy-mm-dd"
func (yt *YTStat) getStatistics(ids string) (*Resp, error) {
	var resp Resp
	// GET https://youtube.googleapis.com/youtube/v3/videos?part=snippet%2CcontentDetails%2Cstatistics
	// &id=Ks-_Mh1QhMc&key=[YOUR_API_KEY] HTTP/1.1

	request, err := http.NewRequest("GET", videosURL, nil)
	if err != nil {
		log.Println(err)
		return &resp, err
	}

	query := request.URL.Query()
	query.Add("key", yt.Api)
	query.Add("id", ids) // id видео через запятую, если несколько
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

	dates := make([]time.Time, 0)
	dates = append(dates,
		time.Date(2021, 04, 1, 0, 0, 0, 0, time.Local),
		time.Date(2021, 04, 8, 0, 0, 0, 0, time.Local),
		time.Date(2021, 04, 15, 0, 0, 0, 0, time.Local),
		time.Date(2021, 04, 22, 0, 0, 0, 0, time.Local),
		time.Date(2021, 04, 29, 0, 0, 0, 0, time.Local),
	)

	videos := make(map[string]*[]string)
	yt.Videos = videos

	for _, date := range dates {
		err := yt.getDataByPublishedDay(&date)
		if err != nil {
			return nil, err
		}
	}

	return &videos, nil
}

// получаем список видео за указанную дату публикации
func (yt *YTStat) getDataByPublishedDay(publishedAfter *time.Time) error {

	today := time.Now()
	if publishedAfter.After(today) {
		return errors.New("no video on channel yet")
	}

	var resp RespSearch

	request, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		log.Println(err)
		return err
	}

	query := request.URL.Query()
	query.Add("channelId", yt.ChannelID)
	query.Add("part", "snippet")
	query.Add("order", "date")
	query.Add("publishedAfter", publishedAfter.Format("2006-01-02T15:04:05Z"))
	// log.Println("publishedAfter: " + publishedAfter.Format("2006-01-02T15:04:05Z"))
	if publishedAfter.Day() != today.Day() {
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
	return yt.formMsgForDate("2021-04-22") + "\n" +
		yt.formMsgForDate("2021-04-29")
}

func (yt *YTStat) FillMsgForVideo(url string) string {
	return yt.formMsgForVideo(url)
}

func (yt *YTStat) formMsgForVideo(url string) string {
	// get video id from url
	idInd := strings.LastIndex(url, "/") + 1
	if strings.HasPrefix(url[idInd:], "watch?v=") {
		idInd += 8
	}

	// make statistics request
	var msgTxt string = "Current statistics for video: \n<b>"

	resp, err := yt.getStatistics(url[idInd:])
	if err != nil {
		return ""
	}
	// form respond msg text
	b := BeautifyNumbers

	if resp != nil {
		for _, v := range resp.Items {
			msgTxt += v.Snippet.Title
			msgTxt += fmt.Sprintf("</b>\n\n%18s|%15s|%15s\n",
				"Views", "Likes", "Dislikes")
			msgTxt += fmt.Sprintf("%15v|%12v|%15v",
				b(v.Statistics.Views), b(v.Statistics.Likes), b(v.Statistics.Dislikes))
		}
	}
	return msgTxt
}

func (yt *YTStat) formMsgForDate(date string) string {
	var text string = fmt.Sprintf("\naired %s:\n", date)
	if yt.Videos[date] == nil {
		return text + "\nNo videos for perfomances on channel.\n<i><b>Perhaps round not aired yet!</b></i>"
	}
	idsStr := yt.formVideoIdsString(yt.Videos[date])

	resp, err := yt.getStatistics(idsStr)
	if err != nil {
		return err.Error()
	}
	resp.sortRespByViews()

	b := BeautifyNumbers

	if resp != nil {
		for i, v := range resp.Items {
			text += fmt.Sprintf("\n%2d:%15v|%12v|%15v|\t<a href=\"http://y2u.be/%s\">%s</a>\n",
				i+1, b(v.Statistics.Views), b(v.Statistics.Likes), b(v.Statistics.Dislikes),
				v.Id, v.Snippet.Title)
		}
	}
	return text
}
