package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Service struct {
	db           Repository
	bot          *tgbotapi.BotAPI
	linkLimit    int
	shortLinkLen int
}

func (s *Service) shortURLHandler(w http.ResponseWriter, r *http.Request) {
	shortURL := mux.Vars(r)["shortURL"]
	link, err := s.db.GetLinkByShortURL(shortURL)
	if err != nil {
		log.Printf("unable to get link by short url %s: %v\n", shortURL, err)
		http.Redirect(w, r, "https://google.com", http.StatusPermanentRedirect)
		return
	}
	link.ClickCount++
	if err := s.db.UpdateLink(link); err != nil {
		log.Printf("unable to update link %d: %v\n", link.ID, err)
	}
	http.Redirect(w, r, *link.URL, http.StatusPermanentRedirect)
}

func main() {
	token := flag.String("t", "", "tg api token")
	linkLimit := flag.Int("l", 5, "link limit in manual")
	dsn := flag.String("d", "postgres://postgres:1234@localhost:5432/postgres?sslmode=disable", "database dsn")
	flag.Parse()

	bot, err := tgbotapi.NewBotAPI(*token)
	if err != nil {
		log.Fatalf("unable to initialize tg client: %v", err)
	}
	db, err := sql.Open("postgres", *dsn)
	if err != nil {
		log.Fatalf("unable to open db conn: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("ping failed: %v", err)
	}

	s := &Service{
		bot:          bot,
		db:           NewRepository(db),
		linkLimit:    *linkLimit,
		shortLinkLen: 8,
	}

	r := mux.NewRouter()
	r.HandleFunc("/sl/{shortURL}", s.shortURLHandler).Methods(http.MethodGet)

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	go s.handleBotUpdates(bot.GetUpdatesChan(u))

	log.Println("starting server on port 443...")

	log.Fatal(http.ListenAndServeTLS("0.0.0.0:443", "./certs/cert", "./certs/key", r))
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
				if err := s.ShowDeleteLink(update.CallbackQuery.From.ID); err != nil {
					log.Printf("unable to add link: %v\n", err)
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

				var short string = generateRandomString(s.shortLinkLen)
				_, err = s.db.GetLinkByShortURL(short)
				for i := 0; err == sql.ErrNoRows && i < 10; i++ {
					short = generateRandomString(s.shortLinkLen)
					_, err = s.db.GetLinkByShortURL(short)
				}
				link.ShortURL = &short

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

func generateRandomString(len int) string {
	res := make([]byte, len)
	rand.Read(res)
	return string(res)
}
