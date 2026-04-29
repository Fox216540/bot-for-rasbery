package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func keyboard() tgbotapi.ReplyKeyboardMarkup {
	kb := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("/info"),
			tgbotapi.NewKeyboardButton("/status"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Добавить заметку"),
			tgbotapi.NewKeyboardButton("Поиск по смыслу"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Показать заметки"),
			tgbotapi.NewKeyboardButton("Удалить заметку"),
			tgbotapi.NewKeyboardButton("/reboot"),
		),
	)
	kb.ResizeKeyboard = true
	return kb
}

func send(bot *tgbotapi.BotAPI, chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard()
	_, _ = bot.Send(msg)
}

func sendWithoutKeyboard(bot *tgbotapi.BotAPI, chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	_, _ = bot.Send(msg)
}

func indentLines(text, prefix string) string {
	if text == "" {
		return ""
	}
	lines := strings.Split(text, "\n")
	for i := range lines {
		lines[i] = prefix + lines[i]
	}
	return strings.Join(lines, "\n")
}

func handleMessage(bot *tgbotapi.BotAPI, cfg Config, store *NotesStore, state *userState, msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	text := strings.TrimSpace(msg.Text)
	userID := msg.From.ID

	if pending := state.get(userID); pending != actionNone {
		switch pending {
		case actionAdd:
			if text == "" {
				send(bot, chatID, "Пустой текст, попробуй еще раз.")
				return
			}
			n, err := store.Add(text)
			if err != nil {
				send(bot, chatID, "Не удалось сохранить заметку.")
				return
			}
			if err := aiAdd(cfg.AIURL, n); err != nil {
				send(bot, chatID, "Сохранил локально, но AI сервис недоступен.")
				state.set(userID, actionNone)
				return
			}
			state.set(userID, actionNone)
			send(bot, chatID, "Заметка сохранена. ID: "+n.ID)
			return
		case actionSearch:
			res, err := aiSearch(cfg.AIURL, text)
			state.set(userID, actionNone)
			if err != nil {
				send(bot, chatID, "Ошибка AI поиска.")
				return
			}
			if len(res) == 0 {
				send(bot, chatID, "Ничего не найдено.")
				return
			}
			lines := []string{"Результаты поиска:"}
			for i, r := range res {
				lines = append(lines, fmt.Sprintf("%d. %s | %.3f | %s", i+1, r.ID, r.Score, r.Text))
			}
			send(bot, chatID, strings.Join(lines, "\n"))
			return
		case actionDelete:
			id := strings.TrimSpace(text)
			if id == "" {
				send(bot, chatID, "Пустой ID, попробуй еще раз.")
				return
			}
			deleted, err := store.DeleteByID(id)
			state.set(userID, actionNone)
			if err != nil {
				send(bot, chatID, "Не удалось удалить заметку локально.")
				return
			}
			if !deleted {
				send(bot, chatID, "Заметка с таким ID не найдена.")
				return
			}
			if err := aiDelete(cfg.AIURL, id); err != nil {
				send(bot, chatID, "Локально удалил, но AI сервис недоступен.")
				return
			}
			send(bot, chatID, "Заметка удалена.")
			return
		}
	}

	switch text {
	case "/start", "/help":
		send(bot, chatID, "Выбери действие кнопкой.")
	case "/info":
		send(bot, chatID, fmt.Sprintf("External IP: %s\nLocal IP: %s\nSSID: %s", getExternalIP(), getLocalIP(), getSSID()))
	case "/status":
		send(
			bot,
			chatID,
			fmt.Sprintf(
				"Uptime: %s\nCPU:\n  Cores: %d\n  Load: %s\n  Temp: %s\nMemory:\n%s",
				getUptime(),
				cpuCores(),
				cpuUsagePercent(),
				cpuTemperature(),
				indentLines(getMemory(), "  "),
			),
		)
	case "/reboot":
		send(bot, chatID, "Rebooting...")
		go func() {
			if err := runReboot(cfg.RebootCommand); err != nil {
				log.Printf("reboot command failed: %v", err)
			}
		}()
	case "Добавить заметку":
		state.set(userID, actionAdd)
		sendWithoutKeyboard(bot, chatID, "Отправь текст заметки следующим сообщением.")
	case "Поиск по смыслу":
		state.set(userID, actionSearch)
		sendWithoutKeyboard(bot, chatID, "Отправь запрос следующим сообщением.")
	case "Удалить заметку":
		items := store.Last(10)
		if len(items) == 0 {
			send(bot, chatID, "Заметок пока нет.")
			return
		}
		lines := []string{"Отправь ID заметки для удаления. Последние заметки:"}
		for i, n := range items {
			lines = append(lines, fmt.Sprintf("%d. %s | %s", i+1, n.ID, n.Text))
		}
		state.set(userID, actionDelete)
		sendWithoutKeyboard(bot, chatID, strings.Join(lines, "\n"))
	case "Показать заметки":
		items := store.Last(10)
		if len(items) == 0 {
			send(bot, chatID, "Заметок пока нет.")
			return
		}
		lines := []string{"Последние заметки:"}
		for i, n := range items {
			lines = append(lines, fmt.Sprintf("%d. %s [%s] %s", i+1, n.ID, n.CreatedAt.Format(time.RFC3339), n.Text))
		}
		send(bot, chatID, strings.Join(lines, "\n"))
	default:
		send(bot, chatID, "Используй кнопки ниже.")
	}
}

func processUpdates(bot *tgbotapi.BotAPI, cfg Config, store *NotesStore, state *userState, updates tgbotapi.UpdatesChannel) {
	for update := range updates {
		if update.Message == nil || update.Message.From == nil {
			continue
		}
		if update.Message.From.ID != cfg.AllowedUserID {
			continue
		}
		if update.Message.Text == "" {
			continue
		}
		handleMessage(bot, cfg, store, state, update.Message)
	}
	log.Println("updates channel closed")
}
