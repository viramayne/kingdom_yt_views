package main

import (
	"fmt"
	"log"
	"net/http"
	"reflect"

	tgbotapi "github.com/Syfaro/telegram-bot-api"
)

func StartBot() {
	//Создаем бота
	bot, err := tgbotapi.NewBotAPI(BotToken)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Authorized on account %s\n", bot.Self.UserName)

	// Инициализируем связь с YouTube
	yt, err := NewYTStatistics()
	if err != nil {
		panic(err)
	}

	// Обновляем список видео выступлений
	err = yt.UpdateVideoIDList()
	if err != nil {
		panic(err)
	}

	//Устанавливаем время обновления
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	//Получаем обновления от бота
	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		panic(err)
	}

	go http.ListenAndServe(":"+Port, nil)
	log.Printf("start listen :%s", Port)
	for update := range updates {
		if update.Message == nil {
			continue
		}

		//Проверяем что от пользователья пришло именно текстовое сообщение
		if reflect.TypeOf(update.Message.Text).Kind() == reflect.String && update.Message.Text != "" {
			var msgText string
			var headTxt string = fmt.Sprintf("Current count of views and likes on videos\n%18s|%15s|%15s|\t%s\n",
				"Views", "Likes", "Dislikes", "Name")
			switch update.Message.Text {
			case "/start":
				msgText = fillMsgForStart(update.Message.Chat.FirstName, update.Message.Chat.LastName)

			case "/introduction_stage":
				msgText = headTxt + yt.FillMsgForIntro()

			case "/1_round":
				msgText = headTxt + yt.FillMsgForFirstRound()

			case "/2_round":
				msgText = headTxt + yt.FillMsgForSecondRound()
			}
			//Отправлем подготовленное сообщение
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, msgText)
			msg.ParseMode = "HTML"
			sentMsg, err := bot.Send(msg)
			if err != nil {
				log.Println(err)
			}
			log.Println(sentMsg.Text)

		}
	}
}

func fillMsgForStart(chatFirstName, chatLastName string) string {
	return fmt.Sprintf("Hello %s %s!\nYou can use commands:\n /introduction_stage\nand\n /1_round\n"+
		"to get current counts of view on YouTube performance videos for show "+
		"<b>Kingdom: Legendary War</b>", chatFirstName, chatLastName)
}
