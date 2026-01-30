package ogimage

import (
	_ "embed"
	"image"
	"image/color"
	"image/png"
	"bytes"
	"math"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

//go:embed fonts/DMSans-Bold.ttf
var dmSansBoldTTF []byte

//go:embed fonts/DMSans-Regular.ttf
var dmSansRegularTTF []byte

var (
	faceBold72   font.Face
	faceBold28   font.Face
	faceRegular28 font.Face
	faceRegular26 font.Face
)

func init() {
	boldFont, err := opentype.Parse(dmSansBoldTTF)
	if err != nil {
		panic("ogimage: parse bold font: " + err.Error())
	}
	regularFont, err := opentype.Parse(dmSansRegularTTF)
	if err != nil {
		panic("ogimage: parse regular font: " + err.Error())
	}

	faceBold72, err = opentype.NewFace(boldFont, &opentype.FaceOptions{Size: 72, DPI: 72, Hinting: font.HintingFull})
	if err != nil {
		panic("ogimage: bold 72 face: " + err.Error())
	}
	faceBold28, err = opentype.NewFace(boldFont, &opentype.FaceOptions{Size: 28, DPI: 72, Hinting: font.HintingFull})
	if err != nil {
		panic("ogimage: bold 28 face: " + err.Error())
	}
	faceRegular28, err = opentype.NewFace(regularFont, &opentype.FaceOptions{Size: 28, DPI: 72, Hinting: font.HintingFull})
	if err != nil {
		panic("ogimage: regular 28 face: " + err.Error())
	}
	faceRegular26, err = opentype.NewFace(regularFont, &opentype.FaceOptions{Size: 26, DPI: 72, Hinting: font.HintingFull})
	if err != nil {
		panic("ogimage: regular 26 face: " + err.Error())
	}
}

const (
	imgW = 1200
	imgH = 630
)

var (
	colBg      = color.NRGBA{R: 0x09, G: 0x09, B: 0x0B, A: 0xFF} // #09090B
	colWhite   = color.NRGBA{R: 0xFA, G: 0xFA, B: 0xFA, A: 0xFF} // #FAFAFA
	colAmber   = color.NRGBA{R: 0xF5, G: 0x9E, B: 0x0B, A: 0xFF} // #F59E0B
	colZinc400 = color.NRGBA{R: 0xA1, G: 0xA1, B: 0xAA, A: 0xFF} // #A1A1AA
	colDark    = color.NRGBA{R: 0x09, G: 0x09, B: 0x0B, A: 0xFF} // text on white bg
)

// Render generates a 1200x630 OG image PNG with the given app name.
func Render(appName string) ([]byte, error) {
	img := image.NewNRGBA(image.Rect(0, 0, imgW, imgH))

	// 1. Dark background with subtle radial vignette.
	// Center is slightly lighter (#111114), edges fade to near-black (#050507).
	drawRadialBg(img)

	// 2. Grid pattern (80px cells, white at ~7% opacity).
	gridColor := color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 18} // ~7%
	for x := 0; x < imgW; x += 80 {
		drawVLine(img, x, 0, imgH, gridColor)
	}
	for y := 0; y < imgH; y += 80 {
		drawHLine(img, 0, imgW, y, gridColor)
	}

	// 3. Amber glow circles.
	drawGlow(img, 1050, 150, 400, colAmber, 30)
	drawGlow(img, 150, 550, 300, colAmber, 28)

	// Cool-blue glow at center for depth.
	drawGlow(img, 600, 315, 500, color.NRGBA{R: 0x40, G: 0x40, B: 0x5C, A: 0xFF}, 20)

	// 4. White rounded-rect logo box at (80,140) 56x56 with first letter.
	fillRoundedRect(img, 80, 140, 56, 56, 12, colWhite)

	firstLetter := string([]rune(appName)[0])
	drawTextCentered(img, faceBold28, colDark, 80, 140, 56, 56, firstLetter)

	// 5. App name text — vertically centered with the logo box.
	// Box is at y=140, height=56, so vertical center is y=168.
	// Use font metrics to align the text's visual center with the box center.
	nameMetrics := faceRegular28.Metrics()
	nameAscent := nameMetrics.Ascent.Ceil()
	nameDescent := nameMetrics.Descent.Ceil()
	nameTextH := nameAscent + nameDescent
	nameY := 140 + (56-nameTextH)/2 + nameAscent
	drawText(img, faceRegular28, colWhite, 152, nameY, appName, false)

	// 6. Headline.
	drawText(img, faceBold72, colWhite, 80, 300, "Bureaucracy", false)
	drawText(img, faceBold72, colWhite, 80, 390, "That Actually ", false)

	// Measure "That Actually " to position "Moves" in amber.
	advPx := measureText(faceBold72, "That Actually ")
	drawText(img, faceBold72, colAmber, 80+advPx, 390, "Moves", false)

	// 7. Subtitle.
	drawText(img, faceRegular26, colZinc400, 80, 470, "A no-nonsense task manager for approval workflows.", false)

	// Encode PNG.
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// drawText renders text onto img. If center is true, text is horizontally
// centered on the given x coordinate; otherwise x is the left edge.
func drawText(img *image.NRGBA, face font.Face, col color.NRGBA, x, y int, s string, center bool) {
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: face,
	}

	startX := fixed.I(x)
	if center {
		w := d.MeasureString(s)
		startX -= w / 2
	}

	d.Dot = fixed.Point26_6{
		X: startX,
		Y: fixed.I(y),
	}
	d.DrawString(s)
}

