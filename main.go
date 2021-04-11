package main

import (
	"log"
	"os"
)

const appname = "kingdom_views"

// TODO: delete and load from config file
var vid_id1 = []string{"BcA_p83ovAU", "FoUsDnDAnqM", "_rQjh07ccP4",
	"Md4ghQRdLwk", "EZIjLSZlUFs", "8trk5WKgjoI"}
var vid_id2 = []string{"4NqQ7dky_Uw", "M1TJGwIMG4Y", "Sa7u3kWeq3s"}

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
