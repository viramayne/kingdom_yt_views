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
	ChatId    int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
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
	var resp *Resp
	var err error
	var text string = fmt.Sprintf("Current count of views and likes on videos\n%18s|%15s|%15s|\t%s", "Views", "Likes", "Dislikes", "Name")
	var vid_ids *[]string = nil

	switch update.Message.Text {

	case "/start":
		text = fmt.Sprintf("Hello %s %s!", update.Message.Chat.FirstName, update.Message.Chat.LastName)

	case "/introduction_stage":
		if b.YT.Videos["2021-04-01"] != nil {
			vid_ids = b.YT.Videos["2021-04-01"]
		}

	case "/1_round":
		if b.YT.Videos["2021-04-08"] != nil && b.YT.Videos["2021-04-15"] != nil {
			vid_ids = ConcatinateArrays(*b.YT.Videos["2021-04-08"], *b.YT.Videos["2021-04-15"])
		}

	default:
		text = "Can not recognize command"
	}

	if vid_ids != nil {
		resp, err = b.YT.MakeReqYTViews(vid_ids)
		if err != nil {
			log.Println(err)
		}
		if resp != nil {
			for i, v := range resp.Items {
				text += fmt.Sprintf("\n%2d:%15v|%12v|%15v|\t<a href=\"http://y2u.be/%s\">%s</a>\n",
					i+1, v.Statistics.Views, v.Statistics.Likes, v.Statistics.Dislikes, v.Id, v.Snippet.Title)
			}
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
