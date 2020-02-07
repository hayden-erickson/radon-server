package transform

import (
	"sync"
	"math"
)

type SimpleProjection struct {
	RowsProcessed int
	Img [][]float64
	Mux sync.Mutex
	Wg sync.WaitGroup
}

func (sp *SimpleProjection) InitImg(total int) {
		sp.Mux.Lock()
		defer sp.Mux.Unlock()

		w := int(math.Ceil(math.Sqrt(2)*float64(total)))
		sp.Img = make([][]float64, w)

		for i := range sp.Img {
			sp.Img[i] = make([]float64, w)
		}
}

func toCartesian(t float64, s float64, theta float64) (int, int) {
	x := t*math.Cos(theta) - s*math.Sin(theta)
	y := t*math.Sin(theta) + s*math.Cos(theta)
	return int(x), int(y)
}

func (sp *SimpleProjection) BackProject( total int, theta float64, sinoRow []uint8) {
	sp.Wg.Add(1)

	go (func() {
		defer sp.Wg.Done()

		if sp.Img == nil {
			sp.InitImg(total)
		}

		N := float64(len(sinoRow))
		w := float64(total)
		for i, o := range sinoRow {
			// t is distance of beam from center
			t := .5*w - (float64(i) + .5) * w/N
			// s is distance along beam
			for s := -.5*w; s < .5*w; s++ {
				x, y := toCartesian(t, s, theta)
				x += len(sp.Img)/2
				y += len(sp.Img)/2

				sp.Mux.Lock()
				sp.Img[x][y] += float64(o)
				sp.Mux.Unlock()
			}
		}

		sp.Mux.Lock()
		sp.RowsProcessed++
		sp.Mux.Unlock()
	})()
}

func (sp *SimpleProjection) GetRowsProcessed() int {
	sp.Mux.Lock()
	defer sp.Mux.Unlock()
	return sp.RowsProcessed
}

func (sp *SimpleProjection) Reset() {
	sp.Mux.Lock()
	defer sp.Mux.Unlock()

	sp.Img = nil
	sp.RowsProcessed = 0
}

func normalize(img [][]float64) [][]uint8 {
	out := make([][]uint8, len(img))

	for i := range out {
		out[i] = make([]uint8, len(img))
	}


	max := 0.0

	for _, row := range img {
		for _, pix := range row {
			if pix > max {
				max = pix
			}
		}
	}

	for i, row := range img {
		for j := range row {
			out[i][j] = uint8(255 * img[i][j]/max)
		}
	}

	return out
}

func (sp *SimpleProjection) GetImg() [][]uint8 {
	sp.Wg.Wait()
	sp.Mux.Lock()
	defer sp.Mux.Unlock()
	return normalize(sp.Img)
}
