package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
)

type pipes struct {
	mu sync.RWMutex

	texture *sdl.Texture
	r       *sdl.Renderer
	speed   int32

	pipes []*pipe
}

func newPipes(r *sdl.Renderer) (*pipes, error) {
	var err error
	var texture *sdl.Texture

	sdl.Do(func() {
		texture, err = img.LoadTexture(r, "resources/imgs/pipe.png")
	})
	if err != nil {
		return nil, fmt.Errorf("could not load pipe image: %v", err)
	}

	ps := &pipes{
		texture: texture,
		speed:   2,
		r:       r,
	}

	go func() {
		for {
			ps.mu.Lock()
			ps.pipes = append(ps.pipes, newPipe())
			ps.mu.Unlock()
			time.Sleep(1 * time.Second)
		}
	}()

	return ps, nil
}

func (ps *pipes) paint() error {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	var err error

	for _, p := range ps.pipes {
		if err = p.paint(ps.r, ps.texture); err != nil {
			return err
		}
	}
	return nil
}

func (ps *pipes) restart() {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.pipes = nil
}

func (ps *pipes) update(sc *score) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	var rem []*pipe
	for _, p := range ps.pipes {
		p.mu.Lock()
		p.x -= ps.speed
		p.mu.Unlock()
		if p.x+p.w > 0 {
			rem = append(rem, p)
		} else {
			sc.increase()
		}
	}
	ps.pipes = rem
}

func (ps *pipes) destroy() {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	sdl.Do(func() {
		ps.texture.Destroy()
	})
}

type pipe struct {
	mu sync.RWMutex

	x        int32
	h        int32
	w        int32
	inverted bool
}

func newPipe() *pipe {

	return &pipe{
		x:        800,
		h:        100 + int32(rand.Intn(300)),
		w:        50,
		inverted: rand.Float32() > 0.5,
	}

}

func (p *pipe) paint(r *sdl.Renderer, texture *sdl.Texture) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var err error

	rect := &sdl.Rect{X: p.x, Y: 600 - p.h, W: p.w, H: p.h}
	flip := sdl.FLIP_NONE
	if p.inverted {
		rect.Y = 0
		flip = sdl.FLIP_VERTICAL
	}

	sdl.Do(func() {
		err = r.CopyEx(texture, nil, rect, 0, nil, flip)
	})
	if err != nil {
		return fmt.Errorf("could not copy pipe: %v", err)
	}
	return nil
}
