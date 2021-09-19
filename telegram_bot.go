package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"sort"
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
		log.Println(err)
	}
	fmt.Printf("Authorized on account %s\n", bot.Self.UserName)

	// Инициализируем связь с YouTube
	yt, err := NewYTStatistics()
	if err != nil {
		log.Println(err)
	}
	db, err := ConnectToDB()
	if err != nil {
		log.Println(err)
	}
	defer db.Close()
	yt.Db = db

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
		log.Println(err)
	}

	go http.ListenAndServe(":"+Port, nil)
	log.Printf("start listen :%s", Port)

	for update := range updates {
		yt.UpdateVideoStatistics()

		// Запрос от inline button
		if update.CallbackQuery != nil {
			fmt.Print(update)
			yt.UpdateVideoStatistics()
			bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data))

			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Data)
			msg.ParseMode = "HTML"

			switch update.CallbackQuery.Data {
			case "/introduction":
				msg.Text, msg.ReplyMarkup = yt.introMsg()

			case "/1round":
				msg.Text, msg.ReplyMarkup = yt.firstRoundMsg()

			case "/2round":
				msg.Text, msg.ReplyMarkup = yt.secondRoundMsg()

			case "/3round":
				msg.Text, msg.ReplyMarkup = yt.thirdRoundMsg()

			case "ateez":
				msg.Text = yt.ateezMsg()

			case "skz":
				msg.Text = yt.skzMsg()

			}
			// Ответ на запрос inline query button
			bot.Send(msg)
		}

		// Если от пользователя пришло текстовое сообщение, в т.ч. комманды
		if update.Message != nil {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
			yt.UpdateVideoStatistics()
			//Проверяем что от пользователья пришло именно текстовое сообщение
			if reflect.TypeOf(update.Message.Text).Kind() == reflect.String && update.Message.Text != "" {

				switch update.Message.Text {
				case "/start":
					msg.Text = fillMsgForStart(update.Message.Chat.FirstName, update.Message.Chat.LastName)
					msg.ReplyMarkup = stageKeyboard

				case "/update":
					yt.UpdateVideoIDList()
					// Перезапись видео в БД

					msg.Text = "Video ids was updated"
					msg.ReplyMarkup = stageKeyboard

				case "/introduction":
					msg.Text, msg.ReplyMarkup = yt.introMsg()

				case "/1round":
					msg.Text, msg.ReplyMarkup = yt.firstRoundMsg()

				case "/2round":
					msg.Text, msg.ReplyMarkup = yt.secondRoundMsg()

				case "/3round":
					msg.Text, msg.ReplyMarkup = yt.thirdRoundMsg()

				case "ateez":
					msg.Text = yt.ateezMsg()

				case "skz":
					msg.Text = yt.skzMsg()

				default:
					// make requst for video
					if strings.HasPrefix(update.Message.Text, "url ") {
						url := strings.Split(update.Message.Text, " ")
						msg.Text += VideoByURL(url[1], yt)
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

func (yt *YTStat) UpdateVideoStatistics() {
	// get videos from db.videos
	videoIds, err := yt.GetVideoIds()
	if err != nil {
		log.Println(err)
	}

	// make request to yt api
	resp, err := yt.getStatistics(videoIds)
	if err != nil {
		log.Println(err)
	}

	// save data to db.statistics
	if err = yt.WriteToDBStatistics(resp); err != nil {
		log.Println(err)
	}
}

func fillMsgForDate(yt *YTStat, date string) (string, tgbotapi.InlineKeyboardMarkup) {
	text, btns, err := formMsgForDate(date, yt)
	if err != nil {
		log.Println(err)
	}
	rows := make([][]tgbotapi.InlineKeyboardButton, 0)
	for _, btn := range btns {
		row := make([]tgbotapi.InlineKeyboardButton, 0)
		row = append(row, btn)
		rows = append(rows, row)
	}
	return text, tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func FirstRound(yt *YTStat) (string, tgbotapi.InlineKeyboardMarkup) {
	text1, keyboard1 := fillMsgForDate(yt, "2021-04-08")

	text2, keyboard2 := fillMsgForDate(yt, "2021-04-15")

	bttns := keyboard1.InlineKeyboard
	bttns = append(bttns, keyboard2.InlineKeyboard...)

	return text1 + text2, tgbotapi.NewInlineKeyboardMarkup(bttns...)
}

func SecondRound(yt *YTStat) (string, tgbotapi.InlineKeyboardMarkup) {
	text1, keyboard1 := fillMsgForDate(yt, "2021-04-22")

	text2, keyboard2 := fillMsgForDate(yt, "2021-04-29")

	bttns := keyboard1.InlineKeyboard
	bttns = append(bttns, keyboard2.InlineKeyboard...)

	return text1 + text2, tgbotapi.NewInlineKeyboardMarkup(bttns...)
}

func ThirdRound(yt *YTStat) (string, tgbotapi.InlineKeyboardMarkup) {
	text1, keyboard1 := fillMsgForDate(yt, "2021-05-13")

	text2, keyboard2 := fillMsgForDate(yt, "2021-05-20")

	text3, keyboard3 := fillMsgForDate(yt, "2021-05-27")

	// collect all buttons to one keyboard
	bttns := keyboard1.InlineKeyboard
	bttns = append(bttns, keyboard2.InlineKeyboard...)
	bttns = append(bttns, keyboard3.InlineKeyboard...)

	return text1 + text2 + text3, tgbotapi.NewInlineKeyboardMarkup(bttns...)
}

func AteezVideos(yt *YTStat) (text string) {
	ids := formIdsSrt(ateezVideoIds[:])
	resp, err := yt.getStatistics(ids)
	if err != nil {
		log.Println(err)
	}
	// form respond msg text
	b := BeautifyNumbers

	if resp != nil && resp.Items != nil {
		for _, v := range resp.Items {
			text += v.Snippet.Title
			text += fmt.Sprintf("</b>\n\n%18s|%15s|%15s\n",
				"Views", "Likes", "Dislikes")
			text += fmt.Sprintf("%15v|%12v|%15v",
				b(v.Statistics.Views), b(v.Statistics.Likes), b(v.Statistics.Dislikes))
		}
	}
	return
}

func SKZVideos(yt *YTStat) string {
	ids := formIdsSrt(skzVideoIds[:])
	resp, err := yt.getStatistics(ids)
	if err != nil {
		log.Println(err)
	}
	// form respond msg text
	b := BeautifyNumbers
	var text string = "\nCurrent statistics for videos: \n"
	if resp != nil && resp.Items != nil {
		// sort.Slice(resp.Items, func(i, j int) (less bool) {
		// 	return resp.Items[i].Statistics.Views > resp.Items[j].Statistics.Views
		// })

		for i, v := range resp.Items {
			text += fmt.Sprintf("%d: %s", i+1, v.Snippet.Title)
			text += fmt.Sprintf("\n%15v|%12v|%15v\n",
				b(v.Statistics.Views), b(v.Statistics.Likes), b(v.Statistics.Dislikes))
		}
	}
	return text
}

func VideoByURL(url string, yt *YTStat) string {
	return yt.formMsgForVideo(url)
}

func formMsgForDate(date string, yt *YTStat) (string, []tgbotapi.InlineKeyboardButton, error) {
	var text string = fmt.Sprintf("\naired %s:\n", date)

	videoIds, err := yt.GetVideosForDateFromBD(date)
	if err != nil || len(videoIds) == 0 {
		return text, []tgbotapi.InlineKeyboardButton{}, errors.New("no video to make request")
	}

	resp, err := yt.ReadFromDBStatistics(videoIds)
	if err != nil {
		return text, []tgbotapi.InlineKeyboardButton{}, err
	}
	sort.Slice(resp, func(i, j int) (less bool) {
		return resp[i].Views > resp[j].Views
	})

	b := BeautifyNumbers
	buttons := make([]tgbotapi.InlineKeyboardButton, len(resp))
	for i, v := range resp {
		title, err := yt.GetVideoTitleFromDB(v.VideoID)
		if err != nil {
			log.Println(err)
			continue
		}
		url := fmt.Sprintf("http://y2u.be/%s\n", v.VideoID)
		button := tgbotapi.InlineKeyboardButton{
			Text: fmt.Sprintf("%2d:\t%s", i+1, title),
			URL:  &url,
		}
		text += fmt.Sprintf("\n%2d:\t%s\n %15v|%12v|%15v\n", i+1, title, b(fmt.Sprint(v.Views)), b(fmt.Sprint(v.Likes)), b(fmt.Sprint(v.Dislikes)))
		buttons[i] = button
	}
	return text, buttons, err
}

func (yt *YTStat) introMsg() (string, tgbotapi.InlineKeyboardMarkup) {
	var headerTxt string = "Current count of views and likes on videos for\n"
	var headTxt string = fmt.Sprintf("%18s|%15s|%15s|\t%s\n", "Name",
		"Views", "Likes", "Dislikes")
	text := headerTxt + "<b>INTRODUCTION STAGE</b>\n" + headTxt

	msgText, replyMarkup := fillMsgForDate(yt, "2021-04-01")
	return text + msgText, replyMarkup
}

func (yt *YTStat) firstRoundMsg() (string, tgbotapi.InlineKeyboardMarkup) {
	var headerTxt string = "Current count of views and likes on videos for\n"
	var headTxt string = fmt.Sprintf("%18s|%15s|%15s|\t%s\n",
		"Views", "Likes", "Dislikes", "Name")
	text := headerTxt + "<b>1 round: TO THE WORLD</b>\n" + headTxt
	msgText, replyMarkup := FirstRound(yt)
	return text + msgText, replyMarkup
}

func (yt *YTStat) secondRoundMsg() (string, tgbotapi.InlineKeyboardMarkup) {
	var headerTxt string = "Current count of views and likes on videos for\n"
	var headTxt string = fmt.Sprintf("%18s|%15s|%15s|\t%s\n",
		"Views", "Likes", "Dislikes", "Name")
	text := headerTxt + "<b>2 round: RE-BORN</b>\n" + headTxt
	msgText, replyMarkup := SecondRound(yt)
	return text + msgText, replyMarkup
}

func (yt *YTStat) thirdRoundMsg() (string, tgbotapi.InlineKeyboardMarkup) {
	var headerTxt string = "Current count of views and likes on videos for\n"
	var headTxt string = fmt.Sprintf("%18s|%15s|%15s|\t%s\n",
		"Views", "Likes", "Dislikes", "Name")
	text := headerTxt + "<b>3 round: - NO LIMIT </b>\n" + headTxt
	msgText, replyMarkup := ThirdRound(yt)
	return text + msgText, replyMarkup
}

func (yt *YTStat) ateezMsg() string {
	var headerTxt string = "Current count of views and likes on videos for\n"
	var headTxt string = fmt.Sprintf("%18s|%15s|%15s|\t%s\n",
		"Views", "Likes", "Dislikes", "Name")
	text := headerTxt + "<b>3 round: - NO LIMIT </b>\n" + headTxt
	msgText := AteezVideos(yt)
	return text + msgText
}

func (yt *YTStat) skzMsg() string {
	var headerTxt string = "Current count of views and likes on videos for\n"
	var headTxt string = fmt.Sprintf("%18s|%15s|%15s|\t%s\n",
		"Views", "Likes", "Dislikes", "Name")
	text := headerTxt + "<b>Stray Kids</b>\n" + headTxt
	msgText := SKZVideos(yt)
	if len(msgText) > 4096 {
		log.Println("more than 4096 characters")
	}
	return text + msgText
}
