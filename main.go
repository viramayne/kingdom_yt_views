package main

import (
	"encoding/json"
	"log"
	"os"
)

const appname = "kingdom_views"

// TODO: delete and load from config file
var vid_id1 = []string{"BcA_p83ovAU", "FoUsDnDAnqM", "_rQjh07ccP4",
	"Md4ghQRdLwk", "EZIjLSZlUFs", "8trk5WKgjoI"}
var vid_id2 = []string{"4NqQ7dky_Uw", "M1TJGwIMG4Y", "Sa7u3kWeq3s"}

type Config struct {
	YTApiKey     string `json:"YOUTUBE_API_KEY"`
	YTChannelURL string `json:"YOUTUBE_CHANNEL_URL"`
	TGBotToken   string `json:"TELEGRAM_BOT_TOKEN"`
	TGBotURL     string `json:"TELEGRAM_URL"`
}

func main() {
	defer log.Printf("%s exiting", appname)

	file, _ := os.Open("config.json")
	decoder := json.NewDecoder(file)
	configuration := Config{}
	err := decoder.Decode(&configuration)
	if err != nil {
		log.Panic(err)
	}
	yt := NewYTStat(configuration.YTApiKey)

	botURL := configuration.TGBotURL + configuration.TGBotToken
	offset := 0
	bot := NewBot(botURL, offset, yt)
	for {
		updates, err := bot.GetUpdates()
		if err != nil {
			log.Println(err.Error())
		}
		for _, update := range updates {
			err = bot.SendResponse(update)
			if err != nil {
				log.Println(err.Error())
			}
			bot.Offset = update.UpdateId + 1
			log.Println(update)
		}
	}
}
