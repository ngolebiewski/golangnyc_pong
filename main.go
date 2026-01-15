package main

import (
	"image/color"
	"log"
	"math/rand/v2"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	sW     = 320 // screen width
	sH     = 240 // screen height
	pW     = 5   // paddle width
	pH     = 40  // paddle height
	pSpeed = 4   // paddle speed
	bW     = 4   // ball width
	bMaxV  = 10  // ball's max velocity
)

type Paddle struct {
	x, y, w, h float32
	score      int
}

type Ball struct {
	x, y, w, h float32 // FYI all float32 becuase that's what GPUs want, etc.
	vx         float32 // velocity x
	vy         float32 // velocity y
	v          float32 // velocity (kind of used as an overall speed)
	maxV       float32 // max velocity so the ball doesn't get INCREDIBLY fast
	isInPlay   bool    // is the ball in play or in a initial/not moving state
}

type Game struct {
	p1 Paddle
	p2 Paddle
	b  Ball
}

func NewGame() *Game {
	return &Game{
		p1: Paddle{
			x:     5,
			y:     sH/2 - pH/2, //center the paddle on the y axis
			w:     pW,
			h:     pH,
			score: 0,
		},
		p2: Paddle{
			x:     sW - 5 - pW, // account for offset from right side of screen and paddle width
			y:     sH/2 - pH/2, //center the paddle on the y axis
			w:     pW,
			h:     pH,
			score: 0,
		},
		b: Ball{
			x:    sW/2 - bW/2,
			y:    sH/2 - bW/2, //center the ball on the y axis drawing a square so height of ball is same as width
			w:    bW,
			h:    bW,
			vx:   0,
			vy:   0,
			v:    1,
			maxV: bMaxV,
		},
	}
}

func (g *Game) Reset() {
	// center paddles and ball and resets ball velocity, keep the scores though
	g.p1.y = sH/2 - pH/2
	g.p2.y = sH/2 - pH/2
	g.b.x = sW/2 - bW/2
	g.b.y = sH/2 - bW/2
	g.b.vy = 0
	g.b.vx = 0
	g.b.v = 1
	g.b.isInPlay = false
}

func (g *Game) Update() error {
	// g.p1.y += 1 //demos that you can update the y every tic, 60 tics per secong (tps)
	//lets add in player1's inputs! Game Loop is input -> update -> draw, REPEAT

	// ---- PLAYER INPUTS ---- //
	// Note can accept game pads controls here to, mouse clicks, touch events.
	if ebiten.IsKeyPressed(ebiten.KeyW) {
		//move paddle 1 up
		// g.p1.y -= 4
		g.p1.y = max(0, g.p1.y-pSpeed)

	}
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		//move paddle 1 down
		g.p1.y = min(sH-g.p1.h, g.p1.y+pSpeed)
	}

	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		//move paddle 2 up
		// g.p1.y -= 4
		g.p2.y = max(0, g.p2.y-pSpeed)

	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		//move paddle 2 down
		g.p2.y = min(sH-g.p2.h, g.p2.y+pSpeed)
	}
	//only serve the ball if it is not in play
	if ebiten.IsKeyPressed(ebiten.KeySpace) && g.b.isInPlay == false {
		g.b.isInPlay = true
		g.b.ServeBall() // serves the ball!
	}

	// ----- Game Controls ------ //
	// Start game over and reset score to 0 vs 0
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		*g = *NewGame()
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyF) {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	}

	// --- Ball Logic + Collision checks --- //
	//check for top || bottom screen colissions and bounce off wall
	if g.b.y <= 0 || g.b.y >= sH-g.b.w {
		g.b.vy = -g.b.vy // flip Y velocity
	}
	//check for out of bounds
	//check left -> if off screen player 2 scores
	if g.b.x <= 0 {
		g.p2.score += 1
		g.Reset()
	}
	//check right
	if g.b.x >= sW {
		g.p1.score += 1
		g.Reset()
	}

	//bounce off paddles with AABB collision checking
	//Ball hit by Paddle 1
	if aabb(g.b.x, g.b.y, g.b.w, g.b.w, g.p1.x, g.p1.y, g.p1.w, g.p1.h) && g.b.vx < 0 {
		g.b.vx = -g.b.vx
		g.b.vy += rand.Float32() / 3 * RandomDirection()
		g.b.v = min(g.b.maxV, g.b.v+.5)
	}

	// Ball hit by Paddle 2
	if aabb(g.b.x, g.b.y, g.b.w, g.b.w, g.p2.x, g.p2.y, g.p2.w, g.p2.h) && g.b.vx > 0 {
		g.b.vx = -g.b.vx
		g.b.vy += rand.Float32() / 3 * RandomDirection()
		g.b.v = min(g.b.maxV, g.b.v+.5)
	}
	// update ball loction now that velocity is set
	g.b.x += g.b.vx * g.b.v
	g.b.y += g.b.vy * g.b.v

	// Ball update

	return nil
}

// a classic Axis-Aligned Bounding-Box Collission check
func aabb(ax, ay, aw, ah, bx, by, bw, bh float32) bool {
	return ax < bx+bw &&
		ax+aw > bx &&
		ay < by+bh &&
		ay+ah > by
}

func (p *Paddle) DrawPaddle(screen *ebiten.Image) {
	vector.FillRect(screen, p.x, p.y, p.w, p.h, color.White, false)
}

func (b *Ball) DrawBall(screen *ebiten.Image) {
	vector.FillRect(screen, b.x, b.y, b.w, b.h, color.White, false)
}

// Random direction generator -1 for left/up and 1 for right/down depending on the X/Y axis we're on
func RandomDirection() float32 {
	if rand.Float32() < .5 {
		return -1
	} else {
		return 1
	}
}

func (b *Ball) ServeBall() {
	//apply initial velocity to ball X is either left or right
	b.vx = RandomDirection()
	b.vy = rand.Float32() * 3 * RandomDirection()
}

func (g *Game) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrintAt(screen, "GoLangNYC Pong!!", sW/2-40, 10) //basically center
	//draw a rectangle
	// vector.FillRect(screen, 5, 10, 5, 40, color.RGBA{100, 220, 0, 255}, false) //Green!
	// vector.FillRect(screen, g.p1.x, g.p1.y, g.p1.w, g.p1.h, color.White, false) // we made this into a function
	g.p1.DrawPaddle(screen)
	g.p2.DrawPaddle(screen)
	g.b.DrawBall(screen)

	// Show player scores
	ebitenutil.DebugPrintAt(screen, strconv.Itoa(g.p1.score), 40, 10)
	ebitenutil.DebugPrintAt(screen, strconv.Itoa(g.p2.score), sW-40, 10)

}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return sW, sH
}

func main() {
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("GoLangNYC 1-25-26 Pong!")

	game := NewGame()
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
