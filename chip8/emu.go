package chip8

import (
	"github.com/veandco/go-sdl2/sdl"
	"unsafe"
)

type Emulator struct {
	chip8    *Chip8
	window   *sdl.Window
	renderer *sdl.Renderer
	texture  *sdl.Texture
}

func NewEmulator() (*Emulator, error) {
	if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
		return nil, err
	}

	window, err := sdl.CreateWindow(
		"CHIP-8",
		sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED,
		800, 600,
		sdl.WINDOW_SHOWN,
	)
	if err != nil {
		return nil, err
	}

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		return nil, err
	}

	texture, err := renderer.CreateTexture(
		sdl.PIXELFORMAT_RGBA8888,
		sdl.TEXTUREACCESS_STREAMING,
		screenWidth, screenHeight,
	)
	if err != nil {
		return nil, err
	}

	return &Emulator{
		chip8:    New(),
		window:   window,
		renderer: renderer,
		texture:  texture,
	}, nil
}

func (e *Emulator) Shutdown() {
	e.texture.Destroy()
	e.renderer.Destroy()
	e.window.Destroy()
	sdl.Quit()
}

func (e *Emulator) LoadROM(name string) {
	e.chip8.LoadROM(name)

}

func (e *Emulator) Cycle() {
	e.chip8.Cycle()

}

func (e *Emulator) ProcessEvents() bool {
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch et := event.(type) {
		case *sdl.QuitEvent:
			return true
		case *sdl.KeyboardEvent:
			switch et.Type {
			case sdl.KEYDOWN:
				switch et.Keysym.Scancode {
				case sdl.SCANCODE_ESCAPE:
					return true
				case sdl.SCANCODE_X:
					e.chip8.keypad[0x0] = 1
				case sdl.SCANCODE_1:
					e.chip8.keypad[0x1] = 1
				case sdl.SCANCODE_2:
					e.chip8.keypad[0x2] = 1
				case sdl.SCANCODE_3:
					e.chip8.keypad[0x3] = 1
				case sdl.SCANCODE_Q:
					e.chip8.keypad[0x4] = 1
				case sdl.SCANCODE_W:
					e.chip8.keypad[0x5] = 1
				case sdl.SCANCODE_E:
					e.chip8.keypad[0x6] = 1
				case sdl.SCANCODE_A:
					e.chip8.keypad[0x7] = 1
				case sdl.SCANCODE_S:
					e.chip8.keypad[0x8] = 1
				case sdl.SCANCODE_D:
					e.chip8.keypad[0x9] = 1
				case sdl.SCANCODE_Z:
					e.chip8.keypad[0xA] = 1
				case sdl.SCANCODE_C:
					e.chip8.keypad[0xB] = 1
				case sdl.SCANCODE_4:
					e.chip8.keypad[0xC] = 1
				case sdl.SCANCODE_R:
					e.chip8.keypad[0xD] = 1
				case sdl.SCANCODE_F:
					e.chip8.keypad[0xE] = 1
				case sdl.SCANCODE_V:
					e.chip8.keypad[0xF] = 1
				}
			case sdl.KEYUP:
				switch et.Keysym.Scancode {
				case sdl.SCANCODE_X:
					e.chip8.keypad[0x0] = 0
				case sdl.SCANCODE_1:
					e.chip8.keypad[0x1] = 0
				case sdl.SCANCODE_2:
					e.chip8.keypad[0x2] = 0
				case sdl.SCANCODE_3:
					e.chip8.keypad[0x3] = 0
				case sdl.SCANCODE_Q:
					e.chip8.keypad[0x4] = 0
				case sdl.SCANCODE_W:
					e.chip8.keypad[0x5] = 0
				case sdl.SCANCODE_E:
					e.chip8.keypad[0x6] = 0
				case sdl.SCANCODE_A:
					e.chip8.keypad[0x7] = 0
				case sdl.SCANCODE_S:
					e.chip8.keypad[0x8] = 0
				case sdl.SCANCODE_D:
					e.chip8.keypad[0x9] = 0
				case sdl.SCANCODE_Z:
					e.chip8.keypad[0xA] = 0
				case sdl.SCANCODE_C:
					e.chip8.keypad[0xB] = 0
				case sdl.SCANCODE_4:
					e.chip8.keypad[0xC] = 0
				case sdl.SCANCODE_R:
					e.chip8.keypad[0xD] = 0
				case sdl.SCANCODE_F:
					e.chip8.keypad[0xE] = 0
				case sdl.SCANCODE_V:
					e.chip8.keypad[0xF] = 0
				}
			}
		}
	}
	return false
}

func (e *Emulator) UpdateScreen() {
	/*
		screen := [64 * 32]uint32{0}
		for i := range screen {
			//screen[i] = 0x00777777 + uint32(rand.Intn(0x00AAAAAA))
			screen[i] = 0x000000FF
		}
		e.texture.Update(nil, unsafe.Pointer(&screen), 4*64)
	*/
	e.texture.Update(nil, unsafe.Pointer(&e.chip8.screen), 4*screenWidth)
	//e.window.UpdateSurface()
	e.renderer.Clear()
	e.renderer.Copy(e.texture, nil, nil)
	e.renderer.Present()
}
