package main

import (
	"fmt"
	"io"
	"os"
	"syscall"
	"time"
)

const (
	i2cSlaveAddress = 0x0703
	aht20Address    = 0x38
	bmp280Address   = 0x77
)

type RoomReading struct {
	TemperatureC    float64
	HumidityPercent float64
	PressureHPa     float64
	PressureMMHg    float64
}

type RoomSensor struct {
	device string
}

func newRoomSensor(device string) *RoomSensor {
	return &RoomSensor{device: device}
}

func (s *RoomSensor) Read() (RoomReading, error) {
	f, err := os.OpenFile(s.device, os.O_RDWR, 0)
	if err != nil {
		return RoomReading{}, fmt.Errorf("open i2c device %s: %w", s.device, err)
	}
	defer f.Close()

	temperature, humidity, err := readAHT20(f)
	if err != nil {
		return RoomReading{}, err
	}
	pressureHPa, err := readBMP280(f)
	if err != nil {
		return RoomReading{}, err
	}

	return RoomReading{
		TemperatureC:    temperature,
		HumidityPercent: humidity,
		PressureHPa:     pressureHPa,
		PressureMMHg:    pressureHPa * 0.750061683,
	}, nil
}

func readAHT20(f *os.File) (float64, float64, error) {
	if err := setI2CSlave(f, aht20Address); err != nil {
		return 0, 0, fmt.Errorf("select aht20: %w", err)
	}
	_, _ = f.Write([]byte{0xBE, 0x08, 0x00})
	time.Sleep(10 * time.Millisecond)

	if _, err := f.Write([]byte{0xAC, 0x33, 0x00}); err != nil {
		return 0, 0, fmt.Errorf("start aht20 measurement: %w", err)
	}
	time.Sleep(80 * time.Millisecond)

	buf := make([]byte, 6)
	if _, err := io.ReadFull(f, buf); err != nil {
		return 0, 0, fmt.Errorf("read aht20 measurement: %w", err)
	}
	rawHumidity := (uint32(buf[1])<<16 | uint32(buf[2])<<8 | uint32(buf[3])) >> 4
	rawTemperature := (uint32(buf[3]&0x0f) << 16) | uint32(buf[4])<<8 | uint32(buf[5])

	humidity := float64(rawHumidity) * 100 / 1048576
	temperature := float64(rawTemperature)*200/1048576 - 50
	return temperature, humidity, nil
}

func readBMP280(f *os.File) (float64, error) {
	if err := setI2CSlave(f, bmp280Address); err != nil {
		return 0, fmt.Errorf("select bmp280: %w", err)
	}
	if err := writeReg(f, 0xF4, 0x27); err != nil {
		return 0, fmt.Errorf("configure bmp280 measurement: %w", err)
	}
	if err := writeReg(f, 0xF5, 0xA0); err != nil {
		return 0, fmt.Errorf("configure bmp280 filter: %w", err)
	}
	time.Sleep(50 * time.Millisecond)

	calib, err := readReg(f, 0x88, 24)
	if err != nil {
		return 0, fmt.Errorf("read bmp280 calibration: %w", err)
	}
	data, err := readReg(f, 0xF7, 6)
	if err != nil {
		return 0, fmt.Errorf("read bmp280 measurement: %w", err)
	}

	adcP := int32(data[0])<<12 | int32(data[1])<<4 | int32(data[2])>>4
	adcT := int32(data[3])<<12 | int32(data[4])<<4 | int32(data[5])>>4
	pressurePa, err := compensateBMP280Pressure(adcT, adcP, bmp280Calibration{
		digT1: u16(calib, 0),
		digT2: i16(calib, 2),
		digT3: i16(calib, 4),
		digP1: u16(calib, 6),
		digP2: i16(calib, 8),
		digP3: i16(calib, 10),
		digP4: i16(calib, 12),
		digP5: i16(calib, 14),
		digP6: i16(calib, 16),
		digP7: i16(calib, 18),
		digP8: i16(calib, 20),
		digP9: i16(calib, 22),
	})
	if err != nil {
		return 0, err
	}
	return pressurePa / 100, nil
}

func setI2CSlave(f *os.File, address int) error {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), uintptr(i2cSlaveAddress), uintptr(address))
	if errno != 0 {
		return errno
	}
	return nil
}

func writeReg(f *os.File, reg, value byte) error {
	_, err := f.Write([]byte{reg, value})
	return err
}

func readReg(f *os.File, reg byte, size int) ([]byte, error) {
	if _, err := f.Write([]byte{reg}); err != nil {
		return nil, err
	}
	buf := make([]byte, size)
	_, err := io.ReadFull(f, buf)
	return buf, err
}

type bmp280Calibration struct {
	digT1 uint16
	digT2 int16
	digT3 int16
	digP1 uint16
	digP2 int16
	digP3 int16
	digP4 int16
	digP5 int16
	digP6 int16
	digP7 int16
	digP8 int16
	digP9 int16
}

func compensateBMP280Pressure(adcT, adcP int32, c bmp280Calibration) (float64, error) {
	var1 := (((adcT >> 3) - (int32(c.digT1) << 1)) * int32(c.digT2)) >> 11
	var2 := (((((adcT >> 4) - int32(c.digT1)) * ((adcT >> 4) - int32(c.digT1))) >> 12) * int32(c.digT3)) >> 14
	tFine := var1 + var2

	pVar1 := int64(tFine) - 128000
	pVar2 := pVar1 * pVar1 * int64(c.digP6)
	pVar2 += (pVar1 * int64(c.digP5)) << 17
	pVar2 += int64(c.digP4) << 35
	pVar1 = ((pVar1*pVar1*int64(c.digP3))>>8 + (pVar1*int64(c.digP2))<<12)
	pVar1 = (((int64(1) << 47) + pVar1) * int64(c.digP1)) >> 33
	if pVar1 == 0 {
		return 0, fmt.Errorf("invalid bmp280 calibration")
	}

	pressure := int64(1048576) - int64(adcP)
	pressure = (((pressure << 31) - pVar2) * 3125) / pVar1
	pVar1 = (int64(c.digP9) * (pressure >> 13) * (pressure >> 13)) >> 25
	pVar2 = (int64(c.digP8) * pressure) >> 19
	pressure = ((pressure + pVar1 + pVar2) >> 8) + (int64(c.digP7) << 4)
	return float64(pressure) / 256, nil
}

func u16(b []byte, offset int) uint16 {
	return uint16(b[offset]) | uint16(b[offset+1])<<8
}

func i16(b []byte, offset int) int16 {
	return int16(u16(b, offset))
}
