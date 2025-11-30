package main

import (
	"fmt"
	"github.com/dstoiko/go-pong-wasm/pong"
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"github.com/hajimehoshi/ebiten/inpututil"
	"image/color"
	"math"
	"math/rand"
	"runtime"
)

// Game is the structure of the game state
type Game struct {
	state    pong.GameState
	aiMode   bool
	ball     *pong.Ball
	player1  *pong.Paddle
	player2  *pong.Paddle
	rally    int
	level    int
	maxScore int
	Network  pong.Network
	Net      int
	Fire     int
	rng      *rand.Rand
}

const (
	initBallVelocity = 5.0
	initPaddleSpeed  = 10.0
	speedUpdateCount = 6
	speedIncrement   = 0.5
)

const (
	windowWidth  = 800
	windowHeight = 600
	Size         = 32
)

// NewGame creates an initializes a new game
func NewGame(aiMode bool) *Game {
	g := &Game{}
	g.init(aiMode)
	g.Network = pong.NewNetwork(4, Size, 8)
	g.rng = rand.New(rand.NewSource(1))
	return g
}

func (g *Game) init(aiMode bool) {
	g.state = pong.StartState
	g.aiMode = aiMode
	if aiMode {
		g.maxScore = 100
	} else {
		g.maxScore = 11
	}

	rng := rand.New(rand.NewSource(1))
	up := pong.NewMatrix(4, 1, make([]float64, 4)...)
	down := pong.NewMatrix(4, 1, make([]float64, 4)...)
	for i := range up.Data {
		up.Data[i] = rng.Float64()
	}
	for i := range down.Data {
		down.Data[i] = rng.Float64()
	}
	g.player1 = &pong.Paddle{
		Position: pong.Position{
			X: pong.InitPaddleShift,
			Y: float32(windowHeight / 2)},
		Score:  0,
		Speed:  initPaddleSpeed,
		Width:  pong.InitPaddleWidth,
		Height: pong.InitPaddleHeight,
		Color:  pong.ObjColor,
		Up:     ebiten.KeyW,
		Down:   ebiten.KeyS,
		UpV:    up,
		DownV:  down,
	}
	g.player2 = &pong.Paddle{
		Position: pong.Position{
			X: windowWidth - pong.InitPaddleShift - pong.InitPaddleWidth,
			Y: float32(windowHeight / 2)},
		Score:  0,
		Speed:  initPaddleSpeed,
		Width:  pong.InitPaddleWidth,
		Height: pong.InitPaddleHeight,
		Color:  pong.ObjColor,
		Up:     ebiten.KeyO,
		Down:   ebiten.KeyK,
	}
	g.ball = &pong.Ball{
		Position: pong.Position{
			X: float32(windowWidth / 2),
			Y: float32(windowHeight / 2)},
		Radius:    pong.InitBallRadius,
		Color:     pong.ObjColor,
		XVelocity: initBallVelocity,
		YVelocity: initBallVelocity,
	}
	g.level = 0
	g.ball.Img, _ = ebiten.NewImage(int(g.ball.Radius*2), int(g.ball.Radius*2), ebiten.FilterDefault)
	g.player1.Img, _ = ebiten.NewImage(g.player1.Width, g.player1.Height, ebiten.FilterDefault)
	g.player2.Img, _ = ebiten.NewImage(g.player2.Width, g.player2.Height, ebiten.FilterDefault)

	pong.InitFonts()
}

func (g *Game) reset(screen *ebiten.Image, state pong.GameState) {
	w, _ := screen.Size()
	g.state = state
	g.rally = 0
	g.level = 0
	if state == pong.StartState {
		g.player1.Score = 0
		g.player2.Score = 0
	}
	g.player1.Position = pong.Position{
		X: pong.InitPaddleShift, Y: pong.GetCenter(screen).Y}
	g.player2.Position = pong.Position{
		X: float32(w - pong.InitPaddleShift - pong.InitPaddleWidth), Y: pong.GetCenter(screen).Y}
	g.ball.Position = pong.GetCenter(screen)
	g.ball.XVelocity = initBallVelocity
	g.ball.YVelocity = initBallVelocity
}

