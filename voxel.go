package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/widget"
	"github.com/jeromelesaux/martine/constants"
)

var (
	count        = 0
	exitProg     = false
	tabCos       [256]uint
	taille_x     = 80
	taille_y     = 50
	window_x     = 320
	window_y     = 200
	heightBitmap [128 * 128]uint
	angle                      = 0x30
	x                          = 0x4000
	y                          = 0x4000
	RgbCPC32     color.Palette = []color.Color{
		constants.Black.Color,
		constants.Blue.Color,
		constants.BrightBlue.Color,
		constants.Red.Color,
		constants.BrightRed.Color,
		constants.Magenta.Color,
		constants.Purple.Color,
		constants.Green.Color,
		constants.BrightBlue.Color,
		constants.Yellow.Color,
		constants.White.Color,
		constants.PastelMagenta.Color,
		constants.BrightGreen.Color,
		constants.BrightCyan.Color,
		constants.BrightYellow.Color,
		constants.BrightWhite.Color,
	}

	/*	RgbCPC32 [32]int = [32]int{0x000000, // Noir              (0) -> #54
			0x00007F, // Bleu              (1) -> #44
			0x0000FF, // Bleu vif          (2) -> #55
			0x7F0000, // Rouge             (3) -> #5C
			0xFF0000, // Rouge vif         (6) -> #4C
			0x7F007F, // Magenta           (4) -> #58
			0xFF007F, // Pourpre           (7) -> #45
			0x007F00, // Vert              (9) -> #56
			0x007F7F, // Turquoise        (10) -> #46
			0x7F7F00, // Jaune            (12) -> #5E
			0x7F7F7F, // Blanc            (13) -> #40
			0xFF7FFF, // Magenta pastel   (17) -> #4F
			0x00FF00, // Vert vif         (18) -> #52
			0x00FFFF, // Turquoise vif    (20) -> #53
			0xFFFF00, // Jaune vif        (24) -> #49
			0xFFFFFF, // Blanc Brillant   (26) -> #4B
		}
	*/
	memBitmap []byte
)

func initTabCos() {
	for i := 0; i < 256; i++ {
		tabCos[i] = uint(math.Cos(float64(i)/40.7436654315) * 256)
	}

}

func drawView(filename string) {
	f, err := os.Create(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot create file (%s) error :%v\n", filename, err)
		return
	}
	defer f.Close()
	im := image.NewNRGBA(image.Rect(0, 0, window_x, window_y))
	draw.Draw(im, im.Bounds(), &image.Uniform{RgbCPC32[0]}, image.ZP, draw.Src)
	var col color.Color
	memBitmap = make([]byte, taille_x*taille_y)
	p := ((y >> 1) & 0x3F80) + (x >> 9)
	Height := taille_y + int(heightBitmap[p])
	for sX := 0; sX < taille_x; sX++ {
		a := angle + sX - taille_x/2
		deltax := tabCos[a&0xFF]
		deltay := tabCos[(a+64)&0xFF]
		minY := taille_y
		tx := x
		ty := y
		for d := 4; d < 256; d += 4 {
			tx += int(deltax)
			ty += int(deltay)
			o := ((ty >> 1) & 0x3F80) + (tx >> 9)
			c := heightBitmap[o]
			hl := Height - int(c)
			y1 := 0

			if hl > 0 {
				y1 = (hl << 6) / d
			}

			if y1 < minY {
				for y0 := y1; y0 < minY; y0++ {
					memBitmap[sX+taille_x*y0] = byte(c >> 4)
					col = RgbCPC32[int(c>>4)]
					for px := 1; px <= (window_x / taille_x); px++ {
						for py := 1; py <= (window_y / taille_y); py++ {
							im.Set((sX+taille_x)*px, y0*py, col)
						}
					}
				}

				minY = y1
				if minY != 1 {
					break
				}
			}
		}
	}

	// Debug

	msg := fmt.Sprintf("x=0x%04X  y=0x%04X  p=0x%04X ", x, y, p)
	if err := png.Encode(f, im); err != nil {
		fmt.Fprintf(os.Stderr, "Cannot encode image (%s) error :%v\n", filename, err)
	}
	fmt.Fprintf(os.Stdout, "%s\n", msg)
}

func KeyDown(key *fyne.KeyEvent) {

	switch key.Name {
	case fyne.KeyEscape:
		os.Exit(1)
	case fyne.KeyRight:
		angle += 4
	case fyne.KeyDown:
		x -= int(tabCos[angle&0xFF])
		y -= int(tabCos[(angle+64)&0xFF])
	case fyne.KeyUp:
		x += int(tabCos[angle&0xFF])
		y += int(tabCos[(angle+64)&0xFF])
	case fyne.KeyLeft:
		angle -= 4
	}
	filename := fmt.Sprintf("%.3d.png", count)
	drawView(filename)
	setImage(filename)
	count++
}

var w fyne.Window

func setImage(filename string) {
	im := canvas.NewImageFromFile(filename)
	im.FillMode = canvas.ImageFillContain
	scroll := widget.NewScrollContainer(im)
	scroll.Resize(fyne.NewSize(window_x, window_y))
	w.SetContent(scroll)
	w.Canvas().SetOnTypedKey(KeyDown)
	w.Resize(fyne.NewSize(window_x, window_y))
}

func main() {
	initTabCos()
	a := app.New()
	w = a.NewWindow("Voxel")

	drawView(fmt.Sprintf("000.png"))
	setImage("000.png")
	w.ShowAndRun()
}
