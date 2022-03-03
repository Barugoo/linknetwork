package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var welcomeText = `
Привет! 
Возникла идея организовать бота для взаимолайков в Linkedin с целью повышения количества просмотров профилей коллег, которые ищут работу с релокацией. Эффект от лайков значителен, так как ваш пост вероятно окажется в ленте всех контактов лайкнувшего человека, среди которых обычно бывают рекрутеры. Таким образом мы сможем помочь друг-другу поскорее найти работу. Если вы хотите в этом поучавствовать вам нужно сделать следующее:

1. Опубликовать в linkedin пост о том, что вы ищите работу с релокацией (желательно на английском) и пометить свой профиль как открытый для поиска работы

2. Мы просим вас пролайкать посты ваших коллег и добавить их в контакты, чтобы повысить охват. Лайкать/добавлять или нет остается на ваше усмотрение%s

3. Затем нажать кнопочку "Добавить ссылку" и вам будет предложено ввести ссылку на пост из пункта 1. После этого он начнет появляться в блоке ссылок для других участников данного флешмоба

После добавления ссылки, при желании вы ее сможете удалить из выдачи. Для этого снова введите /start - появится кнопка для удаления

Так же было бы круто, если бы вы добавились в наш чат для общения - будем держаться вместе! (кнопка для перехода в чат внизу)
Друзья, давайте поможем друг-другу найти работу в это нелегкое время! Всем мир.❤️

🚨 DISCLAIMER: есть инфа, что Linkedin борется с автоматизацией, поэтому в целях безопасности количество выдаваемых ссылок ограничено до 10 (перемешиваются рандомно). В любом случае, лайки и коннекты делаются ручками реальных людей с реальных аккаунтов, поэтому (как мне кажется) вряд ли последуют санкции за взаимолайк.`

var links = map[string]string{}

func InsertLink(db *sql.DB, userID, link string) error {
	_, err := db.Exec("INSERT INTO links (userID, link) VALUES (?, ?)", userID, link)
	return err
}

func GetLinkByUser(db *sql.DB, userID string) (*string, error) {
	row := db.QueryRow("SELECT link FROM links WHERE userID = ?", userID)

	var res string
	if err := row.Scan(&res); err != nil {
		return nil, err
	}
	return &res, nil
}

func UpdateLinkByUser(db *sql.DB, userID, newLink string) error {
	_, err := db.Exec("UPDATE links SET link = ? WHERE userID = ?", newLink, userID)
	return err
}

func DeleteLinkByUser(db *sql.DB, userID string) error {
	_, err := db.Exec("DELETE FROM links WHERE userID = ?", userID)
	return err
}

func GetLinkCount(db *sql.DB) (int64, error) {
	row := db.QueryRow("SELECT COUNT() FROM links")

	var res int64
	if err := row.Scan(&res); err != nil {
		return 0, err
	}
	return res, nil
}

func ListAllLinks(db *sql.DB) (map[string]string, error) {
	rows, err := db.Query("SELECT userID, link FROM links")
	if err != nil {
		return nil, err
	}

	m := make(map[string]string)

	var userID, link string
	for rows.Next() {
		rows.Scan(&userID, &link)
		m[userID] = link
	}

	return m, nil
}

func main() {
	bot, err := tgbotapi.NewBotAPI("token")
	if err != nil {
		log.Fatalf("unable to initialize tg client: %v", err)
	}
	db, err := sql.Open("sqlite3", "store.db")
	if err != nil {
		log.Fatalf("unable to open db conn: %v", err)
	}
	db.Exec(`CREATE TABLE links(
		id INTEGER PRIMARY KEY AUTOINCREMENT, 
		userID TEXT,
		link TEXT
	  );`)

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.CallbackQuery != nil {
			if update.CallbackQuery.Data == "addLink" {
				userID := strconv.FormatInt(update.CallbackQuery.From.ID, 10)
				msg := tgbotapi.NewMessage(update.CallbackQuery.From.ID, "Введите урл на ваш пост с поиском работы:")
				if err := InsertLink(db, userID, "0"); err != nil {
					log.Printf("unable to insert link: %v", err)
					continue
				}
				bot.Send(msg)
				continue
			}
			if update.CallbackQuery.Data == "deleteLink" {
				userID := strconv.FormatInt(update.CallbackQuery.From.ID, 10)
				if err := DeleteLinkByUser(db, userID); err != nil {
					log.Printf("unable to delete link: %v", err)
					continue
				}
				msg := tgbotapi.NewMessage(update.CallbackQuery.From.ID, "Ваша ссылка удалена")
				bot.Send(msg)
				continue
			}
		}
		if update.Message != nil { // If we got a message
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			userID := strconv.FormatInt(update.Message.Chat.ID, 10)

			link, err := GetLinkByUser(db, userID)
			if err != nil && err != sql.ErrNoRows {
				log.Printf("unable to get link: %v", err)
				continue
			}
			if link != nil && *link != "0" {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Вы уже добавили свою ссылку")
				keyboard := tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Удалить вашу ссылку", "deleteLink"),
					),
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonURL("Чат для общения", "https://t.me/+oSNQjFXdNndlYzE6"),
					),
				)
				msg.DisableWebPagePreview = true
				msg.ReplyMarkup = keyboard
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
				if err := UpdateLinkByUser(db, userID, update.Message.Text); err != nil {
					log.Printf("unable to update link: %v", err)
					continue
				}
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ваша ссылка добавлена, спасибо!")
				bot.Send(msg)
				continue
			}

			if update.Message.Command() == "start" {
				var linkText string

				var i int
				links, err := ListAllLinks(db)
				if err != nil {
					log.Printf("unable to list links: %v", err)
					continue
				}
				for _, link := range links {
					if i >= 10 {
						break
					}
					if link == "0" {
						continue
					}
					i++
					linkText += fmt.Sprintf("\n - %s", link)
				}
				if len(links) == 0 {
					linkText = "Пока ссылок нет!"
				}

				msgText := fmt.Sprintf(welcomeText, linkText)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, msgText)

				keyboard := tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Добавить ссылку", "addLink"),
					),
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonURL("Чат для общения", "https://t.me/+oSNQjFXdNndlYzE6"),
					),
				)
				msg.ReplyMarkup = keyboard
				msg.DisableWebPagePreview = true

				bot.Send(msg)
			}

		}
	}
}
