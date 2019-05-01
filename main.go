// +build js,wasm

package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"syscall/js"
)

type Canvas struct {
	height  float64
	width   float64
	element js.Value
}

func NewCanvas(width float64, height float64, element js.Value) *Canvas {

	return &Canvas{
		height:  height,
		width:   width,
		element: element,
	}
}

func (canvas *Canvas) Render(img *image.RGBA) {
	canvas.element.Set("height", canvas.height)
	canvas.element.Set("width", canvas.width)
	ctx := canvas.element.Call("getContext", "2d")
	canvasData := ctx.Call("createImageData", canvas.width, canvas.height)
	canvasData.Get("data").Call("set", js.TypedArrayOf(img.Pix))

	ctx.Call("putImageData", canvasData, 0, 0)
}

func loadImage() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, 700, 900))
	white := color.RGBA{255, 255, 255, 255}
	draw.Draw(img, img.Bounds(), image.NewUniform(white), image.ZP, draw.Src)
	return img
}

func main() {

	el := js.Global().Get("document").Call("getElementById", "canvas")
	canvas := NewCanvas(700, 900, el)

	img := loadImage()

	canvas.Render(img)

	resizeCallback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		fmt.Println("resize")
		return nil
	})
	defer renderFrame.Release()

	js.Global().Call("addEventListener", "resize", resizeCallback)

	done := make(chan struct{}, 0)
	<-done
}
