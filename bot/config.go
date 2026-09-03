package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	BotToken              string
	AllowedUserID         int64
	AIURL                 string
	NotesFile             string
	RebootCommand         string
	HostProc              string
	HostSys               string
	LampPort              string
	LampLocation          string
	LampStateFile         string
	StripURL              string
	ESP32URL              string
	RoomI2CDevice         string
	MeteoURL              string
	MeteoLatitude         float64
	MeteoLongitude        float64
	StatusDayBrightness   int
	StatusNightBrightness int
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
		StripURL:      strings.TrimRight(strings.TrimSpace(os.Getenv("STRIP_SERVICE_URL")), "/"),
		ESP32URL:      strings.TrimRight(strings.TrimSpace(os.Getenv("ESP32_URL")), "/"),
		RoomI2CDevice: strings.TrimSpace(os.Getenv("ROOM_I2C_DEVICE")),
		MeteoURL:      strings.TrimRight(strings.TrimSpace(os.Getenv("METEO_URL")), "/"),
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
	if cfg.StripURL == "" {
		cfg.StripURL = "http://127.0.0.1:8010"
	}
	if cfg.ESP32URL == "" {
		cfg.ESP32URL = "http://192.168.1.211"
	}
	if cfg.RoomI2CDevice == "" {
		cfg.RoomI2CDevice = "/dev/i2c-1"
	}
	if cfg.MeteoURL == "" {
		cfg.MeteoURL = "https://api.open-meteo.com/v1/forecast"
	}
	cfg.MeteoLatitude, err = envFloat("METEO_LATITUDE", 39.4699)
	if err != nil {
		return Config{}, err
	}
	cfg.MeteoLongitude, err = envFloat("METEO_LONGITUDE", -0.3763)
	if err != nil {
		return Config{}, err
	}
	cfg.StatusDayBrightness, err = envInt("STATUS_DAY_BRIGHTNESS", 128)
	if err != nil {
		return Config{}, err
	}
	cfg.StatusNightBrightness, err = envInt("STATUS_NIGHT_BRIGHTNESS", 1)
	if err != nil {
		return Config{}, err
	}
	if err := validateBrightness("STATUS_DAY_BRIGHTNESS", cfg.StatusDayBrightness); err != nil {
		return Config{}, err
	}
	if err := validateBrightness("STATUS_NIGHT_BRIGHTNESS", cfg.StatusNightBrightness); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func envFloat(name string, fallback float64) (float64, error) {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", name, err)
	}
	return parsed, nil
}

func envInt(name string, fallback int) (int, error) {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", name, err)
	}
	return parsed, nil
}

func validateBrightness(name string, value int) error {
	if value < 0 || value > 255 {
		return fmt.Errorf("invalid %s: must be 0..255", name)
	}
	return nil
}
