package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

// https://api.telegram.org/bot<token>/METHOD_NAME

type Update struct {
	UpdateId int     `json:"update_id"`
	Message  Message `json:"message"`
}

type Message struct {
	Chat Chat   `json:"chat"`
	Text string `json:"text"`
}

type Chat struct {
	ChatId int `json:"id"`
}

type RestResponse struct {
	Result []Update `json:"result"`
}

type BotMessage struct {
	ChatId    int    `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode"`
}

type Bot struct {
	BotURL string
	Offset int
	YT     *YTStat
}

func NewBot(botURL string, offset int, yt *YTStat) *Bot {
	return &Bot{BotURL: botURL,
		Offset: offset,
		YT:     yt}
}

// botURL = TELEGRAM_URL+TELEGRAM_BOT_TOKEN
func (b *Bot) GetUpdates() ([]Update, error) {
	resp, err := http.Get(b.BotURL + "/getUpdates" + "?offset=" + strconv.Itoa(b.Offset))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var restResponse RestResponse
	err = json.Unmarshal(body, &restResponse)
	if err != nil {
		return nil, err
	}
	return restResponse.Result, err
}

func (b *Bot) SendResponse(update Update) error {
	var msg BotMessage
	var text string = fmt.Sprintf("Current count of views and likes on videos\n%18s|%15s|%15s|\t%s", "Views", "Likes", "Dislikes", "Name")
	var vid_ids *[]string
	switch update.Message.Text {
	case "/introduction_stage":
		vid_ids = &vid_id1
	case "/1_round":
		vid_ids = &vid_id2
	default:
		vid_ids = nil
	}
	resp, err := b.YT.MakeReqYTViews(vid_ids)
	if err != nil {
		log.Println(err)
		text = "Can not recognize command"
	}
	if resp != nil {
		for i, v := range resp.Items {
			text += fmt.Sprintf("\n%2d:%15v|%12v|%15v|\t<a href=\"http://y2u.be/%s\">%s</a>\n",
				i+1, v.Statistics.Views, v.Statistics.Likes, v.Statistics.Dislikes, v.Id, v.Snippet.Title)
		}
	}
	msg.ChatId = update.Message.Chat.ChatId
	msg.Text = text
	msg.ParseMode = "HTML"
	buf, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	_, err = http.Post(b.BotURL+"/sendMessage", "application/json", bytes.NewBuffer(buf))
	if err != nil {
		return err
	}

	return nil
}
