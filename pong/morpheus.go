package pong

import (
	"math"
	"math/rand"
)

// Vector is a vector
type Vector[T any] struct {
	Meta   T
	Word   string
	Vector []float64
	Avg    float64
	Stddev float64
	Next   *Vector[T]
}

// Config is the configuration for the morpheus algorithm
type Config struct {
	Iterations int
	Size       int
	Divider    int
	Accuracy   int
}

func MorpheusFast[T any](seed int64, config Config, vectors []*Vector[T]) {
	rng := rand.New(rand.NewSource(seed))
	results := make([][]float64, config.Iterations)
	width := config.Size
	cols, rows := width, width
	if config.Divider == 0 {
		rows = int(math.Ceil(math.Log2(float64(width))))
	} else {
		rows /= config.Divider
	}
	for iteration := range config.Iterations {
		a, b := NewMatrix(cols, rows, make([]float64, cols*rows)...),
			NewMatrix(cols, rows, make([]float64, cols*rows)...)
		index := 0
		for range a.Rows {
			for range a.Cols {
				a.Data[index] = rng.NormFloat64()
				b.Data[index] = rng.NormFloat64()
				index++
			}
		}
		aa := a.Softmax(1)
		bb := b.Softmax(1)
		/*graph := pagerank.NewGraph(len(vectors), rng)*/
		x := NewMatrix(cols, len(vectors), make([]float64, cols*len(vectors))...)
		y := NewMatrix(cols, len(vectors), make([]float64, cols*len(vectors))...)
		for i := range vectors {
			for ii, value := range vectors[i].Vector {
				x.Data[i*cols+ii] = value
			}
		}
		for i := range vectors {
			for ii, value := range vectors[i].Vector {
				y.Data[i*cols+ii] = value
			}
		}

		xx := aa.MulT(x).Unit()
		yy := bb.MulT(y).Unit()
		cs := yy.MulT(xx)
		result := PageRank(1.0, 8, rng.Uint32(), cs)
		/*for i := range cs.Rows {
			for ii := range cs.Cols {
				cs := cs.Data[i*cs.Cols+ii]
				if cs < 0 {
					cs = -cs
				}
				graph.Link(uint32(i), uint32(ii), float64(cs))
			}
		}

		result := make([]float64, len(vectors))
		graph.Rank(1.0, 1e-3, func(node int, rank float64) {
			result[node] = rank
		})*/
		results[iteration] = result.Data
	}
	for _, result := range results {
		for i, value := range result {
			vectors[i].Avg += value
		}
	}
	for i, value := range vectors {
		vectors[i].Avg = value.Avg / float64(config.Iterations)
	}

	for _, result := range results {
		for i, value := range result {
			diff := value - vectors[i].Avg
			vectors[i].Stddev += diff * diff
		}
	}
	for i, value := range vectors {
		vectors[i].Stddev = math.Sqrt(value.Stddev / float64(config.Iterations))
	}

	return
}

