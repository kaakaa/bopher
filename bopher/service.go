package bopher

import (
	"fmt"
	"math/rand"
	"os/exec"
	"time"

	"github.com/mattn/gopher"
)

type Service struct {
	Gopher       string
	MaxOfGophers int
}

func (s *Service) goGopher() {
	gophers := gopher.Lookup()

	if len(gophers) > s.MaxOfGophers {
		s.sayGopher(fmt.Sprintf("Here is so crowded! (Max: %d gophers)", s.MaxOfGophers))
		return
	}

	cmd := exec.Command(s.Gopher)
	cmd.Start()
}

func (s *Service) jumpGopher() {
	gophers := gopher.Lookup()
	for _, gopher := range gophers {
		gopher.Jump()
	}
}

func (s *Service) sayGopher(message string) {
	gophers := gopher.Lookup()
	gopher := gophers[rand.Int()%len(gophers)]
	for {
		if err := gopher.Message(message, ""); err == nil {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func (s *Service) byeGopher() {
	gophers := gopher.Lookup()
	for _, gopher := range gophers {
		gopher.Terminate()
	}
	return
}
