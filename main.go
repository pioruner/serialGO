package main

import (
	"time"

	"go.bug.st/serial"
)

func openPort(name string) (serial.Port, error) {
	mode := &serial.Mode{
		BaudRate: 9600,
		Parity:   serial.NoParity,
		DataBits: 8,
		StopBits: serial.OneStopBit,
		InitialStatusBits: &serial.ModemOutputBits{
			DTR: true,
			RTS: false,
		},
	}
	port, err := serial.Open(name, mode)
	if err != nil {
		return nil, err
	}
	*mode.InitialStatusBits = serial.ModemOutputBits{
		DTR: false,
		RTS: false,
	}
	if err := port.SetMode(mode); err != nil {
		port.Close()
		return nil, err
	}
	purge(&port)
	return port, nil
}

func purge(port *serial.Port) {
	serial.Port.ResetInputBuffer(*port)
	serial.Port.ResetOutputBuffer(*port)
}

func main() {
	port, _ := openPort("COM14")
	defer port.Close()
	buf := make([]byte, 100)
	time.Sleep(time.Second)
	n, _ := port.Read(buf)
	println(n)
	port.Close()
	port, _ = openPort("COM14")
	time.Sleep(time.Second)
	n, _ = port.Read(buf)
	println(n)
	//port.SetDTR(true)
}
