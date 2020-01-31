package main

import (
	"fmt"
	"github.com/muesli/clusters"
	"github.com/muesli/kmeans"
	"syscall/js"
	"time"
)

func colorAvg(data []int) int {
	sum := 0
	for _, c := range data {
		sum += c
	}
	avg := sum / len(data)
	return avg
}

func pixelate(cc clusters.Clusters, colorIndexes map[string][]int) {
	canvas := js.Global().Get("document").Call("getElementById", "canvas-pixelate")
	ctx := canvas.Call("getContext", "2d")

	width := 400
	height := 400
	grid := 20

	for _, c := range cc {
		for _, org := range c.Observations {

			key := fmt.Sprintf("%d%d%d", int(org.Coordinates()[0]), int(org.Coordinates()[1]), int(org.Coordinates()[2]))
			if colorIndexes[key] != nil {
				for _, index := range colorIndexes[key] {
					x := index / (width / grid) * grid
					y := index % (height / grid) * grid
					ctx.Set("fillStyle", fmt.Sprintf("rgb(%d, %d, %d) \n", int(c.Center[0]), int(c.Center[1]), int(c.Center[2])))
					ctx.Call("fillRect", x, y, grid, grid)
				}
				delete(colorIndexes, key)
			}
		}
	}

}

func main() {
	for {
		time.Sleep(33 * time.Millisecond)
	var d clusters.Observations

	//fmt.Println("hogehoge")
	//fmt.Printf("step1 %d\n", time.Now().UnixNano() / int64(time.Millisecond))

	canvas := js.Global().Get("document").Call("getElementById", "canvas")
	ctx := canvas.Call("getContext", "2d")
	width := 400
	height := 400
	grid := 20

	var colorIndexes = map[string][]int{}

	//fmt.Printf("step2 %d\n", time.Now().UnixNano() / int64(time.Millisecond))

	//index := 0
	for x := 0; x < width / grid; x++ {
		for y := 0; y < height / grid; y++ {
			//fmt.Printf("step2-1 %d\n", time.Now().UnixNano() / int64(time.Millisecond))
			// 分割したセルの中の色平均を計算
			cell := ctx.Call("getImageData", x*grid, y*grid, grid, grid)
			data := cell.Get("data")
			var r = make([]int, 10)
			var g = make([]int, 10)
			var b = make([]int, 10)
			//var g []int
			//var b []int
			//fmt.Printf("step2-2 %d\n", time.Now().UnixNano() / int64(time.Millisecond))
			//for p := 0; p < grid * grid; p++ {
			// 適当に拾って計算量を削減するテスト
			target := []int{ 0, 11, 22, 33, 44, 55, 66, 77, 88, 99 }
			//for p := 0; p < 1; p++ {
			for i, p := range target {
				r[i] = data.Index(p*4).Int()
				g[i] = data.Index(p*4+1).Int()
				b[i] = data.Index(p*4+2).Int()
			}
			//fmt.Printf("step2-3 %d\n", time.Now().UnixNano() / int64(time.Millisecond))
			ravg := colorAvg(r)
			gavg := colorAvg(g)
			bavg := colorAvg(b)

			//fmt.Printf("step2-4 %d\n", time.Now().UnixNano() / int64(time.Millisecond))
			// 最後に色を置換する際に計算量を減らすためのメモ
			key := fmt.Sprintf("%d%d%d", ravg, gavg, bavg)
			index := x * (width / grid) + y
			//fmt.Println(index)
			colorIndexes[key] = append(colorIndexes[key], index)

			//fmt.Printf("step2-5 %d\n", time.Now().UnixNano() / int64(time.Millisecond))
			d = append(d, clusters.Coordinates{
				float64(ravg),
				float64(gavg),
				float64(bavg),
			})
			//index++

		}
	}

	//fmt.Printf("step3 %d\n", time.Now().UnixNano() / int64(time.Millisecond))
	km := kmeans.New()
	cc, err := km.Partition(d, 16)

	//fmt.Printf("step4 %d\n", time.Now().UnixNano() / int64(time.Millisecond))
	if err != nil {
		panic(err)
	}

	pixelate(cc, colorIndexes)
	//fmt.Printf("step5 %d\n", time.Now().UnixNano() / int64(time.Millisecond))


	}
}

// パフォーマンスがきついリアルタイムで出すには厳しいかもしれない