// Update updates the game state
func (g *Game) Update(screen *ebiten.Image) error {
	switch g.state {
	case pong.StartState:
		if inpututil.IsKeyJustPressed(ebiten.KeyC) {
			g.state = pong.ControlsState
		} else if inpututil.IsKeyJustPressed(ebiten.KeyA) {
			g.aiMode = true
			g.state = pong.PlayState
		} else if inpututil.IsKeyJustPressed(ebiten.KeyV) {
			g.aiMode = false
			g.state = pong.PlayState
		}

	case pong.ControlsState:
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			g.state = pong.StartState
		}
	case pong.PlayState:
		w, _ := screen.Size()

		if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			g.state = pong.PauseState
			break
		}

		g.player1.Update(screen)
		if g.aiMode {
			g.player2.AiUpdate(g.ball)
		} else {
			g.player2.Update(screen)
		}

		xV := g.ball.XVelocity
		g.ball.Update(g.player1, g.player2, screen)
		// rally count
		if xV*g.ball.XVelocity < 0 {
			// score up when ball touches human player's paddle
			if g.aiMode && g.ball.X < float32(w/2) {
				g.player1.Score++
			}

			g.rally++

			// spice things up
			if (g.rally)%speedUpdateCount == 0 {
				g.level++
				g.ball.XVelocity += speedIncrement
				g.ball.YVelocity += speedIncrement
				g.player1.Speed += speedIncrement
				g.player2.Speed += speedIncrement
			}
		}

		if g.ball.X < 0 {
			g.player2.Score++
			if g.aiMode {
				g.state = pong.GameOverState
				break
			}
			g.reset(screen, pong.InterState)
		} else if g.ball.X > float32(w) {
			g.player1.Score++
			if g.aiMode {
				g.state = pong.GameOverState
				break
			}
			g.reset(screen, pong.InterState)
		}

		if g.player1.Score == g.maxScore || g.player2.Score == g.maxScore {
			g.state = pong.GameOverState
		}

	case pong.InterState, pong.PauseState:
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			g.state = pong.PlayState
		} else if inpututil.IsKeyJustPressed(ebiten.KeyR) {
			g.reset(screen, pong.StartState)
		}

	case pong.GameOverState:
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			g.reset(screen, pong.StartState)
		}
	}

	g.Draw(screen)
	if g.Fire == 0 {
		width := g.Network.Width
		rng := rand.New(rand.NewSource(1))
		for i := range Size {
			sum := 0.0
			for h := range windowHeight {
				for w := range windowWidth {
					pixel := screen.At(w, h)
					grayPixel := color.GrayModel.Convert(pixel).(color.Gray)
					x := rng.Intn(6)
					if x == 0 {
						sum += float64(grayPixel.Y)
					} else if x == 1 {
						sum -= float64(grayPixel.Y)
					}
				}
			}
			g.Network.Neurons[g.Net].Vector[width+i] = sum
		}
		{
			sum := 0.0
			for i := range Size {
				sum += math.Abs(g.Network.Neurons[g.Net].Vector[width+i])
			}
			for i := range Size {
				g.Network.Neurons[g.Net].Vector[width+i] /= sum
			}
		}
		g.Net = (g.Net + 1) % 6
		g.Network.Iterate()
		/*up := pong.NCS(g.Network.Neurons[6].Vector[:width], g.player1.UpV.Data)
		down := pong.NCS(g.Network.Neurons[7].Vector[:width], g.player1.DownV.Data)
		if up > down {
			g.player1.PressUp(screen)
		} else {
			g.player1.PressDown(screen)
		}*/
		vectors := make([]*pong.Vector[pong.Neuron], 8)
		for ii := range 6 {
			vector := pong.Vector[pong.Neuron]{}
			vector.Meta = g.Network.Neurons[ii]
			vector.Vector = g.Network.Neurons[ii].Vector[:width]
			vectors[ii] = &vector

		}
		{
			a := pong.Vector[pong.Neuron]{}
			a.Meta = g.Network.Neurons[6]
			a.Vector = g.Network.Neurons[6].Vector[:width]
			vectors[6] = &a
		}
		{
			a := pong.Vector[pong.Neuron]{}
			a.Meta = g.Network.Neurons[7]
			a.Vector = g.Network.Neurons[7].Vector[:width]
			vectors[7] = &a
		}
		config := pong.Config{
			Iterations: 16,
			Size:       width,
			Divider:    1,
		}
		pong.MorpheusFast(rng.Int63(), config, vectors)
		sum := 0.0
		sub := 0.0
		for i := range vectors {
			sum += vectors[i].Stddev
			if i&1 == 0 {
				sub += vectors[i].Stddev
			}
		}
		fmt.Println(sub, sum, sub/sum)
		if g.rng.Float64() > sub/sum {
			g.player1.PressUp(screen)
			fmt.Println("up")
		} else {
			g.player1.PressDown(screen)
			fmt.Println("down")
		}
	}
	g.Fire = (g.Fire + 1) % 1
	return nil
}

// Draw updates the game screen elements drawn
func (g *Game) Draw(screen *ebiten.Image) error {
	screen.Fill(pong.BgColor)

	pong.DrawCaption(g.state, pong.ObjColor, screen)
	pong.DrawBigText(g.state, pong.ObjColor, screen)

	if g.state != pong.ControlsState {
		g.player1.Draw(screen, pong.ArcadeFont, false)
		g.player2.Draw(screen, pong.ArcadeFont, g.aiMode)
		g.ball.Draw(screen)
	}

	ebitenutil.DebugPrint(screen, fmt.Sprintf("TPS: %0.2f", ebiten.CurrentTPS()))

	return nil
}

// Layout sets the screen layout
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return windowWidth, windowHeight
}

func main() {
	// On browsers, let's use fullscreen so that this is playable on any browsers.
	// It is planned to ignore the given 'scale' apply fullscreen automatically on browsers (#571).
	if runtime.GOARCH == "js" || runtime.GOOS == "js" {
		ebiten.SetFullscreen(true)
	}
	ai := true
	g := NewGame(ai)
	if err := ebiten.RunGame(g); err != nil {
		panic(err)
	}
}
