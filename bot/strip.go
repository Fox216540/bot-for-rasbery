package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	lotuslantern "github.com/Rxflex/LotusLantern"
)

const (
	stripDeviceName         = "ELK-BLEDDM 8C"
	defaultStripIdleTimeout = time.Minute
)

type RGB struct {
	R int
	G int
	B int
}

var (
	StripRed    = RGB{255, 0, 0}
	StripGreen  = RGB{0, 255, 0}
	StripBlue   = RGB{0, 0, 255}
	StripYellow = RGB{255, 255, 0}
	StripPurple = RGB{128, 0, 255}
	StripWhite  = RGB{255, 255, 255}
	StripOrange = RGB{255, 128, 0}
)

type stripLamp interface {
	LightOn(on bool) error
	ChangeColorRGB(r, g, b int) error
	ChangeBrightness(brightness, lightMode int) error
	ChangeMode(mode int) error
	ChangeModeSpeed(speed int) error
	Close() error
}

type StripConnector interface {
	Connect(addr, name string) (stripLamp, error)
}

type lotusStripConnector struct{}

func (lotusStripConnector) Connect(addr, name string) (stripLamp, error) {
	return lotuslantern.Connect(addr, name)
}

type StripService struct {
	mac       string
	name      string
	connector StripConnector

	mu      sync.Mutex
	lamp    stripLamp
	idle    *time.Timer
	idleTTL time.Duration
	timerMu sync.Mutex
	timer   *time.Timer
}

func newStripService(mac string) *StripService {
	return &StripService{
		mac:       mac,
		name:      stripDeviceName,
		connector: lotusStripConnector{},
		idleTTL:   defaultStripIdleTimeout,
	}
}

func (s *StripService) On() error {
	if err := s.withLamp("power on", func(l stripLamp) error {
		return l.LightOn(true)
	}); err != nil {
		return fmt.Errorf("power on: %w", err)
	}
	return nil
}

func (s *StripService) Off() error {
	if err := s.withLamp("power off", func(l stripLamp) error {
		return l.LightOn(false)
	}); err != nil {
		return fmt.Errorf("power off: %w", err)
	}
	return nil
}

func (s *StripService) SetRGB(r, g, b int) error {
	return s.SetColor(RGB{r, g, b})
}

func (s *StripService) SetColor(rgb RGB) error {
	if err := validateRGB(rgb.R, rgb.G, rgb.B); err != nil {
		return err
	}
	if err := s.withLamp(fmt.Sprintf("color rgb %d %d %d", rgb.R, rgb.G, rgb.B), func(l stripLamp) error {
		return l.ChangeColorRGB(rgb.R, rgb.G, rgb.B)
	}); err != nil {
		return fmt.Errorf("set color: %w", err)
	}
	return nil
}

func (s *StripService) SetBrightness(percent int) error {
	value, err := brightnessValue(percent)
	if err != nil {
		return err
	}
	if err := s.withLamp(fmt.Sprintf("brightness %d%%", percent), func(l stripLamp) error {
		return l.ChangeBrightness(value, 0)
	}); err != nil {
		return fmt.Errorf("set brightness: %w", err)
	}
	return nil
}

func (s *StripService) SetSpeed(value int) error {
	if value < 0 || value > 255 {
		return fmt.Errorf("invalid strip speed: %d", value)
	}
	if err := s.withLamp(fmt.Sprintf("speed %d", value), func(l stripLamp) error {
		return l.ChangeModeSpeed(value)
	}); err != nil {
		return fmt.Errorf("set speed: %w", err)
	}
	return nil
}

func (s *StripService) SetMode(mode int) error {
	if mode < 1 || mode > 127 {
		return fmt.Errorf("invalid strip mode: %d", mode)
	}
	if err := s.withLamp(fmt.Sprintf("mode %d", mode), func(l stripLamp) error {
		return l.ChangeMode(mode)
	}); err != nil {
		return fmt.Errorf("set mode: %w", err)
	}
	return nil
}

func (s *StripService) SetTimer(d time.Duration) error {
	if d <= 0 {
		return fmt.Errorf("invalid strip timer duration: %s", d)
	}
	s.timerMu.Lock()
	if s.timer != nil {
		s.timer.Stop()
	}
	s.timer = time.AfterFunc(d, func() {
		if err := s.Off(); err != nil {
			log.Printf("strip timer off failed: %v", err)
		}
	})
	s.timerMu.Unlock()
	log.Printf("strip timer set: %s", d)
	return nil
}

func (s *StripService) CancelTimer() error {
	s.timerMu.Lock()
	defer s.timerMu.Unlock()
	if s.timer != nil {
		s.timer.Stop()
		s.timer = nil
	}
	log.Print("strip timer canceled")
	return nil
}

func (s *StripService) Close() error {
	if err := s.CancelTimer(); err != nil {
		return fmt.Errorf("cancel strip timer: %w", err)
	}
	if err := s.Disconnect(); err != nil {
		return fmt.Errorf("disconnect strip: %w", err)
	}
	return nil
}

