package main

import (
	"encoding/base64"
	"fmt"
	"image"
	_ "image/jpeg"
	"os"
	"strings"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
)

func run(img image.Image) {
	cfg := pixelgl.WindowConfig{
		Title:  "Captcha Viewer",
		Bounds: pixel.R(0, 0, float64(img.Bounds().Dx()), float64(img.Bounds().Dy())),
		VSync:  true,
	}

	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	picData := pixel.PictureDataFromImage(img)
	sprite := pixel.NewSprite(picData, picData.Bounds())

	for !win.Closed() {
		win.Clear(colornames.White)
		sprite.Draw(win, pixel.IM.Moved(win.Bounds().Center()))
		win.Update()
	}
}

func DisplayCaptcha(base64str string) {
	reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(base64str))
	img, _, err := image.Decode(reader)
	if err != nil {
		fmt.Println("Error decoding base64 string:", err)
		os.Exit(1)
	}

	pixelgl.Run(func() {
		run(img)
	})
}
