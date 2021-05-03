package main

import (
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strings"

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
		log.Println(err)
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
			var headerTxt string = "Current count of views and likes on videos for\n"
			var headTxt string = fmt.Sprintf("%18s|%15s|%15s|\t%s\n",
				"Views", "Likes", "Dislikes", "Name")

			switch update.Message.Text {
			case "/start":
				msgText = fillMsgForStart(update.Message.Chat.FirstName, update.Message.Chat.LastName)

			case "/update":
				yt.UpdateVideoIDList()
				msgText = "Video ids was updated"

			case "/introduction":
				msgText = headerTxt + "<b>INTRODUCTION STAGE</b>\n"
				msgText += headTxt + yt.FillMsgForIntro()

			case "/1round":
				msgText = headerTxt + "<b>1 round: TO THE WORLD</b>\n"
				msgText += headTxt + yt.FillMsgForFirstRound()

			case "/2round":
				msgText = headerTxt + "<b>2 round: RE-BORN</b>\n"
				msgText += headTxt + yt.FillMsgForSecondRound()

			case "/3round":
				msgText = headerTxt + "<b>3 round: - </b>\n"
				msgText += headTxt + yt.FillMsgForThirdRound()

			default:
				// make requst for video
				if strings.HasPrefix(update.Message.Text, "url ") {
					url := strings.Split(update.Message.Text, " ")
					msgText += yt.FillMsgForVideo(url[1])
				} else {
					msgText = "Can not recognise command"
				}
			}
			//Отправлем подготовленное сообщение
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, msgText)
			msg.ParseMode = "HTML"
			_, err := bot.Send(msg)
			if err != nil {
				log.Println(err)
			}
			// log.Println(sentMsg.Text)
		}
	}
}

func fillMsgForStart(chatFirstName, chatLastName string) string {
	return fmt.Sprintf("Hello %s %s!\nYou can use commands:\n /introduction_stage\nand\n /1_round\n"+
		"to get current counts of view on YouTube performance videos for show "+
		"<b>Kingdom: Legendary War</b>", chatFirstName, chatLastName)
}
