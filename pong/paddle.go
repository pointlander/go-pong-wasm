package pong

import (
	"fmt"
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/inpututil"
	"github.com/hajimehoshi/ebiten/text"
	"golang.org/x/image/font"
	"image/color"
	"math"
	"math/rand"
	"slices"
	"strconv"
)

// Paddle is a pong paddle
type Paddle struct {
	Position
	Score        int
	Speed        float32
	Width        int
	Height       int
	Color        color.Color
	Up           ebiten.Key
	Down         ebiten.Key
	Img          *ebiten.Image
	pressed      keysPressed
	scorePrinted scorePrinted
	UpV          Matrix[float64]
	DownV        Matrix[float64]
}

const (
	InitPaddleWidth  = 20
	InitPaddleHeight = 100
	InitPaddleShift  = 50
)

type keysPressed struct {
	up   bool
	down bool
}

type scorePrinted struct {
	score   int
	printed bool
	x       int
	y       int
}

// Neuron a neuron
type Neuron struct {
	Connections []int
	Vector      []float64
}

// Network is a neural network
type Network struct {
	Rng       *rand.Rand
	Width     int
	Embedding int
	Neurons   []Neuron
}

// NewNetwork creates a new neural network
func NewNetwork(width, embedding, size int) Network {
	neurons := make([]Neuron, size)
	for i := range neurons {
		neurons[i].Connections = make([]int, width)
		neurons[i].Vector = make([]float64, width+embedding)
	}
	return Network{
		Rng:       rand.New(rand.NewSource(1)),
		Width:     width,
		Embedding: embedding,
		Neurons:   neurons,
	}
}

// NeuralMode neural mode
func (n *Network) Iterate() {
	rng := n.Rng
	neurons := n.Neurons
	width := n.Width
	embedding := n.Embedding
	for i := range neurons {
		for ii := range neurons[i].Connections {
			next := rng.Intn(len(neurons))
			for next == i {
				next = rng.Intn(len(neurons))
			}
			neurons[i].Connections[ii] = next
		}
		for ii := range neurons[i].Vector[:width] {
			neurons[i].Vector[ii] = float64(rng.Intn(256))
		}
	}
	{
		for i := range neurons {
			next := rng.Intn(len(neurons))
			for next == i || slices.Contains(neurons[i].Connections[:], next) {
				next = rng.Intn(len(neurons))
			}
			vectors := make([]*Vector[Neuron], 6)
			index := 0
			for ii := range neurons[i].Connections {
				vector := Vector[Neuron]{}
				vector.Meta = neurons[neurons[i].Connections[ii]]
				vector.Vector = neurons[neurons[i].Connections[ii]].Vector
				vectors[ii] = &vector
				index++
			}
			{
				a := Vector[Neuron]{}
				a.Meta = neurons[next]
				a.Vector = neurons[next].Vector
				vectors[index] = &a
				index++
			}
			{
				a := Vector[Neuron]{}
				a.Meta = neurons[i]
				a.Vector = neurons[i].Vector
				vectors[index] = &a
				index++
			}
			config := Config{
				Iterations: 16,
				Size:       width + embedding,
				Divider:    1,
			}
			MorpheusFast(rng.Int63(), config, vectors)
			{
				max, index := 0.0, 0
				for i := range vectors[:len(vectors)-1] {
					if vectors[i].Stddev > max {
						max, index = vectors[i].Stddev, i
					}
				}
				if index != len(vectors)-2 {
					neurons[i].Connections[index] = next
				}
			}
		}
		fmt.Println(neurons)
		fmt.Println()
		previous, neuron := 0, 0
		for range 1024 {
			for i := range neurons[neuron].Vector[:width] {
				if neurons[neuron].Vector[i] > 128 {
					for i, value := range neurons[neuron].Vector[:width] {
						neurons[neuron].Vector[i] = math.Round(value / 2)
					}
					break
				}
			}
			sum := 0.0
			for _, value := range neurons[neuron].Vector[:width] {
				sum += value
			}
			total, index, selected := 0.0, 0, float64(rng.Intn(int(sum)))
			for i, value := range neurons[neuron].Vector[:width] {
				total += value
				if selected < total {
					index = i
					break
				}
			}
			for i, value := range neurons[neuron].Connections {
				if value == previous {
					neurons[neuron].Vector[i]++
					break
				}
			}
			previous, neuron = neuron, neurons[neuron].Connections[index]
		}
	}
}

func (p *Paddle) Update(screen *ebiten.Image) {
	_, h := screen.Size()

	if inpututil.IsKeyJustPressed(p.Up) {
		p.pressed.down = false
		p.pressed.up = true
	} else if inpututil.IsKeyJustReleased(p.Up) || !ebiten.IsKeyPressed(p.Up) {
		p.pressed.up = false
	}
	if inpututil.IsKeyJustPressed(p.Down) {
		p.pressed.up = false
		p.pressed.down = true
	} else if inpututil.IsKeyJustReleased(p.Down) || !ebiten.IsKeyPressed(p.Down) {
		p.pressed.down = false
	}

	if p.pressed.up {
		p.Y -= p.Speed
	} else if p.pressed.down {
		p.Y += p.Speed
	}

	if p.Y-float32(p.Height/2) < 0 {
		p.Y = float32(1 + p.Height/2)
	} else if p.Y+float32(p.Height/2) > float32(h) {
		p.Y = float32(h - p.Height/2 - 1)
	}
}

func (p *Paddle) PressUp(screen *ebiten.Image) {
	_, h := screen.Size()
	p.Y -= p.Speed
	if p.Y-float32(p.Height/2) < 0 {
		p.Y = float32(1 + p.Height/2)
	} else if p.Y+float32(p.Height/2) > float32(h) {
		p.Y = float32(h - p.Height/2 - 1)
	}
}

func (p *Paddle) PressDown(screen *ebiten.Image) {
	_, h := screen.Size()
	p.Y += p.Speed
	if p.Y-float32(p.Height/2) < 0 {
		p.Y = float32(1 + p.Height/2)
	} else if p.Y+float32(p.Height/2) > float32(h) {
		p.Y = float32(h - p.Height/2 - 1)
	}
}

func (p *Paddle) AiUpdate(b *Ball) {
	// unbeatable haha
	p.Y = b.Y
}

func (p *Paddle) Draw(screen *ebiten.Image, scoreFont font.Face, ai bool) {
	// draw player's paddle
	pOpts := &ebiten.DrawImageOptions{}
	pOpts.GeoM.Translate(float64(p.X), float64(p.Y-float32(p.Height/2)))
	p.Img.Fill(p.Color)
	screen.DrawImage(p.Img, pOpts)

	// draw player's score if needed
	if !ai {
		if p.scorePrinted.score != p.Score && p.scorePrinted.printed {
			p.scorePrinted.printed = false
		}
		if p.scorePrinted.score == 0 && !p.scorePrinted.printed {
			p.scorePrinted.x = int(p.X + (GetCenter(screen).X-p.X)/2)
			p.scorePrinted.y = int(2 * 30)
		}
		if (p.scorePrinted.score == 0 || p.scorePrinted.score != p.Score) && !p.scorePrinted.printed {
			p.scorePrinted.score = p.Score
			p.scorePrinted.printed = true
		}
		s := strconv.Itoa(p.scorePrinted.score)
		text.Draw(screen, s, scoreFont, p.scorePrinted.x, p.scorePrinted.y, p.Color)
	}
}
