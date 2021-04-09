package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strconv"
)

// Resp struct to get needed info from yt api
type Resp struct {
	Items []Item `json:"items"`
}

type Item struct {
	Id         string     `json:"id"`
	Name       Snippet    `json:"snippet"`
	Statistics Statistics `json:"statistics"`
}

type Snippet struct {
	Title string `json:"title"`
}

type Statistics struct {
	Views string `json:"viewCount"`
	Likes string `json:"likeCount"`
	// Dislikes string `json:"dislikeCount"`
}

type YTStat struct {
	Api string
}

func NewYTStat(api string) *YTStat {
	return &YTStat{Api: api}
}

// MakeReqYTViews Отправляет запрос на получение статистики о видео
func (yt *YTStat) MakeReqYTViews(vid_ids *[]string) (*Resp, error) {
	if vid_ids == nil {
		return nil, errors.New("no video to make request")
	}
	resp, err := yt.getStatistics(vid_ids)
	if err != nil {
		return resp, err
	}
	// Сортируем по убыванию числа просмотров видео
	sortRespByViews(resp)

	// // Вывод данных в консоль
	// log.Println(" \t Views \t\t Likes \t\t Name")
	// for i, v := range resp.Items {
	// 	log.Printf("%d: %9v  %12v  %s", i+1, v.Statistics.Views, v.Statistics.Likes, v.Name.Title)
	// }
	return resp, nil
}

// Сортировка полученной информации по количеству просмотров видео в убывающем порядке
func sortRespByViews(resp *Resp) {
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

	request, err := http.NewRequest("GET", "https://www.googleapis.com/youtube/v3/videos", nil)
	if err != nil {
		log.Println(err)
		return &resp, err
	}

	vidIds := yt.getVideoIds(vid_ids)
	query := request.URL.Query()
	query.Add("key", yt.Api)                // https://developers.google.com/youtube/v3/getting-started
	query.Add("id", vidIds)                 // video IDs to get sub count of
	query.Add("part", "snippet,statistics") // leave this how it is
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
		return &resp, nil
	}

	return &resp, nil
}

// GET https://youtube.googleapis.com/youtube/v3/channels?part=snippet&forUsername=Mnet&key=[YOUR_API_KEY] HTTP/1.1

// Authorization: Bearer [YOUR_ACCESS_TOKEN]
// Accept: application/json
// youtube channel url "https://www.youtube.com/c/Mnet"

// GET https://youtube.googleapis.com/youtube/v3/search?part=snippet&channelId=UCbD8EppRX3ZwJSou-TVo90A&order=title&publishedAfter=2021-03-31T00%3A00%3A00Z&q=%5B%ED%92%80%EB%B2%84%EC%A0%84%5D%20-%EC%A7%81%EC%BA%A0%20-%ED%9A%8C&safeSearch=none&key=[YOUR_API_KEY] HTTP/1.1
