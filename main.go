package main

import (
	"fmt"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/muesli/clusters"
	"github.com/muesli/kmeans"
	"syscall/js"
)
const width = 200
const height = 200
const grid = 5
const colorCluster = 3

func colorAvg(data []int) int {
	sum := 0
	for _, c := range data {
		sum += c
	}
	avg := sum / len(data)
	return avg
}

func pixelate(cc clusters.Clusters, colorIndexes map[string][]int, ctx js.Value) {
	for _, c := range cc {
		for _, org := range c.Observations {

			key := fmt.Sprintf("%.4f%.4f%.4f", org.Coordinates()[0], org.Coordinates()[1], org.Coordinates()[2])
			if colorIndexes[key] != nil {
				for _, index := range colorIndexes[key] {
					x := index / (width / grid) * grid
					y := index % (height / grid) * grid

					c2 := colorful.Lab(c.Center[0]/100, c.Center[1]/100, c.Center[2]/100)
					ctx.Set("fillStyle", c2.Hex())
					ctx.Call("fillRect", x, y, grid, grid)
				}
				delete(colorIndexes, key)
			}
		}
	}

}

func convertLab(r, g, b float64) (float64, float64, float64) {
	cc := colorful.Color{r, g, b}
	return cc.Lab()
}


func main() {
	c := make(chan struct{})
	var draw js.Func

	canvas := js.Global().Get("document").Call("getElementById", "canvas")
	ctx := canvas.Call("getContext", "2d")

	canvas2 := js.Global().Get("document").Call("getElementById", "canvas-pixelate")
	ctx2 := canvas2.Call("getContext", "2d")

	total := grid * grid

	d := make([]clusters.Observation, width / grid * height / grid)

	draw = js.FuncOf(func(this js.Value, args []js.Value) interface{} {

		var colorIndexes = map[string][]int{}

		for x := 0; x < width / grid; x++ {
			for y := 0; y < height / grid; y++ {


				// 分割したセルの中の色平均を計算
				cell := ctx.Call("getImageData", x*grid, y*grid, grid, grid)
				data := cell.Get("data")
				uint8Arr := js.Global().Get("Uint8Array").New(data)
				received := make([]byte, data.Get("length").Int())
				_ = js.CopyBytesToGo(received, uint8Arr)
				var rs = make([]int, total)
				var gs = make([]int, total)
				var bs = make([]int, total)
				for p := 0; p < total; p++ {
					rs[p] = int(received[p*4])
					gs[p] = int(received[p*4+1])
					bs[p] = int(received[p*4+2])
				}
				ravg := float64(colorAvg(rs))
				gavg := float64(colorAvg(gs))
				bavg := float64(colorAvg(bs))

				ravg, gavg, bavg = convertLab(ravg, gavg, bavg)

				// 最後に色を置換する際に計算量を減らすためのメモ
				key := fmt.Sprintf("%.4f%.4f%.4f", ravg, gavg, bavg)
				index := x * (width / grid) + y
				colorIndexes[key] = append(colorIndexes[key], index)

				d[index] = clusters.Coordinates{
					ravg,
					gavg,
					bavg,
				}
			}
		}

		km := kmeans.New()
		cc, err := km.Partition(d, colorCluster)

		if err != nil {
			panic(err)
		}

		pixelate(cc, colorIndexes, ctx2)

		js.Global().Call("requestAnimationFrame", draw)
		return nil
	})
	defer draw.Release()

	js.Global().Call("requestAnimationFrame", draw)
	<-c
}
