package main

import (
	"fmt"
	"os"
	"time"

	"github.com/ffd2subroutine/chip8/chip8"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Provide a ROM")
		os.Exit(1)
	}
	rom := os.Args[1]
	e, err := chip8.NewEmulator()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer e.Shutdown()

	e.LoadROM(rom)
	last := time.Now()
	exit := false
	for !exit {
		exit = e.ProcessEvents()

		now := time.Now()
		dt := now.Sub(last)
		if dt > 1 {
			e.Cycle()
			e.UpdateScreen()
		}
	}
}
