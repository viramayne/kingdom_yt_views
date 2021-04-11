package main

import (
	"encoding/json"
	"errors"
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

func NewYTStat(api string, channelURL string) (*YTStat, error) {
	// Запросим Channel ID из URL канала
	channelID, err := makeReqForChannelID(channelURL, api)
	if err != nil {
		return nil, err
	}
	videos, err := getListOfVideos(api, channelID)
	if err != nil {
		return nil, err
	}
	return &YTStat{Api: api,
		ChannelID: channelID,
		Videos:    videos}, nil
}

func makeReqForChannelID(channelURL string, api string) (string, error) {
	idInd := strings.LastIndex(channelURL, "/") + 1

	request, err := http.NewRequest("GET", channelsURL, nil)
	if err != nil {

		return "", err
	}

	query := request.URL.Query()
	query.Add("key", api)
	query.Add("forUsername", channelURL[idInd:])
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

// MakeReqYTViews Отправляет запрос на получение статистики о видео
func (yt *YTStat) MakeReqYTViews(vid_ids *[]string) (*Resp, error) {
	if vid_ids == nil {
		return nil, errors.New("no video to make request")
	}
	resp, err := yt.getStatistics(vid_ids)
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
func (yt *YTStat) getVideoIds(vid_ids *[]string) string {
	var ids string
	for _, id := range *vid_ids {
		ids += id + ","
	}
	return ids[:len(ids)-1]
}

// getStatistics получение статистики о видео
func (yt *YTStat) getStatistics(vid_ids *[]string) (*Resp, error) {
	if vid_ids == nil {
		return nil, errors.New("can not get statistics for no list of videos")
	}
	var resp Resp

	request, err := http.NewRequest("GET", videosURL, nil)
	if err != nil {
		log.Println(err)
		return &resp, err
	}

	vidIds := yt.getVideoIds(vid_ids)
	query := request.URL.Query()
	query.Add("key", yt.Api)
	query.Add("id", vidIds)
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

func getListOfVideos(api, channelID string) (map[string]*[]string, error) {
	var resp RespSearch

	request, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	// GET https://youtube.googleapis.com/youtube/v3/search?part=snippet&channelId=UCbD8EppRX3ZwJSou-TVo90A&
	// order=date&publishedAfter=2021-03-31T00%3A00%3A00Z&q=%5B%ED%92%80%EB%B2%84%EC%A0%84%20&key=[YOUR_API_KEY] HTTP/1.1

	query := request.URL.Query()
	query.Add("channelId", channelID)
	query.Add("part", "snippet")
	query.Add("order", "date")
	query.Add("publishedAfter", "2021-03-31T00:00:00Z")
	query.Add("maxResults", "25")
	query.Add("q", "풀버전 킹덤 레전더리워")
	query.Add("key", api)

	request.URL.RawQuery = query.Encode()

	client := &http.Client{}
	r, err := client.Do(request)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer r.Body.Close()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	err = json.Unmarshal(body, &resp)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, errors.New("no results")
	}
	videos := make(map[string]*[]string)
	intro := make([]string, 0, 6)
	round1 := make([]string, 0, 6)

	for _, item := range resp.Items {
		if strings.HasPrefix(item.Snippet.Title, "[풀버전]") {
			if _, _, day := item.Snippet.PublishedTime.Date(); day == 1 {
				intro = append(intro, item.Id.VideoID)
			} else if day == 8 || day == 15 {
				round1 = append(round1, item.Id.VideoID)
			}
		}
	}
	videos["intro"] = &intro
	videos["1_round"] = &round1
	return videos, nil
}