func Morpheus[T any](seed int64, config Config, vectors []*Vector[T], mutate ...func(cs *Matrix[float64])) [][]float64 {
	rng := rand.New(rand.NewSource(seed))
	results := make([][]float64, config.Iterations)
	width := 2 * config.Size
	cols, rows := width, width
	if config.Divider == 0 {
		rows = int(math.Ceil(math.Log2(float64(width))))
	} else {
		rows /= config.Divider
	}
	for iteration := range config.Iterations {
		a, b := NewMatrix(cols, rows, make([]float64, cols*rows)...),
			NewMatrix(cols, rows, make([]float64, cols*rows)...)
		index := 0
		for range a.Rows {
			for range a.Cols {
				a.Data[index] = rng.NormFloat64()
				b.Data[index] = rng.NormFloat64()
				index++
			}
		}
		aa := a.Softmax(1)
		bb := b.Softmax(1)
		/*graph := pagerank.NewGraph(len(vectors), rng)*/
		x := NewMatrix(cols, len(vectors), make([]float64, cols*len(vectors))...)
		y := NewMatrix(cols, len(vectors), make([]float64, cols*len(vectors))...)
		for i := range vectors {
			for ii, value := range vectors[i].Vector {
				if value < 0 {
					x.Data[i*cols+config.Size+ii] = -value
					continue
				}
				x.Data[i*cols+ii] = value
			}
		}
		for i := range vectors {
			for ii, value := range vectors[i].Vector {
				if value < 0 {
					y.Data[i*cols+config.Size+ii] = -value
					continue
				}
				y.Data[i*cols+ii] = value
			}
		}

		xx := aa.MulT(x).Unit()
		yy := bb.MulT(y).Unit()
		cs := yy.MulT(xx)
		/*if len(mutate) == 1 {
			mutate[0](&cs)
		}
		for i := range cs.Rows {
			for ii := range cs.Cols {
				cs := cs.Data[i*cs.Cols+ii]
				if cs < 0 {
					cs = -cs
				}
				graph.Link(uint32(i), uint32(ii), float64(cs))
			}
		}

		result := make([]float64, len(vectors))
		graph.Rank(1.0, 1e-3, func(node int, rank float64) {
			result[node] = rank
		})*/
		result := PageRank(1.0, 256, rng.Uint32(), cs)
		results[iteration] = result.Data
	}
	for _, result := range results {
		for i, value := range result {
			vectors[i].Avg += value
		}
	}
	for i, value := range vectors {
		vectors[i].Avg = value.Avg / float64(config.Iterations)
	}

	for _, result := range results {
		for i, value := range result {
			diff := value - vectors[i].Avg
			vectors[i].Stddev += diff * diff
		}
	}
	for i, value := range vectors {
		vectors[i].Stddev = math.Sqrt(value.Stddev / float64(config.Iterations))
	}

	cov := make([][]float64, len(vectors))
	for i := range cov {
		cov[i] = make([]float64, len(vectors))
	}
	for _, measures := range results {
		for i, v := range measures {
			for ii, vv := range measures {
				diff1 := vectors[i].Avg - v
				diff2 := vectors[ii].Avg - vv
				cov[i][ii] += diff1 * diff2
			}
		}
	}
	if len(results) > 0 {
		for i := range cov {
			for ii := range cov[i] {
				cov[i][ii] = cov[i][ii] / float64(len(results))
			}
		}
	}
	return cov
}

func MorpheusGramSchmidt[T any](seed int64, config Config, vectors []*Vector[T], mutate ...func(cs *Matrix[float64])) [][]float64 {
	rng := rand.New(rand.NewSource(seed))
	results := make([][]float64, config.Iterations)
	width := config.Size
	cols, rows := width, width
	if config.Divider == 0 {
		rows = int(math.Ceil(math.Log2(float64(width))))
	} else {
		rows /= config.Divider
	}
	for iteration := range config.Iterations {
		a, b := NewMatrix(cols, rows, make([]float64, cols*rows)...),
			NewMatrix(cols, rows, make([]float64, cols*rows)...)
		index := 0
		for range a.Rows {
			for range a.Cols {
				a.Data[index] = rng.NormFloat64()
				b.Data[index] = rng.NormFloat64()
				index++
			}
		}
		aa := a.GramSchmidt().T()
		bb := b.GramSchmidt().T()
		x := NewMatrix(cols, len(vectors), make([]float64, cols*len(vectors))...)
		y := NewMatrix(cols, len(vectors), make([]float64, cols*len(vectors))...)
		for i := range vectors {
			for ii, value := range vectors[i].Vector {
				x.Data[i*cols+ii] = value
			}
		}
		for i := range vectors {
			for ii, value := range vectors[i].Vector {
				y.Data[i*cols+ii] = value
			}
		}

		xx := aa.MulT(x).Unit()
		yy := bb.MulT(y).Unit()
		cs := yy.MulT(xx)
		if len(mutate) == 1 {
			mutate[0](&cs)
		}
		accuracy := 256
		if config.Accuracy > 0 {
			accuracy = config.Accuracy
		}
		result := PageRank(1.0, accuracy, rng.Uint32(), cs)
		r := make([]float64, len(result.Data))
		for key, value := range result.Data {
			r[key] = float64(value)
		}
		results[iteration] = r
	}
	for _, result := range results {
		for i, value := range result {
			vectors[i].Avg += value
		}
	}
	for i, value := range vectors {
		vectors[i].Avg = value.Avg / float64(config.Iterations)
	}

	for _, result := range results {
		for i, value := range result {
			diff := value - vectors[i].Avg
			vectors[i].Stddev += diff * diff
		}
	}
	for i, value := range vectors {
		vectors[i].Stddev = math.Sqrt(value.Stddev / float64(config.Iterations))
	}

	cov := make([][]float64, len(vectors))
	for i := range cov {
		cov[i] = make([]float64, len(vectors))
	}
	for _, measures := range results {
		for i, v := range measures {
			for ii, vv := range measures {
				diff1 := vectors[i].Avg - v
				diff2 := vectors[ii].Avg - vv
				cov[i][ii] += diff1 * diff2
			}
		}
	}
	if len(results) > 0 {
		for i := range cov {
			for ii := range cov[i] {
				cov[i][ii] = cov[i][ii] / float64(len(results))
			}
		}
	}
	return cov
}

