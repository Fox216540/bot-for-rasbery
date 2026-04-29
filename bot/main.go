package main

import (
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}
	setProcRoot(cfg.HostProc)
	setSysRoot(cfg.HostSys)

	store, err := newNotesStore(cfg.NotesFile)
	if err != nil {
		log.Fatalf("notes init error: %v", err)
	}

	if err := waitForAI(cfg.AIURL, 2*time.Minute); err != nil {
		log.Fatalf("ai service unavailable: %v", err)
	}
	if err := reindexToAI(store, cfg.AIURL); err != nil {
		log.Fatalf("ai reindex failed: %v", err)
	}

	bot, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		log.Fatalf("telegram init error: %v", err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	state := newUserState()
	processUpdates(bot, cfg, store, state, updates)
}
