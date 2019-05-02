//transform+build js,wasm

package main

import (
	"fmt"
	"github.com/anthonynsimon/bild/transform"
	"image"
	"image/color"
	"image/draw"
	"syscall/js"
)

var ZoomFactors = []float64{0.25, 0.33, 0.5, 0.66, 0.75, 0.8, 0.9, 1, 1.1, 1.25, 1.5, 1.75, 2, 2.5, 3}

type Page struct {
	width         int
	height        int
	rotationAngle float64
	zoomFactor    float64
	originalImg   image.Image
	canvas        *Canvas
}

func NewPage(src image.Image, width int, height int, canvas *Canvas) *Page {
	return &Page{
		width:         width,
		height:        height,
		rotationAngle: 0,
		zoomFactor:    1,
		originalImg:   src,
		canvas:        canvas,
	}
}

func (page *Page) Render() {
	width := int(float64(page.width) * page.zoomFactor)
	height := int(float64(page.height) * page.zoomFactor)

	img := transform.Resize(page.originalImg, width, height, transform.NearestNeighbor)
	img = transform.Rotate(img, page.rotationAngle, &transform.RotationOptions{ResizeBounds: true, Pivot: &image.Point{0, 0}})
	bounds := img.Bounds()

	fmt.Println("render", page.zoomFactor, page.rotationAngle, bounds.Max.X, bounds.Max.Y)
	page.canvas.Render(img, bounds.Max.X, bounds.Max.Y)
}

func (page *Page) Rotate() {
	page.rotationAngle += 90
	page.Render()
}

func (page *Page) ZoomIn() {
	newZoomFactor := ZoomFactors[len(ZoomFactors)-1]
	for i := len(ZoomFactors) - 1; i >= 0; i-- {
		if ZoomFactors[i] > page.zoomFactor {
			newZoomFactor = ZoomFactors[i]
		}
	}
	page.Zoom(newZoomFactor)
}

func (page *Page) ZoomOut() {
	newZoomFactor := ZoomFactors[0]
	for _, factor := range ZoomFactors {
		if factor < page.zoomFactor {
			newZoomFactor = factor
		}
	}
	page.Zoom(newZoomFactor)
}

func (page *Page) Zoom(factor float64) {
	page.zoomFactor = factor
	page.Render()
}

type Canvas struct {
	element js.Value
	onClick js.Func
}

func NewCanvas(element js.Value) *Canvas {
	return &Canvas{
		element: element,
	}
}

func (canvas *Canvas) Render(img *image.RGBA, width int, height int) {
	canvas.element.Set("width", width)
	canvas.element.Set("height", height)
	ctx := canvas.element.Call("getContext", "2d")
	canvasData := ctx.Call("createImageData", width, height)
	canvasData.Get("data").Call("set", js.TypedArrayOf(img.Pix))

	ctx.Call("putImageData", canvasData, 0, 0)
}

func (canvas *Canvas) AttachOnClick() {
	callback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		rect := args[0].Get("target").Call("getBoundingClientRect")
		x := args[0].Get("clientX").Int() - rect.Get("left").Int()
		y := args[0].Get("clientY").Int() - rect.Get("top").Int()

		x, y = canvas.ParseCoordinates(x, y)

		fmt.Println("click", x, y)
		return nil
	})
	canvas.onClick = callback
	canvas.element.Call("addEventListener", "click", callback)
}

func (canvas *Canvas) RemoveOnClick() {
	canvas.onClick.Release()
}

func (canvas *Canvas) ParseCoordinates(x int, y int) (int, int) {
	return x, y
}

func loadImage() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, 700, 900))
	white := color.RGBA{255, 255, 255, 255}
	draw.Draw(img, img.Bounds(), image.NewUniform(white), image.ZP, draw.Src)
	return img
}

func main() {
	el := js.Global().Get("document").Call("getElementById", "canvas")
	canvas := NewCanvas(el)
	canvas.AttachOnClick()
	defer canvas.RemoveOnClick()

	img := loadImage()
	page := NewPage(img, 700, 900, canvas)
	page.Render()

	resizeCallback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		fmt.Println("resize")
		return nil
	})
	defer resizeCallback.Release()
	js.Global().Call("addEventListener", "resize", resizeCallback)

	zoomIn := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		fmt.Println("zoomIn")
		page.ZoomIn()
		return nil
	})
	defer zoomIn.Release()
	js.Global().Set("zoomIn", zoomIn)

	zoomOut := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		fmt.Println("zoomOut")
		page.ZoomOut()
		return nil
	})
	defer zoomOut.Release()
	js.Global().Set("zoomOut", zoomOut)

	rotateClockwise := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		fmt.Println("rotateClockwise")
		page.Rotate()
		return nil
	})
	defer rotateClockwise.Release()
	js.Global().Set("rotateClockwise", rotateClockwise)

	done := make(chan struct{}, 0)
	<-done
}
