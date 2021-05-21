package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
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

	// Регистрируем обработку сигналов прерывания приложения, чтобы иметь возможность плавно его завершить
	interruptSignalChan := make(chan os.Signal, 1)
	signal.Notify(interruptSignalChan, os.Interrupt)
	termSignalChan := make(chan os.Signal, 1)
	signal.Notify(termSignalChan, syscall.SIGTERM)

	StartBot()

mainLoop:
	for {
		select {
		case <-interruptSignalChan:
			log.Printf("interrupt signal\n")
			break mainLoop
		case <-termSignalChan:
			log.Printf("term signal\n")
			break mainLoop

		}
	}

}
