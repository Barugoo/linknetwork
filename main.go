package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"strings"

	_ "github.com/mattn/go-sqlite3"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var welcomeText = `
Привет! 
Возникла идея организовать бота для взаимолайков в Linkedin с целью повышения количества просмотров профилей коллег, которые ищут работу с релокацией. 

Эффект от лайков значителен, так как ваш пост вероятно окажется в ленте всех контактов лайкнувшего человека, среди которых обычно бывают рекрутеры. 
Таким образом мы сможем помочь друг-другу поскорее найти работу.

🙅‍♀️ Мы не сохраняем ничего, кроме ссылки и айдишника пользователя для идентификации. Исходники доступны по кнопке👇

Тем временем уже %d коллег добавили ссылки😍`

var manualText = `
Если вы хотите в этом поучавствовать вам нужно сделать следующее:

1. Опубликовать в linkedin пост о том, что вы ищите работу с релокацией (желательно на английском) и пометить свой профиль как открытый для поиска работы

2. Мы просим вас пролайкать посты ваших коллег и добавить их в контакты, чтобы повысить охват. Лайкать/добавлять или нет остается на ваше усмотрение%s

3. Затем нажать кнопочку "Добавить ссылку" и вам будет предложено ввести ссылку на пост из пункта 1. После этого он начнет появляться в блоке ссылок для других участников данного флешмоба

После добавления ссылки, при желании вы ее сможете удалить из выдачи. Для этого снова введите /start - появится кнопка для удаления
`

var endText = `
Так же было бы круто, если бы вы добавились в наш чат для общения - будем держаться вместе! (кнопка для перехода в чат внизу)
Друзья, давайте поможем друг-другу найти работу в это нелегкое время! Всем мир.❤️
`

func main() {
	bot, err := tgbotapi.NewBotAPI("token")
	if err != nil {
		log.Fatalf("unable to initialize tg client: %v", err)
	}
	db, err := sql.Open("sqlite3", "store.db")
	if err != nil {
		log.Fatalf("unable to open db conn: %v", err)
	}
	InitDB(db)

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.CallbackQuery != nil {
			if update.CallbackQuery.Data == "showManual" {
				if err := ShowManual(bot, db, update.CallbackQuery.From.ID); err != nil {
					log.Printf("unable to show manual: %v\n", err)
				}
				continue
			}
			if update.CallbackQuery.Data == "addLink" {
				if err := AddLink(bot, db, update.CallbackQuery.From.ID); err != nil {
					log.Printf("unable to add link: %v\n", err)
				}
				continue
			}
			if update.CallbackQuery.Data == "deleteLink" {
				if err := DeleteLink(bot, db, update.CallbackQuery.From.ID); err != nil {
					log.Printf("unable to delete link: %v\n", err)
				}
				continue
			}
		}
		if update.Message != nil { // If we got a message
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			link, err := GetLinkByUser(db, update.Message.Chat.ID)
			if err != nil && err != sql.ErrNoRows {
				log.Printf("unable to get link: %v", err)
				continue
			}
			if link != nil && *link != "0" {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Вы уже добавили свою ссылку")
				msg.DisableWebPagePreview = true
				msg.ReplyMarkup = GetKeyboard(true)
				bot.Send(msg)
				continue
			}

			if link != nil && *link == "0" && !update.Message.IsCommand() {
				urlLink, err := url.Parse(update.Message.Text)
				if err != nil || !strings.Contains(urlLink.Host, "linkedin") {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Некорректная ссылка! Попробуйте еще раз")
					bot.Send(msg)
					continue
				}
				if err := UpdateLinkByUser(db, update.Message.Chat.ID, update.Message.Text); err != nil {
					log.Printf("unable to update link: %v", err)
					continue
				}
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ваша ссылка добавлена, спасибо!"+endText)

				msg.ReplyMarkup = GetKeyboard(true)
				msg.DisableWebPagePreview = true

				bot.Send(msg)
				continue
			}

			if update.Message.Command() == "start" {
				linkCount, err := LinkCount(db)
				if err != nil {
					log.Printf("unable to get link count: %v", err)
					continue
				}
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf(welcomeText, linkCount))
				msg.ReplyMarkup = GetKeyboard(false)
				msg.DisableWebPagePreview = true

				bot.Send(msg)
			}
		}
	}
}
