package main

import (
	"fmt"
	"github.com/muesli/clusters"
	"github.com/muesli/kmeans"
	"syscall/js"
	"github.com/lucasb-eyer/go-colorful"
)

func colorAvg(data []float64) float64 {
	sum := 0.0
	for _, c := range data {
		sum += c
	}
	avg := sum / float64(len(data))
	return avg
}

func pixelate(cc clusters.Clusters, colorIndexes map[string][]int) {
	canvas := js.Global().Get("document").Call("getElementById", "canvas-pixelate")
	ctx := canvas.Call("getContext", "2d")

	width := 400
	height := 400
	grid := 10

	for _, c := range cc {
		for _, org := range c.Observations {

			//key := fmt.Sprintf("%d%d%d", int(org.Coordinates()[0]), int(org.Coordinates()[1]), int(org.Coordinates()[2]))
			key := fmt.Sprintf("%d%d%d", org.Coordinates()[0], org.Coordinates()[1], org.Coordinates()[2])
			if colorIndexes[key] != nil {
				for _, index := range colorIndexes[key] {
					x := index / (width / grid) * grid
					y := index % (height / grid) * grid
					c := colorful.Lab(c.Center[0], c.Center[1], c.Center[2])
					//ctx.Set("fillStyle", fmt.Sprintf("rgb(%d, %d, %d) \n", int(c.Center[0]), int(c.Center[1]), int(c.Center[2])))
					ctx.Set("fillStyle", c.Hex())
					ctx.Call("fillRect", x, y, grid, grid)
				}
				delete(colorIndexes, key)
			}
		}
	}

}

func convertLab(hex string) (float64, float64, float64) {
	//fmt.Printf("%s\n", hex)
	c, err := colorful.Hex(hex)
	if err != nil {
		//fmt.Printf("%+v\n", err)
	}
	l, a, b := c.Lab()
	return l, a, b
}

func main() {
	var d clusters.Observations

	canvas := js.Global().Get("document").Call("getElementById", "canvas")
	ctx := canvas.Call("getContext", "2d")
	width := 400
	height := 400
	grid := 10
	total := grid * grid

	var colorIndexes = map[string][]int{}

	for x := 0; x < width / grid; x++ {
		for y := 0; y < height / grid; y++ {
			// 分割したセルの中の色平均を計算
			cell := ctx.Call("getImageData", x*grid, y*grid, grid, grid)
			data := cell.Get("data")
			var rs = make([]float64, total)
			var gs = make([]float64, total)
			var bs = make([]float64, total)
			for p := 0; p < total; p++ {
				r := data.Index(p*4).Int()
				g := data.Index(p*4+1).Int()
				b := data.Index(p*4+2).Int()
				labL, labA, labB := convertLab(fmt.Sprintf("#%x%x%x", r, g, b))
				rs[p] = labL
				gs[p] = labA
				bs[p] = labB
			}
			ravg := colorAvg(rs)
			gavg := colorAvg(gs)
			bavg := colorAvg(bs)

			//c := colorful.Lab(ravg, gavg, bavg)
			//hex := c.Hex()
			//r, _ := strconv.ParseUint(hex[1:2], 16, 0)
			//g, _ := strconv.ParseUint(hex[3:4], 16, 0)
			//b, _ := strconv.ParseUint(hex[5:6], 16, 0)
			//key := fmt.Sprintf("%d%d%d", r, g, b)
			//fmt.Printf("%s\n", key)

			// 最後に色を置換する際に計算量を減らすためのメモ
			//key := fmt.Sprintf("%d%d%d", ravg, gavg, bavg)
			key := fmt.Sprintf("%d%d%d", ravg, gavg, bavg)
			index := x * (width / grid) + y
			colorIndexes[key] = append(colorIndexes[key], index)

			d = append(d, clusters.Coordinates{
				float64(ravg),
				float64(gavg),
				float64(bavg),
			})
		}
	}

	km := kmeans.New()
	cc, err := km.Partition(d, 8)

	if err != nil {
		panic(err)
	}

	pixelate(cc, colorIndexes)
}
