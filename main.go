package main

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	_ "image/png"
	"math"
	"os"

	"github.com/pborman/getopt/v2"
	"golang.org/x/image/draw"
)

// types
type byteSliceAsImage interface {
	ColorModel() color.Model
	Bounds() image.Rectangle
	At(x, y int) color.Color
}

type ImageData struct {
	ImageContents []byte
	ImageWidth    int
}

func (img ImageData) ColorModel() color.Model { return color.RGBAModel }

func (img ImageData) Bounds() image.Rectangle {
	width := img.ImageWidth
	height := len(img.ImageContents) / 4 / width
	rect := image.Rectangle{image.Point{0, 0}, image.Point{width, height}}
	return rect
}

func (img ImageData) At(x, y int) color.Color {
	point := y*img.ImageWidth + x
	var color color.Color = color.RGBA{uint8(img.ImageContents[point*4]), uint8(img.ImageContents[point*4+1]), uint8(img.ImageContents[point*4+2]), uint8(255)}
	return color
}

// matrices
var horizontal = [3][3]int{
	{-1, 0, 1},
	{-1, 0, 1},
	{-1, 0, 1}}

var vertical = [3][3]int{
	{-1, -1, -1},
	{0, 0, 0},
	{1, 1, 1}}

// init variables
var (
	outPath = "./output.jpg"
)

func init() {
	getopt.HelpColumn = 20

	getopt.SetParameters("file")
	getopt.FlagLong(&outPath, "out", 'o', "Set output path")
	optHelp := getopt.BoolLong("help", 'h', "Display help")

	getopt.Parse()

	if *optHelp {
		getopt.Usage()
		os.Exit(0)
	} else if getopt.NArgs() != 1 {
		fmt.Println("Error: Provide one input file")
		fmt.Println()
		getopt.Usage()
		os.Exit(1)
	}
}

func main() {

	path := getopt.Arg(0)

	f, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	b := img.Bounds()
	converted := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(converted, converted.Bounds(), img, b.Min, draw.Src)
	processImg(grayscale(converted))
}

func processImg(img *image.NRGBA) {
	f, _ := os.Create(outPath)
	width := img.Rect.Size().X
	height := img.Rect.Size().Y
	pixels := img.Pix
	leng := len(pixels)

	newFrame := make([]byte, leng)
	for y := 1; y < height-1; y += 1 {
		for x := 1; x < width-1; x += 1 {
			idx1 := img.PixOffset(x-1, y-1)
			idx2 := img.PixOffset(x, y-1)
			idx3 := img.PixOffset(x+1, y-1)
			idx4 := img.PixOffset(x-1, y)
			idx5 := img.PixOffset(x, y)
			idx6 := img.PixOffset(x+1, y)
			idx7 := img.PixOffset(x-1, y+1)
			idx8 := img.PixOffset(x, y+1)
			idx9 := img.PixOffset(x+1, y+1)

			horizontalGradient :=
				(horizontal[0][0] * int(pixels[idx1])) +
					(horizontal[0][1] * int(pixels[idx2])) +
					(horizontal[0][2] * int(pixels[idx3])) +
					(horizontal[1][0] * int(pixels[idx4])) +
					(horizontal[1][1] * int(pixels[idx5])) +
					(horizontal[1][2] * int(pixels[idx6])) +
					(horizontal[2][0] * int(pixels[idx7])) +
					(horizontal[2][1] * int(pixels[idx8])) +
					(horizontal[2][2] * int(pixels[idx9]))
			horizontalGradient = int(math.Abs(float64(horizontalGradient)))

			verticalGradient := (vertical[0][0] * int(pixels[idx1])) +
				(vertical[0][1] * int(pixels[idx2])) +
				(vertical[0][2] * int(pixels[idx3])) +
				(vertical[1][0] * int(pixels[idx4])) +
				(vertical[1][1] * int(pixels[idx5])) +
				(vertical[1][2] * int(pixels[idx6])) +
				(vertical[2][0] * int(pixels[idx7])) +
				(vertical[2][1] * int(pixels[idx8])) +
				(vertical[2][2] * int(pixels[idx9]))
			verticalGradient = int(math.Abs(float64(verticalGradient)))

			mag := byte(math.Sqrt(math.Pow(float64(horizontalGradient), 2.0) + math.Pow(float64(verticalGradient), 2.0)))
			newFrame[idx1] = mag
			newFrame[idx1+1] = mag
			newFrame[idx1+2] = mag
			newFrame[idx1+3] = 100

		}
	}
	var newImage byteSliceAsImage = ImageData{newFrame, width}
	jpeg.Encode(f, newImage, nil)
}

func grayscale(img *image.NRGBA) *image.NRGBA {
	width := img.Rect.Size().X
	pixels := img.Pix
	leng := len(pixels)

	var newFrame []byte
	x := 0
	y := 0
	for i := 0; i < leng; i += 4 {
		gray := (0.299*float64(pixels[i]) + 0.587*float64(pixels[i+1]) + 0.114*float64(pixels[i+2]))

		newFrame = append(newFrame, byte(gray), byte(gray), byte(gray), 255)

		if (x+1)%width != 0 {
			x++
		} else {
			y++
			x = 0
		}
	}
	var newImage byteSliceAsImage = ImageData{newFrame, width}
	b := newImage.Bounds()
	out := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(out, out.Bounds(), newImage, b.Min, draw.Src)
	return out
}
