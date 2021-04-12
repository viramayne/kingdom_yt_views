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
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("$PORT must be set")
	}

	log.Println("Getting parameters from env")
	configuration := Config{
		YTApiKey:     os.Getenv("YOUTUBE_API_KEY"),
		YTChannelURL: os.Getenv("YOUTUBE_CHANNEL_URL"),
		TGBotToken:   os.Getenv("TELEGRAM_BOT_TOKEN"),
		TGBotURL:     os.Getenv("TELEGRAM_URL"),
	}

	log.Printf("Creating YouTube client for %s\n", configuration.YTChannelURL)
	yt, err := NewYTStat(configuration.YTApiKey, configuration.YTChannelURL)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Creating Telegram bot client")
	offset := 0
	bot := NewBot(configuration.TGBotURL+configuration.TGBotToken, offset, yt)
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