func Morpheus2[T any](seed int64, config Config, vectors []*Vector[T], g map[string]map[string]uint64) [][]float64 {
	rng := rand.New(rand.NewSource(seed))
	results := make([][]float64, config.Iterations)
	width := 2 * config.Size
	cols, rows := width, width
	if config.Divider == 0 {
		rows = int(math.Ceil(math.Log2(float64(width))))
	} else {
		rows /= config.Divider
	}
	for iteration := range config.Iterations {
		a, b := NewMatrix(cols, rows, make([]float64, cols*rows)...),
			NewMatrix(cols, rows, make([]float64, cols*rows)...)
		index := 0
		for range a.Rows {
			for range a.Cols {
				a.Data[index] = rng.NormFloat64()
				b.Data[index] = rng.NormFloat64()
				index++
			}
		}
		aa := a.Softmax(1)
		bb := b.Softmax(1)
		/*graph := pagerank.NewGraph(len(vectors), rng)*/
		x := NewMatrix(cols, len(vectors), make([]float64, cols*len(vectors))...)
		y := NewMatrix(cols, len(vectors), make([]float64, cols*len(vectors))...)
		for i := range vectors {
			for ii, value := range vectors[i].Vector {
				if value < 0 {
					x.Data[i*cols+config.Size+ii] = -value
					continue
				}
				x.Data[i*cols+ii] = value
			}
		}
		for i := range vectors {
			for ii, value := range vectors[i].Vector {
				if value < 0 {
					y.Data[i*cols+config.Size+ii] = -value
					continue
				}
				y.Data[i*cols+ii] = value
			}
		}

		xx := aa.MulT(x).Unit()
		yy := bb.MulT(y).Unit()
		cs := yy.MulT(xx)
		result := PageRank(1.0, 256, rng.Uint32(), cs)
		/*for i := range cs.Rows {
			for ii := range cs.Cols {
				from := g[vectors[i].Word]
				if from != nil {
					to := from[vectors[ii].Word]
					if to > 0 {
						graph.Link(uint32(i), uint32(ii), float64(cs.Data[i*cs.Cols+ii]))
					}
				}
			}
		}

		result := make([]float64, len(vectors))
		graph.Rank(1.0, 1e-3, func(node int, rank float64) {
			result[node] = rank
		})*/
		results[iteration] = result.Data
	}
	for _, result := range results {
		for i, value := range result {
			vectors[i].Avg += value
		}
	}
	for i, value := range vectors {
		vectors[i].Avg = value.Avg / float64(config.Iterations)
	}

	for _, result := range results {
		for i, value := range result {
			diff := value - vectors[i].Avg
			vectors[i].Stddev += diff * diff
		}
	}
	for i, value := range vectors {
		vectors[i].Stddev = math.Sqrt(value.Stddev / float64(config.Iterations))
	}

	cov := make([][]float64, len(vectors))
	for i := range cov {
		cov[i] = make([]float64, len(vectors))
	}
	for _, measures := range results {
		for i, v := range measures {
			for ii, vv := range measures {
				diff1 := vectors[i].Avg - v
				diff2 := vectors[ii].Avg - vv
				cov[i][ii] += diff1 * diff2
			}
		}
	}
	if len(results) > 0 {
		for i := range cov {
			for ii := range cov[i] {
				cov[i][ii] = cov[i][ii] / float64(len(results))
			}
		}
	}
	return cov
}

