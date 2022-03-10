package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/url"
	"strings"

	_ "github.com/lib/pq"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Service struct {
	db        Repository
	bot       *tgbotapi.BotAPI
	linkLimit int
}

func main() {
	token := flag.String("t", "", "tg api token")
	linkLimit := flag.Int("l", 5, "link limit in manual")
	flag.Parse()

	bot, err := tgbotapi.NewBotAPI(*token)
	if err != nil {
		log.Fatalf("unable to initialize tg client: %v", err)
	}
	db, err := sql.Open("postgres", "store.db")
	if err != nil {
		log.Fatalf("unable to open db conn: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("ping failed: %v", err)
	}

	s := &Service{
		bot:       bot,
		db:        NewRepository(db),
		linkLimit: *linkLimit,
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	s.handleBotUpdates(bot.GetUpdatesChan(u))
}

func (s *Service) handleBotUpdates(updates tgbotapi.UpdatesChannel) {
	for update := range updates {
		if update.CallbackQuery != nil {
			if update.CallbackQuery.Data == "showManual" {
				if err := s.ShowManual(update.CallbackQuery.From.ID); err != nil {
					log.Printf("unable to show manual: %v\n", err)
				}
				continue
			}
			if update.CallbackQuery.Data == "addLink" {
				if err := s.ShowAddLink(update.CallbackQuery.From.ID); err != nil {
					log.Printf("unable to add link: %v\n", err)
				}
				continue
			}
			if update.CallbackQuery.Data == "deleteLink" {
				if err := s.db.DeleteLinkByUserID(update.CallbackQuery.From.ID); err != nil {
					log.Printf("unable to delete link: %v\n", err)
				}
				continue
			}
		}
		if update.Message != nil { // If we got a message
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			link, err := s.db.GetLinkByUserID(update.Message.Chat.ID)
			if err != nil && err != sql.ErrNoRows {
				log.Printf("unable to get link: %v\n", err)
				continue
			}
			if link != nil && link.URL != nil {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Вы уже добавили свою ссылку")
				msg.DisableWebPagePreview = true
				msg.ReplyMarkup = GetKeyboard(KeyboardModeDeleteLink)
				s.bot.Send(msg)
				continue
			}

			if link != nil && link.URL == nil && !update.Message.IsCommand() {
				urlLink, err := url.Parse(update.Message.Text)
				if err != nil || !strings.Contains(urlLink.Host, "linkedin") {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Некорректная ссылка! Попробуйте еще раз")
					s.bot.Send(msg)
					continue
				}
				url := update.Message.Text
				link.URL = &url

				if err := s.db.UpdateLink(link); err != nil {
					log.Printf("unable to update link: %v", err)
					continue
				}
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ваша ссылка добавлена, спасибо!"+endText)

				msg.ReplyMarkup = GetKeyboard(KeyboardModeDeleteLink)
				msg.DisableWebPagePreview = true

				s.bot.Send(msg)
				continue
			}

			if update.Message.Command() == "start" {
				linkCount, err := s.db.GetLinkCount()
				if err != nil {
					log.Printf("unable to get link count: %v", err)
					continue
				}
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf(welcomeText, linkCount))
				msg.ReplyMarkup = GetKeyboard(KeyboardModeShowManual)
				msg.DisableWebPagePreview = true

				s.bot.Send(msg)
			}
		}
	}
}
