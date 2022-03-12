package main

import (
	"database/sql"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (s *Service) ShowDeleteLink(userID int64) error {
	if err := s.db.DeleteLinkByUserID(userID); err != nil {
		return fmt.Errorf("unable to delete link: %w", err)
	}
	msg := tgbotapi.NewMessage(userID, "Ваша ссылка удалена")
	msg.ReplyMarkup = GetKeyboard(KeyboardModeAddLink)
	_, err := s.bot.Send(msg)
	return err
}

func (s *Service) ShowAddLink(userID int64) error {
	msg := tgbotapi.NewMessage(userID, "Введите урл на ваш пост с поиском работы:")
	link := &Link{
		UserID:     userID,
		ClickCount: 0,
	}
	s.db.CreateLink(link)

	_, err := s.bot.Send(msg)
	return err
}

func (s *Service) ShowManual(userID int64) error {
	var linkText string
	links, err := s.db.ListAllLinks(s.linkLimit)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("unable to list links: %w", err)
	}
	for _, link := range links {
		if link != nil && link.ShortURL != nil {
			linkText += fmt.Sprintf("\n - %s", fmt.Sprintf("https://linknetworkbot.com/sl/%s", *link.ShortURL))
		}
	}
	if len(links) == 0 {
		linkText = "Пока ссылок нет!"
	}

	msg := tgbotapi.NewMessage(userID, fmt.Sprintf(manualText, linkText))

	msg.ReplyMarkup = GetKeyboard(KeyboardModeAddLink)
	msg.DisableWebPagePreview = true

	_, err = s.bot.Send(msg)
	return err
}

type KeyboardMode int8

const (
	KeyboardModeAddLink KeyboardMode = iota
	KeyboardModeDeleteLink
	KeyboardModeShowManual
)

func GetKeyboard(keyboardMode KeyboardMode) tgbotapi.InlineKeyboardMarkup {
	var button []tgbotapi.InlineKeyboardButton
	switch keyboardMode {
	case KeyboardModeAddLink:
		button = tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Добавить ссылку", "addLink"),
		)
	case KeyboardModeDeleteLink:
		button = tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Удалить вашу ссылку", "deleteLink"),
		)
	case KeyboardModeShowManual:
		button = tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Хочу участвовать", "showManual"),
		)
	}
	return tgbotapi.NewInlineKeyboardMarkup(
		button,
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("Исходный код", "https://github.com/Barugoo/linknetwork"),
			tgbotapi.NewInlineKeyboardButtonURL("Чат для общения", "https://t.me/+oSNQjFXdNndlYzE6"),
		),
	)
}

var welcomeText = `
Привет!👋
Возникла идея организовать бота для взаимолайков в Linkedin с целью повышения количества просмотров профилей коллег, которые ищут работу с релокацией. 

Эффект от лайков значителен, так как ваш пост вероятно окажется в ленте всех контактов лайкнувшего человека, среди которых обычно бывают рекрутеры. 
Таким образом мы сможем помочь друг-другу поскорее найти работу.

🙅‍♀️ Мы не сохраняем ничего, кроме ссылки и айдишника пользователя для идентификации. Исходники доступны по кнопке👇

Тем временем уже %d коллег добавили ссылки😍`

var manualText = `
Если вы хотите в этом поучаствовать вам нужно сделать следующее:

1. Опубликовать в linkedin пост о том, что вы ищите работу с релокацией (желательно на английском) и пометить свой профиль как открытый для поиска работы

2. Мы просим вас пролайкать посты ваших коллег и добавить их в контакты, чтобы повысить охват. Лайкать/добавлять или нет остается на ваше усмотрение. Здесь выведены ссылки с наименьшим количеством кликов в данный момент:%s

3. Затем нажать кнопочку "Добавить ссылку" и вам будет предложено ввести ссылку на пост из пункта 1. На бэке мы прогоняем вашу ссылку через сокращатель, чтобы экономить место и собирать статистику по кликам (только количество)

После добавления ссылки, при желании вы ее сможете удалить из выдачи. Для этого снова введите /start - появится кнопка для удаления
`

var endText = `
Так же было бы круто, если бы вы добавились в наш чат для общения - будем держаться вместе! (кнопка для перехода в чат внизу)
Друзья, давайте поможем друг-другу найти работу в это нелегкое время! Всем мир.❤️
`
