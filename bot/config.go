package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	BotToken      string
	AllowedUserID int64
	AIURL         string
	NotesFile     string
	RebootCommand string
	HostProc      string
	HostSys       string
	LampPort      string
	LampLocation  string
	LampStateFile string
}

func loadConfig() (Config, error) {
	_ = godotenv.Load()
	cfg := Config{
		BotToken:      strings.TrimSpace(os.Getenv("BOT_TOKEN")),
		AIURL:         strings.TrimRight(strings.TrimSpace(os.Getenv("AI_URL")), "/"),
		NotesFile:     strings.TrimSpace(os.Getenv("NOTES_FILE")),
		RebootCommand: strings.TrimSpace(os.Getenv("REBOOT_COMMAND")),
		HostProc:      strings.TrimSpace(os.Getenv("HOST_PROC")),
		HostSys:       strings.TrimSpace(os.Getenv("HOST_SYS")),
		LampPort:      strings.TrimSpace(os.Getenv("LAMP_PORT")),
		LampLocation:  strings.TrimSpace(os.Getenv("LAMP_HUB_LOCATION")),
		LampStateFile: strings.TrimSpace(os.Getenv("LAMP_STATE_FILE")),
	}
	if cfg.BotToken == "" {
		return Config{}, fmt.Errorf("BOT_TOKEN is required")
	}
	id, err := strconv.ParseInt(strings.TrimSpace(os.Getenv("ALLOWED_USER_ID")), 10, 64)
	if err != nil {
		return Config{}, fmt.Errorf("invalid ALLOWED_USER_ID: %w", err)
	}
	cfg.AllowedUserID = id
	if cfg.AIURL == "" {
		cfg.AIURL = "http://localhost:8000"
	}
	if cfg.NotesFile == "" {
		cfg.NotesFile = "./notes.json"
	}
	if cfg.RebootCommand == "" {
		cfg.RebootCommand = "sudo reboot"
	}
	if cfg.HostProc == "" {
		cfg.HostProc = "/proc"
	}
	if cfg.HostSys == "" {
		cfg.HostSys = "/sys"
	}
	if cfg.LampPort == "" {
		cfg.LampPort = "4"
	}
	if cfg.LampLocation == "" {
		cfg.LampLocation = "1-1"
	}
	if cfg.LampStateFile == "" {
		cfg.LampStateFile = "./lamp_state.json"
	}
	return cfg, nil
}
