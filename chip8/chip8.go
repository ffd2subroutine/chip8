package chip8

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
)

var ErrUnknownOpcode = errors.New("unknown opcode")

const (
	// Programs are loaded starting at this address.
	programAddress = 0x200 // uint16?
	// 16 built-in characters are loaded starting at this address.
	// Where we store the fonts is actually unspecified, but must be whithin
	// memory reserved for the interpeter.
	fontAddress = 0x50 // uint16?
	fontSize    = 5

	screenWidth  = 64
	screenHeight = 32
)

// The font set of 16 built-in characters(sprites).
// For example, the character 0 is represented as: 0xF0, 0x90, 0x90, 0x90, 0xF0.
// Let's see the binary representations of these values:
// HEX  Binary   Sprite
// 0xF0 11110000 ****____
// 0x90 10010000 *__*____
// 0x90 10010000 *__*____
// 0x90 10010000 *__*____
// 0xF0 11110000 ****____
// Now, if you closely take a look at the first four bits of each these values,
// you will see a zero there. We can see that the width of the sprite that
// represents this character is 8 bits(1 byte) but that applies to any sprite
// not just to characters.
var (
	fontset = [fontSize * 16]uint8{
		0xF0, 0x90, 0x90, 0x90, 0xF0, // 0
		0x20, 0x60, 0x20, 0x20, 0x70, // 1
		0xF0, 0x10, 0xF0, 0x80, 0xF0, // 2
		0xF0, 0x10, 0xF0, 0x10, 0xF0, // 3
		0x90, 0x90, 0xF0, 0x10, 0x10, // 4
		0xF0, 0x80, 0xF0, 0x10, 0xF0, // 5
		0xF0, 0x80, 0xF0, 0x90, 0xF0, // 6
		0xF0, 0x10, 0x20, 0x40, 0x40, // 7
		0xF0, 0x90, 0xF0, 0x90, 0xF0, // 8
		0xF0, 0x90, 0xF0, 0x10, 0xF0, // 9
		0xF0, 0x90, 0xF0, 0x90, 0x90, // A
		0xE0, 0x90, 0xE0, 0x90, 0xE0, // B
		0xF0, 0x80, 0x80, 0x80, 0xF0, // C
		0xE0, 0x90, 0x90, 0x90, 0xE0, // D
		0xF0, 0x80, 0xF0, 0x80, 0xF0, // E
		0xF0, 0x80, 0xF0, 0x80, 0x80, // F
	}
)

type Chip8 struct {
	// 4K of memory:
	// 0x000 - 0x1FF reserved for the CHIP-8 interpreter where 0x050 - 0x0A0
	// will be used as a storage space for 16 built-in characters.
	// 0x200 - 0xFFF availabe, kind of, for the programs that we load from the ROM.
	memory [4096]uint8
	// General purpose data registers.
	v [16]uint8
	// The address register, also called the index register, used to store
	// memory addresses when performing operations involving reading and
	// writing to and from memory.
	i uint16
	// The program counter that holds the address of the next instruction to execute.
	pc uint16
	// The stack that is used to store return addresses when ever the CALL instruction gets executed.
	stack [16]uint16
	// The stack pointer.
	sp uint8

	delayTimer uint8
	soundTimer uint8

	keypad [16]uint8
	screen [screenWidth * screenHeight]uint32
}

func New() *Chip8 {
	c := &Chip8{
		pc: programAddress,
	}
	// Load the fonts into CHIP-8 memory.
	for i, v := range fontset {
		c.memory[fontAddress+i] = v
	}
	return c
}

func (c *Chip8) LoadROM(name string) error {
	// No need for io.ReadAll(f), this is shorter.
	buf, err := os.ReadFile(name)
	if err != nil {
		return err
	}
	// TODO: check the ROM size, to prevent writing over available memory.
	for i, v := range buf {
		c.memory[programAddress+i] = v
	}
	return nil
}

