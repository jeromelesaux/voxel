package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"

	"golang.org/x/image/bmp"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"github.com/jeromelesaux/martine/constants"
	"github.com/nfnt/resize"
)

// concat results
// ffmpeg -i images/%03d.png -vcodec mpeg4 test.mp4
var (
	count        = 0
	tabCos       [256]uint16
	taille_x     uint8  = 80
	taille_y     uint8  = 50
	window_x     uint16 = 320
	window_y     uint16 = 200
	heightBitmap [128 * 128]byte
	height       uint16
	angle        byte          = 0x30
	x            uint16        = 0x4000
	y            uint16        = 0x4000
	RgbCPC32     color.Palette = []color.Color{
		constants.Black.Color,
		constants.Blue.Color,
		constants.BrightBlue.Color,
		constants.Red.Color,
		constants.BrightRed.Color,
		constants.Magenta.Color,
		constants.Mauve.Color,
		constants.Green.Color,
		constants.Cyan.Color,
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
		tabCos[i] = uint16(math.Cos(float64(i)/40.7436654315) * 256)
	}
}

func initHeightBitmap() {
	f, err := os.Open("HEIGHT.BMP")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot open file HEIGHT.BMP error :%v\n", err)
		os.Exit(-1)
	}
	defer f.Close()
	h, err := bmp.Decode(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot decode file HEIGHT.BMP error :%v\n", err)
		os.Exit(-1)
	}
	for x0 := 0; x0 < h.Bounds().Max.X; x0++ {
		for y0 := 0; y0 < h.Bounds().Max.Y; y0++ {
			r, g, b, _ := h.At(x0, y0).RGBA()
			//	fmt.Fprintf(os.Stdout, "x:%d,y:%d,index:%d,value:%.2x\n", x0, y0, x0+(128*y0), uint8(r>>8+g>>8+b>>8))
			heightBitmap[x0+(128*y0)] = uint8(r>>8 + g>>8 + b>>8)
		}
	}
}

func drawView(filename string) {
	f, err := os.Create(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot create file (%s) error :%v\n", filename, err)
		return
	}
	defer f.Close()
	im := image.NewNRGBA(image.Rect(0, 0, int(taille_x), int(taille_y)))
	draw.Draw(im, im.Bounds(), &image.Uniform{RgbCPC32[0]}, image.Point{X: 0, Y: 0}, draw.Src)
	var col color.Color
	memBitmap = make([]byte, uint16(taille_x)*uint16(taille_y))

	p := ((y >> 1) & 0x3F80) + (x >> 9)
	height = uint16(taille_y + heightBitmap[p]) // Hauteur à la position x,y
	fmt.Fprintf(os.Stdout, "height:%d\n", height)
	if height < 150 {
		height += 150 - height + 10
		// height += 50
	}
	if height > 250 {
		height -= 250 - height + 10
	}
	//	height += 50
	// height = 150

	var sX uint8
	for sX = 0; sX < taille_x; sX++ {
		a := angle + sX - taille_x/2
		deltax := tabCos[a]
		a += 64
		deltay := tabCos[a]
		minY := uint16(taille_y)
		tx := x
		ty := y
		clear := false
		for d := 1; d < 64; d++ {
			tx += deltax
			ty += deltay
			o := ((ty & 0x7F00) >> 1) + (tx >> 9)
			c := heightBitmap[o] >> 4 // Hauteur à tracer
			hl := height - uint16(c)
			var y1 uint16
			if hl > 0 {
				y1 = (hl << 3) / uint16(d)
			}

			if y1 < uint16(minY) {
				if !clear {
					clear = true
					col = RgbCPC32[0]
					var y0 uint16
					for y0 = 0; y0 < minY; y0++ {
						im.Set(int(sX), int(y0), col)
						memBitmap[uint16(sX)+(uint16(taille_x)*y0)] = 0
					}
				}

				var y0 uint16
				for y0 = y1; y0 < minY; y0++ {
					col = RgbCPC32[c]
					im.Set(int(sX), int(y0), col)
					memBitmap[uint16(sX)+(uint16(taille_x)*y0)] = c
				}

				minY = y1
				/*	if minY > 1 {
					break
				}*/
				if minY <= 0 {
					break
				}
			}
		}
	}

	// Debug

	msg := fmt.Sprintf("x=0x%04X  y=0x%04X  p=0x%04X ", x, y, p)
	imResized := resize.Resize(uint(window_x), uint(window_y), im, resize.Lanczos3)
	if err := png.Encode(f, imResized); err != nil {
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
		x -= uint16(tabCos[angle&0xFF])
		y -= uint16(tabCos[(angle+64)&0xFF])
	case fyne.KeyUp:
		x += uint16(tabCos[angle&0xFF])
		y += uint16(tabCos[(angle+64)&0xFF])
	case fyne.KeyLeft:
		angle -= 4
	}
	filename := fmt.Sprintf("images/%.3d.png", count)
	drawView(filename)
	setImage(filename)
	count++
}

var w fyne.Window

func setImage(filename string) {
	im := canvas.NewImageFromFile(filename)
	//	im.FillMode = canvas.ImageFillStretch
	scroll := container.NewScroll(im)
	scroll.Resize(fyne.NewSize(float32(window_x), float32(window_y)))
	w.SetContent(scroll)
}

func main() {
	a := app.New()
	w = a.NewWindow("Voxel")
	os.RemoveAll("images")
	os.Mkdir("images", 0o755)
	initTabCos()
	initHeightBitmap()
	drawView("images/000.png")
	setImage("images/000.png")
	w.Canvas().SetOnTypedKey(KeyDown)
	w.Resize(fyne.NewSize(float32(window_x), float32(window_y)))
	w.ShowAndRun()
}
