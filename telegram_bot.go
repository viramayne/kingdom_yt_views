package main

import (
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strings"

	tgbotapi "github.com/Syfaro/telegram-bot-api"
)

var stageKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("introduction stage", "/introduction"),
		tgbotapi.NewInlineKeyboardButtonData("1 round", "/1round"),
	),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("2 round", "/2round"),
		tgbotapi.NewInlineKeyboardButtonData("3 round", "/3round"),
	),
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

		// Запрос от inline button
		if update.CallbackQuery != nil {
			fmt.Print(update)

			bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data))

			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Data)
			msg.ParseMode = "HTML"

			switch update.CallbackQuery.Data {
			case "/introduction":
				msg.Text = yt.introMsg()
				msg.ReplyMarkup = stageKeyboard

			case "/1round":
				msg.Text = yt.firstRoundMsg()
				msg.ReplyMarkup = stageKeyboard

			case "/2round":
				msg.Text = yt.secondRoundMsg()
				msg.ReplyMarkup = stageKeyboard

			case "/3round":
				msg.Text = yt.thirdRoundMsg()
				msg.ReplyMarkup = stageKeyboard
			}
			// Ответ на запрос inline query button
			bot.Send(msg)
		}

		// Если от пользователя пришло текстовое сообщение, в т.ч. комманды
		if update.Message != nil {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)

			//Проверяем что от пользователья пришло именно текстовое сообщение
			if reflect.TypeOf(update.Message.Text).Kind() == reflect.String && update.Message.Text != "" {

				switch update.Message.Text {
				case "/start":
					msg.Text = fillMsgForStart(update.Message.Chat.FirstName, update.Message.Chat.LastName)
					msg.ReplyMarkup = stageKeyboard

				case "/update":
					yt.UpdateVideoIDList()
					msg.Text = "Video ids was updated"
					msg.ReplyMarkup = stageKeyboard

				case "/introduction":
					msg.Text = yt.introMsg()
					msg.ReplyMarkup = stageKeyboard

				case "/1round":
					msg.Text = yt.firstRoundMsg()
					msg.ReplyMarkup = stageKeyboard

				case "/2round":
					msg.Text = yt.secondRoundMsg()
					msg.ReplyMarkup = stageKeyboard

				case "/3round":
					msg.Text = yt.thirdRoundMsg()
					msg.ReplyMarkup = stageKeyboard

				default:
					// make requst for video
					if strings.HasPrefix(update.Message.Text, "url ") {
						url := strings.Split(update.Message.Text, " ")
						msg.Text += yt.FillMsgForVideo(url[1])
						msg.ReplyMarkup = stageKeyboard
					} else {
						msg.Text = "Can not recognise command"
					}
				}
				//Отправлем подготовленное сообщение
				msg.ParseMode = "HTML"
				_, err := bot.Send(msg)
				if err != nil {
					log.Println(err)
				}
			}
		}
	}
}

func fillMsgForStart(chatFirstName, chatLastName string) string {
	return fmt.Sprintf("Hello %s %s!", chatFirstName, chatLastName)
}
