package main

import (
	"os"
)

const (
	YT_Channel = "https://www.youtube.com/c/Mnet"
)

var (
	YT_Key   = os.Getenv("YOUTUBE_API_KEY")
	BotToken = os.Getenv("TELEGRAM_BOT_TOKEN")
	Port     = os.Getenv("PORT")
)

func main() {

	StartBot()
}
