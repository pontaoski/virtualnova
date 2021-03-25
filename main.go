package main

import (
	"fmt"
	"image/color"
	"io/ioutil"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
)

type VirtualNovaVM struct {
	vm vm
}

func (v *VirtualNovaVM) Update() error {
	for i := 0; i < 500_000; i++ {
		v.vm.tick()
	}
	return nil
}

func (v *VirtualNovaVM) draw() {
}

func (v *VirtualNovaVM) Draw(screen *ebiten.Image) {
	const base = ramsize - screenSizeBytes
	const globalBase = romsize + base

	screen.Fill(color.White)
	for hPX := 0; hPX < screenWidth; hPX++ {
		for vPX := 0; vPX < screenHeight; vPX++ {
			bs := (base + memory(hPX) + memory(vPX*screenWidth))

			r, g, b := bs+0, bs+1, bs+2

			c := color.RGBA{
				R: v.vm.ram[r],
				G: v.vm.ram[g],
				B: v.vm.ram[b],
				A: 255,
			}

			if c.R != 0 || c.G != 0 || c.B != 0 {
				fmt.Printf("color: %+v\n", c)
			}

			screen.Set(hPX, vPX, c)
		}
	}
}

func (g *VirtualNovaVM) Layout(outsideWidth, outsideHeight int) (w, h int) {
	return screenWidth, screenHeight
}

func main() {
	vm := VirtualNovaVM{}

	data, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	copy(vm.vm.rom[:], data)

	ebiten.SetMaxTPS(60)
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Virtual Nova")

	if err := ebiten.RunGame(&vm); err != nil {
		panic(err)
	}
}