func (s *StripService) withLamp(command string, run func(stripLamp) error) error {
	start := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stopIdleLocked()

	lamp, err := s.ensureConnected(command)
	if err != nil {
		return err
	}

	log.Printf("strip command start command=%q", command)
	if err := run(lamp); err != nil {
		log.Printf("strip command failed command=%q duration=%s error=%v", command, time.Since(start), err)
		s.disconnectLocked(command)
		lamp, connectErr := s.ensureConnected(command)
		if connectErr != nil {
			return fmt.Errorf("reconnect strip: %w", connectErr)
		}
		if retryErr := run(lamp); retryErr != nil {
			s.disconnectLocked(command)
			return retryErr
		}
	}

	log.Printf("strip command done command=%q duration=%s", command, time.Since(start))
	s.scheduleIdleLocked(command)
	return nil
}

func (s *StripService) ensureConnected(command string) (stripLamp, error) {
	if s.lamp != nil {
		return s.lamp, nil
	}
	log.Printf("strip connect start mac=%s command=%q", s.mac, command)
	lamp, err := s.connector.Connect(s.mac, s.name)
	if err != nil {
		log.Printf("strip connect failed mac=%s command=%q error=%v", s.mac, command, err)
		log.Printf("strip connect retry mac=%s command=%q", s.mac, command)
		lamp, err = s.connector.Connect(s.mac, s.name)
		if err == nil {
			log.Printf("strip connect retry succeeded mac=%s command=%q", s.mac, command)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("connect strip: %w", err)
	}
	s.lamp = lamp
	log.Printf("strip connected mac=%s command=%q", s.mac, command)
	return s.lamp, nil
}

func (s *StripService) Disconnect() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stopIdleLocked()
	return s.disconnectLocked("manual disconnect")
}

func (s *StripService) disconnectLocked(command string) error {
	s.stopIdleLocked()
	if s.lamp == nil {
		return nil
	}
	err := s.lamp.Close()
	s.lamp = nil
	if err != nil {
		log.Printf("strip disconnect failed command=%q error=%v", command, err)
	} else {
		log.Printf("strip disconnected command=%q", command)
	}
	return err
}

func (s *StripService) scheduleIdleLocked(command string) {
	if s.idleTTL <= 0 {
		return
	}
	s.stopIdleLocked()
	s.idle = time.AfterFunc(s.idleTTL, func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		if err := s.disconnectLocked("idle timeout"); err != nil {
			log.Printf("strip idle disconnect failed after command=%q error=%v", command, err)
			return
		}
		log.Printf("strip idle disconnect completed after command=%q timeout=%s", command, s.idleTTL)
	})
	log.Printf("strip idle disconnect scheduled command=%q timeout=%s", command, s.idleTTL)
}

func (s *StripService) stopIdleLocked() {
	if s.idle != nil {
		s.idle.Stop()
		s.idle = nil
	}
}

func parseRGB(text string) (int, int, int, error) {
	text = strings.TrimSpace(text)
	if strings.HasPrefix(text, "#") {
		hex := strings.TrimPrefix(text, "#")
		if len(hex) != 6 {
			return 0, 0, 0, fmt.Errorf("invalid hex rgb")
		}
		v, err := strconv.ParseUint(hex, 16, 32)
		if err != nil {
			return 0, 0, 0, fmt.Errorf("invalid hex rgb: %w", err)
		}
		return int((v >> 16) & 0xff), int((v >> 8) & 0xff), int(v & 0xff), nil
	}

	parts := strings.Fields(text)
	if len(parts) != 3 {
		return 0, 0, 0, fmt.Errorf("rgb must contain three numbers")
	}
	r, err := parseRGBPart(parts[0])
	if err != nil {
		return 0, 0, 0, err
	}
	g, err := parseRGBPart(parts[1])
	if err != nil {
		return 0, 0, 0, err
	}
	b, err := parseRGBPart(parts[2])
	if err != nil {
		return 0, 0, 0, err
	}
	return r, g, b, nil
}

func parseRGBPart(text string) (int, error) {
	v, err := strconv.Atoi(text)
	if err != nil {
		return 0, fmt.Errorf("invalid rgb value %q: %w", text, err)
	}
	if v < 0 || v > 255 {
		return 0, fmt.Errorf("rgb value out of range: %d", v)
	}
	return v, nil
}

func validateRGB(r, g, b int) error {
	for _, v := range []int{r, g, b} {
		if v < 0 || v > 255 {
			return fmt.Errorf("rgb value out of range: %d", v)
		}
	}
	return nil
}

func brightnessValue(percent int) (int, error) {
	if percent < 0 || percent > 100 {
		return 0, fmt.Errorf("invalid strip brightness: %d", percent)
	}
	return percent * 255 / 100, nil
}

func stripTimerDuration(name string) (time.Duration, error) {
	durations := map[string]time.Duration{
		"Через 10 минут": 10 * time.Minute,
		"Через 30 минут": 30 * time.Minute,
		"Через 1 час":    time.Hour,
		"Через 2 часа":   2 * time.Hour,
	}
	duration, ok := durations[name]
	if !ok {
		return 0, fmt.Errorf("unknown strip timer: %s", name)
	}
	return duration, nil
}
