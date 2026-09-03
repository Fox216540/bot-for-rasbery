package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func keyboard() tgbotapi.ReplyKeyboardMarkup {
	kb := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Комната"),
			tgbotapi.NewKeyboardButton("Статус"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Лампа"),
			tgbotapi.NewKeyboardButton("Лента"),
			tgbotapi.NewKeyboardButton("ESP32 RGB"),
		),
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

func roomKeyboard() tgbotapi.ReplyKeyboardMarkup {
	kb := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Температура"),
			tgbotapi.NewKeyboardButton("Влажность"),
			tgbotapi.NewKeyboardButton("Давление"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Назад"),
		),
	)
	kb.ResizeKeyboard = true
	return kb
}

func statusKeyboard() tgbotapi.ReplyKeyboardMarkup {
	kb := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Занят"),
			tgbotapi.NewKeyboardButton("Свободен"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Разговариваю"),
			tgbotapi.NewKeyboardButton("Полузанят"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Назад"),
		),
	)
	kb.ResizeKeyboard = true
	return kb
}

func esp32Keyboard() tgbotapi.ReplyKeyboardMarkup {
	kb := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("ESP32 online"),
			tgbotapi.NewKeyboardButton("ESP32 выкл"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("ESP32 красный"),
			tgbotapi.NewKeyboardButton("ESP32 зеленый"),
			tgbotapi.NewKeyboardButton("ESP32 синий"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("ESP32 красный 20%"),
			tgbotapi.NewKeyboardButton("ESP32 красный 50%"),
			tgbotapi.NewKeyboardButton("ESP32 красный 100%"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("ESP32 RGB вручную"),
			tgbotapi.NewKeyboardButton("Назад"),
		),
	)
	kb.ResizeKeyboard = true
	return kb
}

func stripKeyboard() tgbotapi.ReplyKeyboardMarkup {
	kb := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Включить"),
			tgbotapi.NewKeyboardButton("Выключить"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Цвет"),
			tgbotapi.NewKeyboardButton("Анимации"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Яркость"),
			tgbotapi.NewKeyboardButton("Скорость"),
			tgbotapi.NewKeyboardButton("Таймер"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Назад"),
		),
	)
	kb.ResizeKeyboard = true
	return kb
}

func stripColorKeyboard() tgbotapi.ReplyKeyboardMarkup {
	kb := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Красный"),
			tgbotapi.NewKeyboardButton("Зеленый"),
			tgbotapi.NewKeyboardButton("Синий"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Желтый"),
			tgbotapi.NewKeyboardButton("Фиолетовый"),
			tgbotapi.NewKeyboardButton("Белый"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Оранжевый"),
			tgbotapi.NewKeyboardButton("RGB"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Назад"),
		),
	)
	kb.ResizeKeyboard = true
	return kb
}

func stripBrightnessKeyboard() tgbotapi.ReplyKeyboardMarkup {
	kb := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("10%"),
			tgbotapi.NewKeyboardButton("20%"),
			tgbotapi.NewKeyboardButton("30%"),
			tgbotapi.NewKeyboardButton("40%"),
			tgbotapi.NewKeyboardButton("50%"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("60%"),
			tgbotapi.NewKeyboardButton("70%"),
			tgbotapi.NewKeyboardButton("80%"),
			tgbotapi.NewKeyboardButton("90%"),
			tgbotapi.NewKeyboardButton("100%"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Назад"),
		),
	)
	kb.ResizeKeyboard = true
	return kb
}

func stripSpeedKeyboard() tgbotapi.ReplyKeyboardMarkup {
	kb := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Очень медленно"),
			tgbotapi.NewKeyboardButton("Медленно"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Средне"),
			tgbotapi.NewKeyboardButton("Быстро"),
			tgbotapi.NewKeyboardButton("Очень быстро"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Назад"),
		),
	)
	kb.ResizeKeyboard = true
	return kb
}

func stripTimerKeyboard() tgbotapi.ReplyKeyboardMarkup {
	kb := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Через 10 минут"),
			tgbotapi.NewKeyboardButton("Через 30 минут"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Через 1 час"),
			tgbotapi.NewKeyboardButton("Через 2 часа"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Выключить таймер"),
			tgbotapi.NewKeyboardButton("Назад"),
		),
	)
	kb.ResizeKeyboard = true
	return kb
}

func stripModeKeyboard() tgbotapi.ReplyKeyboardMarkup {
	kb := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Режим 1"),
			tgbotapi.NewKeyboardButton("Режим 2"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Режим 3"),
			tgbotapi.NewKeyboardButton("Режим 4"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Назад"),
		),
	)
	kb.ResizeKeyboard = true
	return kb
}

func sendWithKeyboard(bot *tgbotapi.BotAPI, chatID int64, text string, markup tgbotapi.ReplyKeyboardMarkup) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = markup
	_, _ = bot.Send(msg)
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

func handleMessage(bot *tgbotapi.BotAPI, cfg Config, store *NotesStore, lamp *LampService, strip *StripClient, esp32 *ESP32Client, room *RoomSensor, statusLight *StatusLightService, state *userState, msg *tgbotapi.Message) {
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
		case actionStripRGB:
			r, g, b, err := parseRGB(text)
			state.set(userID, actionNone)
			if err != nil {
				sendWithKeyboard(bot, chatID, "Некорректный RGB. Используй формат: 255 128 0 или #FF8000.", stripColorKeyboard())
				return
			}
			if err := strip.SetRGB(r, g, b); err != nil {
				log.Printf("strip rgb failed: %v", err)
				sendWithKeyboard(bot, chatID, "Не удалось подключиться к LED-ленте.\n\nПопробуйте еще раз.", stripColorKeyboard())
				return
			}
			sendWithKeyboard(bot, chatID, "Цвет изменен.", stripKeyboard())
			return
		case actionESP32RGB:
			r, g, b, err := parseRGB(text)
			state.set(userID, actionNone)
			if err != nil {
				sendWithKeyboard(bot, chatID, "Некорректный RGB. Используй формат: 255 128 0 или #FF8000.", esp32Keyboard())
				return
			}
			if err := esp32.SetRGB(r, g, b); err != nil {
				log.Printf("esp32 rgb failed: %v", err)
				sendWithKeyboard(bot, chatID, "Не удалось подключиться к ESP32.\n\nПопробуйте еще раз.", esp32Keyboard())
				return
			}
			sendWithKeyboard(bot, chatID, "ESP32: цвет изменен.", esp32Keyboard())
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
	case "Комната":
		sendWithKeyboard(bot, chatID, "Комната", roomKeyboard())
	case "Температура":
		handleRoomReading(bot, chatID, room, "Температура", func(r RoomReading) string {
			return fmt.Sprintf("Температура: %.2f °C", r.TemperatureC)
		})
	case "Влажность":
		handleRoomReading(bot, chatID, room, "Влажность", func(r RoomReading) string {
			return fmt.Sprintf("Влажность: %.2f%%", r.HumidityPercent)
		})
	case "Давление":
		handleRoomReading(bot, chatID, room, "Давление", func(r RoomReading) string {
			return fmt.Sprintf("Давление: %.2f мм рт. ст. (%.2f hPa)", r.PressureMMHg, r.PressureHPa)
		})
	case "Статус":
		sendWithKeyboard(bot, chatID, "Статус", statusKeyboard())
	case "Занят":
		handleStatusCommand(bot, chatID, statusLight, "занят", RGB{255, 0, 0})
	case "Свободен":
		handleStatusCommand(bot, chatID, statusLight, "свободен", RGB{0, 255, 0})
	case "Разговариваю":
		handleStatusCommand(bot, chatID, statusLight, "разговариваю", RGB{128, 0, 255})
	case "Полузанят":
		handleStatusCommand(bot, chatID, statusLight, "полузанят", RGB{255, 128, 0})
	case "Лампа":
		on, err := lamp.Toggle()
		if err != nil {
			log.Printf("lamp toggle failed: %v", err)
			send(bot, chatID, "Не удалось переключить лампу.")
			return
		}
		if on {
			send(bot, chatID, "Лампа включилась.")
			return
		}
		send(bot, chatID, "Лампа выключилась.")
	case "Лента":
		sendWithKeyboard(bot, chatID, "Лента", stripKeyboard())
	case "ESP32 RGB":
		sendWithKeyboard(bot, chatID, "ESP32 RGB", esp32Keyboard())
	case "Назад":
		send(bot, chatID, "Главное меню.")
	case "Включить":
		handleStripCommand(bot, chatID, stripKeyboard(), "Лента включена.", func() error { return strip.On() })
	case "Выключить":
		handleStripCommand(bot, chatID, stripKeyboard(), "Лента выключена.", func() error { return strip.Off() })
	case "Цвет":
		sendWithKeyboard(bot, chatID, "Цвет", stripColorKeyboard())
	case "Красный":
		handleStripCommand(bot, chatID, stripColorKeyboard(), "Цвет изменен.", func() error { return strip.SetColor(StripRed) })
	case "Зеленый":
		handleStripCommand(bot, chatID, stripColorKeyboard(), "Цвет изменен.", func() error { return strip.SetColor(StripGreen) })
	case "Синий":
		handleStripCommand(bot, chatID, stripColorKeyboard(), "Цвет изменен.", func() error { return strip.SetColor(StripBlue) })
	case "Желтый":
		handleStripCommand(bot, chatID, stripColorKeyboard(), "Цвет изменен.", func() error { return strip.SetColor(StripYellow) })
	case "Фиолетовый":
		handleStripCommand(bot, chatID, stripColorKeyboard(), "Цвет изменен.", func() error { return strip.SetColor(StripPurple) })
	case "Белый":
		handleStripCommand(bot, chatID, stripColorKeyboard(), "Цвет изменен.", func() error { return strip.SetColor(StripWhite) })
	case "Оранжевый":
		handleStripCommand(bot, chatID, stripColorKeyboard(), "Цвет изменен.", func() error { return strip.SetColor(StripOrange) })
	case "RGB":
		state.set(userID, actionStripRGB)
		sendWithoutKeyboard(bot, chatID, "Отправь RGB: 255 128 0 или #FF8000.")
	case "Яркость":
		sendWithKeyboard(bot, chatID, "Яркость", stripBrightnessKeyboard())
	case "10%", "20%", "30%", "40%", "50%", "60%", "70%", "80%", "90%", "100%":
		percent, _ := strconv.Atoi(strings.TrimSuffix(text, "%"))
		handleStripCommand(bot, chatID, stripBrightnessKeyboard(), "Яркость изменена.", func() error { return strip.SetBrightness(percent) })
	case "Скорость":
		sendWithKeyboard(bot, chatID, "Скорость", stripSpeedKeyboard())
	case "Очень медленно":
		handleStripCommand(bot, chatID, stripSpeedKeyboard(), "Скорость изменена.", func() error { return strip.SetSpeed(10) })
	case "Медленно":
		handleStripCommand(bot, chatID, stripSpeedKeyboard(), "Скорость изменена.", func() error { return strip.SetSpeed(25) })
	case "Средне":
		handleStripCommand(bot, chatID, stripSpeedKeyboard(), "Скорость изменена.", func() error { return strip.SetSpeed(50) })
	case "Быстро":
		handleStripCommand(bot, chatID, stripSpeedKeyboard(), "Скорость изменена.", func() error { return strip.SetSpeed(75) })
	case "Очень быстро":
		handleStripCommand(bot, chatID, stripSpeedKeyboard(), "Скорость изменена.", func() error { return strip.SetSpeed(100) })
	case "Анимации":
		sendWithKeyboard(bot, chatID, "Анимации", stripModeKeyboard())
	case "Режим 1":
		handleStripCommand(bot, chatID, stripModeKeyboard(), "Анимация изменена.", func() error { return strip.SetMode(1) })
	case "Режим 2":
		handleStripCommand(bot, chatID, stripModeKeyboard(), "Анимация изменена.", func() error { return strip.SetMode(2) })
	case "Режим 3":
		handleStripCommand(bot, chatID, stripModeKeyboard(), "Анимация изменена.", func() error { return strip.SetMode(3) })
	case "Режим 4":
		handleStripCommand(bot, chatID, stripModeKeyboard(), "Анимация изменена.", func() error { return strip.SetMode(4) })
	case "Таймер":
		sendWithKeyboard(bot, chatID, "Таймер", stripTimerKeyboard())
	case "Через 10 минут", "Через 30 минут", "Через 1 час", "Через 2 часа":
		duration, _ := stripTimerDuration(text)
		handleStripCommand(bot, chatID, stripTimerKeyboard(), "Таймер установлен.", func() error { return strip.SetTimer(duration) })
	case "Выключить таймер":
		handleStripCommand(bot, chatID, stripTimerKeyboard(), "Таймер выключен.", func() error { return strip.CancelTimer() })
	case "ESP32 online":
		handleESP32Command(bot, chatID, esp32Keyboard(), "ESP32 online.", func() error { return esp32.Ping() })
	case "ESP32 выкл":
		handleESP32Command(bot, chatID, esp32Keyboard(), "ESP32 выключен.", func() error { return esp32.Off() })
	case "ESP32 красный":
		handleESP32Command(bot, chatID, esp32Keyboard(), "ESP32: красный.", func() error { return esp32.Red() })
	case "ESP32 зеленый":
		handleESP32Command(bot, chatID, esp32Keyboard(), "ESP32: зеленый.", func() error { return esp32.Green() })
	case "ESP32 синий":
		handleESP32Command(bot, chatID, esp32Keyboard(), "ESP32: синий.", func() error { return esp32.Blue() })
	case "ESP32 красный 20%":
		handleESP32Command(bot, chatID, esp32Keyboard(), "ESP32: красный 20%.", func() error { return esp32.RedBrightness(50) })
	case "ESP32 красный 50%":
		handleESP32Command(bot, chatID, esp32Keyboard(), "ESP32: красный 50%.", func() error { return esp32.RedBrightness(128) })
	case "ESP32 красный 100%":
		handleESP32Command(bot, chatID, esp32Keyboard(), "ESP32: красный 100%.", func() error { return esp32.RedBrightness(255) })
	case "ESP32 RGB вручную":
		state.set(userID, actionESP32RGB)
		sendWithoutKeyboard(bot, chatID, "Отправь RGB для ESP32: 255 128 0 или #FF8000.")
	default:
		send(bot, chatID, "Используй кнопки ниже.")
	}
}

func handleStripCommand(bot *tgbotapi.BotAPI, chatID int64, markup tgbotapi.ReplyKeyboardMarkup, success string, fn func() error) {
	if err := fn(); err != nil {
		log.Printf("strip command failed: %v", err)
		sendWithKeyboard(bot, chatID, "Не удалось подключиться к LED-ленте.\n\nПопробуйте еще раз.", markup)
		return
	}
	sendWithKeyboard(bot, chatID, success, markup)
}

func handleRoomReading(bot *tgbotapi.BotAPI, chatID int64, room *RoomSensor, name string, format func(RoomReading) string) {
	reading, err := room.Read()
	if err != nil {
		log.Printf("room %s failed: %v", name, err)
		sendWithKeyboard(bot, chatID, "Не удалось прочитать датчики комнаты.", roomKeyboard())
		return
	}
	sendWithKeyboard(bot, chatID, format(reading), roomKeyboard())
}

func handleStatusCommand(bot *tgbotapi.BotAPI, chatID int64, statusLight *StatusLightService, name string, rgb RGB) {
	brightness, err := statusLight.Set(context.Background(), rgb)
	if err != nil {
		log.Printf("status light failed: %v", err)
		sendWithKeyboard(bot, chatID, "Не удалось установить статус.", statusKeyboard())
		return
	}
	sendWithKeyboard(bot, chatID, fmt.Sprintf("Статус: %s. Яркость: %d/255.", name, brightness), statusKeyboard())
}

func handleESP32Command(bot *tgbotapi.BotAPI, chatID int64, markup tgbotapi.ReplyKeyboardMarkup, success string, fn func() error) {
	if err := fn(); err != nil {
		log.Printf("esp32 command failed: %v", err)
		sendWithKeyboard(bot, chatID, "Не удалось подключиться к ESP32.\n\nПопробуйте еще раз.", markup)
		return
	}
	sendWithKeyboard(bot, chatID, success, markup)
}

func processUpdates(bot *tgbotapi.BotAPI, cfg Config, store *NotesStore, lamp *LampService, strip *StripClient, esp32 *ESP32Client, room *RoomSensor, statusLight *StatusLightService, state *userState, updates tgbotapi.UpdatesChannel) {
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
		handleMessage(bot, cfg, store, lamp, strip, esp32, room, statusLight, state, update.Message)
	}
	log.Println("updates channel closed")
}
