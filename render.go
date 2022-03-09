package main

import (
	"database/sql"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func DeleteLink(bot *tgbotapi.BotAPI, db *sql.DB, userID int64) error {
	if err := DeleteLinkByUser(db, userID); err != nil {
		return fmt.Errorf("unable to delete link: %v", err)
	}
	_, err := bot.Send(tgbotapi.NewMessage(userID, "Ваша ссылка удалена"))
	return err
}

func AddLink(bot *tgbotapi.BotAPI, db *sql.DB, userID int64) error {
	msg := tgbotapi.NewMessage(userID, "Введите урл на ваш пост с поиском работы:")
	if err := InsertLink(db, userID, "0"); err != nil {
		return fmt.Errorf("unable to insert link: %v", err)
	}
	_, err := bot.Send(msg)
	return err
}

func ShowManual(bot *tgbotapi.BotAPI, db *sql.DB, userID int64) error {
	var linkText string
	var i int
	links, err := ListAllLinks(db)
	if err != nil {
		return fmt.Errorf("unable to list links: %v", err)
	}
	// order of iterating here is random because of implementation of 'map' in Go
	for _, link := range links {
		if i >= 5 {
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

	msg := tgbotapi.NewMessage(userID, fmt.Sprintf(manualText, linkText))

	msg.ReplyMarkup = GetKeyboard(false)
	msg.DisableWebPagePreview = true

	if err := InsertLink(db, userID, "0"); err != nil {
		return fmt.Errorf("unable to insert links: %v", err)
	}
	_, err = bot.Send(msg)
	return err
}

func GetKeyboard(linkIsAdded bool) tgbotapi.InlineKeyboardMarkup {
	firstButton := tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Добавить ссылку", "addLink"),
	)
	if linkIsAdded {
		firstButton = tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Удалить вашу ссылку", "deleteLink"),
		)
	}
	return tgbotapi.NewInlineKeyboardMarkup(
		firstButton,
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("Исходный код", "https://github.com/Barugoo/linknetwork"),
			tgbotapi.NewInlineKeyboardButtonURL("Чат для общения", "https://t.me/+oSNQjFXdNndlYzE6"),
		),
	)
}
