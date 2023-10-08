package main

import (
	"fmt"
	"os"
	"time"

	"github.com/ffd2subroutine/chip8/chip8"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Provide a ROM file!")
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

	exit := false
	for !exit {
		exit = e.ProcessEvents()
		e.Cycle()
		e.UpdateScreen()
		time.Sleep(14 * time.Millisecond)
	}
}