func (c *Chip8) Cycle() {
	// Fetch the opcode(instruction).
	// The opcode in memory is 2 bytes long so we fetch two bytes pointed by
	// the program counter, that means the locations pointed by pc and pc + 1,
	// and we merge them.
	// TODO: explain the process in more detail.
	opcode := uint16(c.memory[c.pc])<<8 | uint16(c.memory[c.pc+1])
	//fmt.Printf("%x\n", opcode)
	//fmt.Printf("PC before increment: %x\n", c.pc)

	// Increment the program counter so that we point to the next opcode.
	c.pc += 2
	//fmt.Printf("%x\n", c.pc)

	// TODO: do we just log unknown opcodes?
	// Decode and execute the opcode
	err := c.decode(opcode)
	if err != nil {
		fmt.Printf("%v\n", err)
	}

	if c.delayTimer > 0 {
		c.delayTimer--
	}
	// TODO
	if c.soundTimer > 0 {
		c.soundTimer--
	}
	/*
		col := 0
		for _, v := range c.screen {
			fmt.Printf("%d", v)
			if col > 62 {
				fmt.Println("")
				col = 0
			} else {
				col++
			}
		}
		fmt.Println("")
	*/
}

func (c *Chip8) decode(opcode uint16) error {
	//fmt.Println("in decode")
	// Get the first nibble(half-byte)
	switch first := opcode >> 12; first {
	case 0x0: // 0N[NN]
		fmt.Println("in 0x0")
		// We are only interested in the third and fourth nibble here.
		switch nn := opcode & 0x00FF; nn {
		case 0xE0: // CLS Clear the screen.
			fmt.Println("in 0x00E")
			c.screen = [screenWidth * screenHeight]uint32{}
			// TODO: set the draw flag so that we can issue an update in our rendering code.
		case 0xEE: // RET Return from a subroutine
			// First decrement the stack pointer.
			c.sp--
			// Then set the program counter to the address previously
			// stored in the stack.
			c.pc = c.stack[c.sp]
		default:
			return ErrUnknownOpcode
		}
	case 0x1: // 1NNN JP Jump to address nnn
		fmt.Println("in 0x1")
		// That means that we have to extract the second, third and fourth
		// nibble from the opcode and store it in the program counter.
		// nnn := opcode & 0x0FFF
		// c.pc = nnn
		c.pc = opcode & 0x0FFF
		//fmt.Printf("PC in 0x1: %x\n", c.pc)
	case 0x2: // 2NNN CALL Execute subroutine at address NNN
		fmt.Println("in 0x2")
		// Store the address in the program counter in to the stack
		// and decrement the stack pointer
		c.stack[c.sp] = c.pc
		c.sp++
		// Extract the address in the opcode and store it in the program counter.
		c.pc = opcode & 0x0FFF
	case 0x3: // 3XNN SE Skip the next instruction if the value of register Vx equals NN.
		fmt.Println("in 0x3")
		// Extract X from the opcode.
		x := opcode & 0x0F00 >> 8
		// Extract last two nibbles.
		nn := opcode & 0x00FF
		if c.v[x] == uint8(nn) {
			c.pc += 2
		}
	case 0x4: // 4XNN SNE Skip the next instruction if the value of register Vx is not equal to NN.
		fmt.Println("in 0x4")
		x := opcode & 0x0F00 >> 8
		nn := opcode & 0x00FF
		if c.v[x] != uint8(nn) {
			c.pc += 2
		}
	case 0x5: // 5XY0 SE Skip the next instruction if the value of register Vx is equal to the value of register Vy.
		fmt.Println("in 0x5")

		// Extract X from the opcode.
		x := opcode & 0x0F00 >> 8
		// Extract Y from the opcode.
		y := opcode & 0x00F0 >> 4
		if c.v[x] == c.v[y] {
			c.pc += 2
		}
	case 0x6: // 6XNN Store number NN in register Vx
		fmt.Println("in 0x6")
		// Yes we could do this in one line: c.v[opcode&0x0F00>>8] = uint8(opcode&0x00FF)
		x := opcode & 0x0F00 >> 8
		nn := opcode & 0x00FF
		c.v[x] = uint8(nn)
	case 0x7: // 7XNN Add the value nn to register vx, in other words,
		fmt.Println("in 0x7")
		// increment the existing value of Vx by NN, Vx += NN.
		x := opcode & 0x0F00 >> 8
		nn := opcode & 0x00FF
		c.v[x] += uint8(nn)
	case 0x8: // 8XY[N] Based on what we have for the fourth nibble N we will do appropriate operation.
		fmt.Println("in 0x8")
		x := opcode & 0x0F00 >> 8
		y := opcode & 0x00F0 >> 4
		// Get the fourth nibble and act accordingly
		switch n := opcode & 0x000F; n {
		case 0x0: // 8XY0 Store the value of register Vy in register Vx.
			c.v[x] = c.v[y]
		case 0x1: // 8XY1 Set Vx to Vx OR Vy.
			c.v[x] |= c.v[y]
		case 0x2: // 8XY2 Set Vx to Vx AND Vy.
			c.v[x] &= c.v[y]
		case 0x3: // 8XY3 Set Vx to Vx XOR Vy.
			c.v[x] ^= c.v[y]
		case 0x4: // 8XY4 Add the value of register Vy to Vx. Set VF to 1 if carry occurs, 0 if not.
			res := uint16(x) + uint16(y)
			if res > 0xFF {
				c.v[0xF] = 1
			} else {
				c.v[0xF] = 0
			}
			c.v[x] = uint8(res)
		case 0x5: // 8XY5 Subtract the value of register Vy from Vx. Set VF to 0 if borrow occurs, 1 if not.
			if c.v[x] > c.v[y] {
				c.v[0xF] = 1
			} else {
				c.v[0xF] = 0
			}
			c.v[x] -= c.v[y]
		case 0x6: // 8XY6 Store the value of register Vy(we are actually storing Vx instead) shifted right one bit in Vx
			// and set register VF to the least significant bit prior to the shift.
			c.v[0xF] = c.v[x] & 0x1 // 0b00000001
			// NOTE: https://github.com/mattmikolay/chip-8/wiki/CHIP%E2%80%908-Instruction-Set#notes
			//c.v[x] = c.v[y] >> 1
			c.v[x] >>= 1
		case 0x7: // 8XY7 Set register Vx to the value of Vy minux Vx. Set VF to 0 if borrow occurs, 1 if not.
			if c.v[y] > c.v[x] {
				c.v[0xF] = 1
			} else {
				c.v[0xF] = 0
			}
			c.v[x] = c.v[y] - c.v[x]
		case 0xE: // 8XYE Store the value of Vy(for now we will store it in Vx instead) shifted left one bit in Vx
			// and set register VF to the most significant bit prior to the shift.
			c.v[0xF] = c.v[x] & 0x80 >> 7 // c.v[x] & 0b10000000 >> 7 or we could shift first (c.v[x] >> 7) & 0x1 // 0b00000001
			// NOTE: https://github.com/mattmikolay/chip-8/wiki/CHIP%E2%80%908-Instruction-Set#notes
			//c.v[x] = c.v[y] << 1
			c.v[x] <<= 1
		default:
			return ErrUnknownOpcode
		}
	case 0x9: // 9XY0 Skip the next instruction if the value of Vx is not equal to Vy.
		fmt.Println("in 0x9")
		x := opcode & 0x0F00 >> 8
		y := opcode & 0x00F0 >> 4
		if c.v[x] != c.v[y] {
			c.pc += 2
		}
	case 0xA: // ANNN Store memory address NNN in register I
		fmt.Println("in 0xA")
		// nnn := opcode & 0x0FFF
		c.i = opcode & 0x0FFF
	case 0xB: // BNNN Jump to address NNN + V0
		fmt.Println("in 0xB")
		nnn := opcode & 0x0FFF
		c.pc = nnn + uint16(c.v[0])
	case 0xC: // CXNN Set Vx to a random number with a mask of NN
		fmt.Println("in 0xC")
		x := opcode & 0x0F00 >> 8
		nn := opcode & 0x00FF
		c.v[x] = uint8(rand.Intn(256)) & uint8(nn)
	case 0xD: // DXYN
		fmt.Println("in 0xD")
		x := opcode & 0x0F00 >> 8
		y := opcode & 0x00F0 >> 4
		// Number of bytes for the sprite that we have to load from memory.
		n := opcode & 0x000F
		xpos := c.v[x]
		ypos := c.v[y]

		c.v[0xF] = 0
		// Get all the bytes for the sprite.
		for yy := uint16(0); yy < n; yy++ {
			sbyte := c.memory[c.i+yy]
			//Go through all the pixels(bits) in the sprite byte.
			for xx := uint16(0); xx < 8; xx++ {
				// If the sprite pixel is set then check the value of the pixel on our screen as well.
				if spix := sbyte & (0x80 >> xx); spix == 1 {
					// Calculate the index of the screen pixel matching this sprite pixel's position.
					//idx := (uint16(xpos)+xx)%screenWidth + ((uint16(ypos)+yy)%screenHeight)*screenWidth
					idx := (uint16(xpos) + xx) + (uint16(ypos)+yy)*screenWidth
					// If the screen pixel is also set then set VF to 1.
					if c.screen[idx] == 0xFFFFFFFF {
						c.v[0xF] = 1
					}
					c.screen[idx] ^= 0xFFFFFFFF
					// TODO: set the draw flag so that we can issue a draw call in our rendering code.
				}
			}
		}
	case 0xE: // EX[NN] - We are only interested in X here, switch on NN.
		fmt.Println("in 0xE")
		x := opcode & 0x0F00 >> 8
		switch nn := opcode & 0x0FF; nn {
		case 0x9E: //  EX9E Skip the next instruction if the key stored in Vx is pressed.
			if key := c.v[x]; c.keypad[key] == 1 {
				c.pc += 2
			}
		case 0xA1: //  EXA1 Skip the next instruction if the key stored in Vx is not pressed.
			if key := c.v[x]; c.keypad[key] == 0 {
				c.pc += 2
			}
		default:
			return ErrUnknownOpcode
		}
	case 0xF: // FX[NN] We are interested in X, act accordingly based on NN.
		fmt.Println("in 0xF")
		x := opcode & 0x0F00 >> 8
		switch nn := opcode & 0x00FF; nn {
		case 0x07: // FX07 Store the current value of the delay timer in Vx.
			c.v[x] = c.delayTimer
		case 0x0A: // FX0A Wait for a keypress and store the result in Vx.
			// Keep rewinding the program counter until one of the keys in the keypad gets set to 0x1.
			c.pc -= 2
			// TODO: Use a switch statement here, extract the code and put it in a separate function.
			for i := uint8(0); i < 16; i++ {
				if c.keypad[i] == 1 {
					c.v[x] = i
					// Move to the next opcode.
					c.pc = +2
				}
			}
		case 0x15: // FX15 Set the delay timer to the value of Vx.
			c.delayTimer = c.v[x]
		case 0x18: // FX18 Set the sound timer to the value of Vx.
			c.soundTimer = c.v[x]
		case 0x1E: // FX1E Add the value stored in Vx to I.
			c.i += uint16(c.v[x])
		case 0x29: // FX29 Set I to the memory address of one of the 16 built-in characters stored in Vx.
			// So if Vx == 0, we need to set I to the address of 0 character(sprite)
			c.i = fontAddress + uint16(c.v[x])*fontSize

		case 0x33: // FX33 Store the binary-coded decimal equivalent of the value stored in register Vx at addresses I, I + 1, and I + 2
			c.memory[c.i] = c.v[x] / 100
			c.memory[c.i+1] = c.v[x] / 10 % 10
			c.memory[c.i+2] = c.v[x] % 100 % 10
		case 0x55: // FX55 Store the values of V0 to Vx inclusive in memory starting at address I. I is set to I + X + 1 after operation
			for i := uint16(0); i <= x; i++ {
				c.memory[c.i+i] = c.v[i]
			}
			// NOTE: https://github.com/mattmikolay/chip-8/wiki/CHIP%E2%80%908-Instruction-Set#notes
			// Update I.
			//c.i = c.i + x + 1
		case 0x65: // FX65 Fill V0 to Vx inclusive with the values stored in memory starting at address I. I is set to I + X + 1 after operation
			for i := uint16(0); i <= x; i++ {
				c.v[i] = c.memory[c.i+i]
			}
			// NOTE: https://github.com/mattmikolay/chip-8/wiki/CHIP%E2%80%908-Instruction-Set#notes
			// Update I.
			//c.i = c.i + x + 1
		default:
			return ErrUnknownOpcode
		}
	default:
		return ErrUnknownOpcode
	}
	return nil
}
