package main

import (
	"log"
	"os"
)

const appname = "kingdom_views"

type Config struct {
	YTApiKey     string
	YTChannelURL string
	TGBotToken   string
	TGBotURL     string
}

func main() {
	defer log.Printf("%s exiting", appname)
	configuration := Config{
		YTApiKey:     os.Getenv("YOUTUBE_API_KEY"),
		YTChannelURL: os.Getenv("YOUTUBE_CHANNEL_URL"),
		TGBotToken:   os.Getenv("TELEGRAM_BOT_TOKEN"),
		TGBotURL:     os.Getenv("TELEGRAM_URL"),
	}

	yt, err := NewYTStat(configuration.YTApiKey, configuration.YTChannelURL)
	if err != nil {
		log.Fatalln(err)
	}

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