func Morpheus3[T any](seed int64, config Config, vectors []*Vector[T]) [][]float64 {
	//rng := RNG(1)
	rng2 := rand.New(rand.NewSource(1))
	results := make([][]float64, config.Iterations)
	width := 2 * config.Size
	cols, rows := width, width
	if config.Divider == 0 {
		rows = int(math.Ceil(math.Log2(float64(width))))
	} else {
		rows /= config.Divider
	}
	a, b := NewMatrix(cols, rows, make([]float64, cols*rows)...),
		NewMatrix(cols, rows, make([]float64, cols*rows)...)
	index := 0
	for range a.Rows {
		for range a.Cols {
			a.Data[index] = rng2.NormFloat64()
			b.Data[index] = rng2.NormFloat64()
			index++
		}
	}
	aa := a.Softmax(1)
	bb := b.Softmax(1)
	/*graph := pagerank.NewGraph(len(vectors), rng2)*/
	x := NewMatrix(cols, len(vectors), make([]float64, cols*len(vectors))...)
	y := NewMatrix(cols, len(vectors), make([]float64, cols*len(vectors))...)
	for i := range vectors {
		for ii, value := range vectors[i].Vector {
			if value < 0 {
				x.Data[i*cols+config.Size+ii] = -value
				continue
			}
			x.Data[i*cols+ii] = value
		}
	}
	for i := range vectors {
		for ii, value := range vectors[i].Vector {
			if value < 0 {
				y.Data[i*cols+config.Size+ii] = -value
				continue
			}
			y.Data[i*cols+ii] = value
		}
	}

	xx := aa.MulT(x).Unit()
	yy := bb.MulT(y).Unit()
	cs := yy.MulT(xx)
	for iteration := range config.Iterations {
		/*for i := range cs.Rows {
			for ii := range cs.Cols {
				graph.Link(uint32(i), uint32(ii), float64(cs.Data[i*cs.Cols+ii])+.01*float64(rng.Float32()))
			}
		}

		result := make([]float64, len(vectors))
		graph.Rank(1.0, 1e-3, func(node int, rank float64) {
			result[node] = rank
		})*/
		result := PageRank(1.0, 256, rng2.Uint32(), cs)
		results[iteration] = result.Data
	}
	for _, result := range results {
		for i, value := range result {
			vectors[i].Avg += value
		}
	}
	for i, value := range vectors {
		vectors[i].Avg = value.Avg / float64(config.Iterations)
	}

	for _, result := range results {
		for i, value := range result {
			diff := value - vectors[i].Avg
			vectors[i].Stddev += diff * diff
		}
	}
	for i, value := range vectors {
		vectors[i].Stddev = math.Sqrt(value.Stddev / float64(config.Iterations))
	}

	cov := make([][]float64, len(vectors))
	for i := range cov {
		cov[i] = make([]float64, len(vectors))
	}
	for _, measures := range results {
		for i, v := range measures {
			for ii, vv := range measures {
				diff1 := vectors[i].Avg - v
				diff2 := vectors[ii].Avg - vv
				cov[i][ii] += diff1 * diff2
			}
		}
	}
	if len(results) > 0 {
		for i := range cov {
			for ii := range cov[i] {
				cov[i][ii] = cov[i][ii] / float64(len(results))
			}
		}
	}
	return cov
}

func MorpheusMarkov[T any, F Float](seed int64, config Config, vectors []*Vector[T]) Matrix[F] {
	rng := rand.New(rand.NewSource(seed))
	width := config.Size
	cols, rows := width, width
	if config.Divider == 0 {
		rows = int(math.Ceil(math.Log2(float64(width))))
	} else {
		rows /= config.Divider
	}

	x := NewMatrix(cols, len(vectors), make([]F, cols*len(vectors))...)
	y := NewMatrix(cols, len(vectors), make([]F, cols*len(vectors))...)
	for i := range vectors {
		for ii, value := range vectors[i].Vector {
			x.Data[i*cols+ii] = F(value)
		}
	}
	for i := range vectors {
		for ii, value := range vectors[i].Vector {
			y.Data[i*cols+ii] = F(value)
		}
	}

	xx := x.Unit()
	yy := y.Unit()
	adj := yy.MulT(xx)
	/*for i := range adj.Cols {
		adj.Data[i*adj.Cols+i] = 0
	}*/
	return PageRankMarkov(.85, 1024, rng.Uint32(), adj)
}
