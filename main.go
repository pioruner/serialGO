package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"unsafe"

	"go.bug.st/serial"
)

const packetSize = 25

func main() {
	portName := "COM14" // или "/dev/ttyUSB0"
	baud := 9600

	mode := &serial.Mode{
		BaudRate: baud,
		Parity:   serial.NoParity,
		DataBits: 8,
		StopBits: serial.OneStopBit,
		InitialStatusBits: &serial.ModemOutputBits{
			DTR: true,  // открываем сразу с поднятым DTR
			RTS: false, // RTS не нужен, если прибор не требует
		},
	}

	port, err := serial.Open(portName, mode)
	if err != nil {
		log.Fatalf("failed to open port: %v", err)
	}
	defer port.Close()

	// Шаг 1: опускаем DTR
	if err := port.SetDTR(false); err != nil {
		log.Fatalf("failed to clear DTR: %v", err)
	}
	time.Sleep(50 * time.Millisecond)

	// Шаг 2: читаем 1 байт (ожидаемый F8)
	buf := make([]byte, 1)
	_, _ = port.Read(buf) // можно не проверять, если знаем что там F8
	log.Printf("Init byte: 0x%X", buf[0])

	// Шаг 3: снова поднимаем DTR
	if err := port.SetDTR(true); err != nil {
		log.Fatalf("failed to set DTR: %v", err)
	}
	log.Println("Device ready, starting to read packets...")

	// Устанавливаем обработчик Ctrl+C
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Чтение в отдельной горутине
	done := make(chan struct{})
	go func() {
		packet := make([]byte, packetSize)
		for {
			n, err := port.Read(packet)
			if err != nil {
				log.Printf("read error: %v", err)
				continue
			}
			if n < packetSize {
				// Если пришло меньше — читаем остаток
				remain := packetSize - n
				rest := make([]byte, remain)
				_, err := port.Read(rest)
				if err != nil {
					log.Printf("read remainder error: %v", err)
					continue
				}
				copy(packet[n:], rest)
			}

			// 3-й, 4-й, 5-й, 6-й байт — float32 (индексация с нуля)
			value := binary.LittleEndian.Uint32(packet[2:6])
			f := float32frombits(value)

			fmt.Printf("Packet: % X | Value: %f\n", packet, f)
		}
	}()

	<-stop
	close(done)
	log.Println("Stopping reader, exiting...")
}

func float32frombits(b uint32) float32 {
	return *(*float32)(unsafe.Pointer(&b))
}