// drawTextCentered renders text centered both horizontally and vertically
// within the rectangle defined by (rx, ry, rw, rh).
func drawTextCentered(img *image.NRGBA, face font.Face, col color.NRGBA, rx, ry, rw, rh int, s string) {
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: face,
	}

	bounds, adv := d.BoundString(s)
	// Glyph pixel width and height from the bounding box.
	glyphW := adv.Ceil()
	glyphH := (bounds.Max.Y - bounds.Min.Y).Ceil()
	// bounds.Min.Y is negative (ascent above baseline).
	ascent := (-bounds.Min.Y).Ceil()

	d.Dot = fixed.Point26_6{
		X: fixed.I(rx + (rw-glyphW)/2),
		Y: fixed.I(ry + (rh-glyphH)/2 + ascent),
	}
	d.DrawString(s)
}

// measureText returns the advance width of s in pixels.
func measureText(face font.Face, s string) int {
	d := &font.Drawer{Face: face}
	return d.MeasureString(s).Ceil()
}

// drawRadialBg fills the image with a dark radial gradient — slightly
// lighter at center, fading to near-black at the edges.
func drawRadialBg(img *image.NRGBA) {
	cx, cy := float64(imgW)/2, float64(imgH)/2
	maxDist := math.Sqrt(cx*cx + cy*cy)

	// Inner color: #1C1C22, Outer color: #08080A
	for y := 0; y < imgH; y++ {
		for x := 0; x < imgW; x++ {
			dx := float64(x) - cx
			dy := float64(y) - cy
			t := math.Sqrt(dx*dx+dy*dy) / maxDist
			if t > 1 {
				t = 1
			}
			// Smooth step for a more natural falloff.
			t = t * t * (3 - 2*t)

			r := uint8(lerp(0x1C, 0x08, t))
			g := uint8(lerp(0x1C, 0x08, t))
			b := uint8(lerp(0x22, 0x0A, t))
			img.SetNRGBA(x, y, color.NRGBA{R: r, G: g, B: b, A: 0xFF})
		}
	}
}

func lerp(a, b uint8, t float64) float64 {
	return float64(a)*(1-t) + float64(b)*t
}

// fillRect fills a rectangle on img.
func fillRect(img *image.NRGBA, x0, y0, w, h int, col color.NRGBA) {
	for y := y0; y < y0+h; y++ {
		for x := x0; x < x0+w; x++ {
			img.SetNRGBA(x, y, col)
		}
	}
}

// fillRoundedRect draws a filled rounded rectangle.
func fillRoundedRect(img *image.NRGBA, x0, y0, w, h, r int, col color.NRGBA) {
	for y := y0; y < y0+h; y++ {
		for x := x0; x < x0+w; x++ {
			dx, dy := 0, 0

			// Check which corner region we're in.
			if x < x0+r && y < y0+r {
				dx, dy = x0+r-x, y0+r-y
			} else if x >= x0+w-r && y < y0+r {
				dx, dy = x-(x0+w-r-1), y0+r-y
			} else if x < x0+r && y >= y0+h-r {
				dx, dy = x0+r-x, y-(y0+h-r-1)
			} else if x >= x0+w-r && y >= y0+h-r {
				dx, dy = x-(x0+w-r-1), y-(y0+h-r-1)
			}

			if dx*dx+dy*dy <= r*r {
				img.SetNRGBA(x, y, col)
			}
		}
	}
}

// drawVLine draws a vertical line.
func drawVLine(img *image.NRGBA, x, y0, y1 int, col color.NRGBA) {
	if x < 0 || x >= imgW {
		return
	}
	for y := y0; y < y1; y++ {
		blendPixel(img, x, y, col)
	}
}

// drawHLine draws a horizontal line.
func drawHLine(img *image.NRGBA, x0, x1, y int, col color.NRGBA) {
	if y < 0 || y >= imgH {
		return
	}
	for x := x0; x < x1; x++ {
		blendPixel(img, x, y, col)
	}
}

// drawGlow draws a soft radial gradient circle (additive-like blend).
func drawGlow(img *image.NRGBA, cx, cy, radius int, col color.NRGBA, maxAlpha uint8) {
	r2 := float64(radius * radius)
	for y := cy - radius; y <= cy+radius; y++ {
		if y < 0 || y >= imgH {
			continue
		}
		for x := cx - radius; x <= cx+radius; x++ {
			if x < 0 || x >= imgW {
				continue
			}
			dx := float64(x - cx)
			dy := float64(y - cy)
			dist2 := dx*dx + dy*dy
			if dist2 > r2 {
				continue
			}
			// Smooth falloff using cosine interpolation.
			t := math.Sqrt(dist2) / float64(radius)
			alpha := uint8(float64(maxAlpha) * (1 - t*t))
			blendPixel(img, x, y, color.NRGBA{R: col.R, G: col.G, B: col.B, A: alpha})
		}
	}
}

// blendPixel alpha-blends src onto the existing pixel at (x,y).
func blendPixel(img *image.NRGBA, x, y int, src color.NRGBA) {
	if src.A == 0 {
		return
	}
	dst := img.NRGBAAt(x, y)

	sa := uint32(src.A)
	da := uint32(dst.A)
	oa := sa + da*(255-sa)/255

	if oa == 0 {
		return
	}

	r := (uint32(src.R)*sa + uint32(dst.R)*da*(255-sa)/255) / oa
	g := (uint32(src.G)*sa + uint32(dst.G)*da*(255-sa)/255) / oa
	b := (uint32(src.B)*sa + uint32(dst.B)*da*(255-sa)/255) / oa

	img.SetNRGBA(x, y, color.NRGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: uint8(oa)})
}
