package bopher

import "fmt"

type Config struct {
	Mattermost   Mattermost `json:"mattermost"`
	Gopher       string     `json:"gopher"`
	MaxOfGophers int        `json:"max_of_gophers"`
}

type Mattermost struct {
	Host    string  `json:"host"`
	Port    string  `json:"port"`
	Bot     Bot     `json:"bot"`
	Team    string  `json:"team"`
	Channel Channel `json:"channel"`
}

type Bot struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	First    string `json:"first_name"`
	Last     string `json:"last_name"`
}

type Channel struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Purpose     string `json:"purpose"`
}

func (m *Mattermost) HttpURL() string {
	if len(m.Port) == 0 {
		return fmt.Sprintf("http://%s", m.Host)
	} else {
		return fmt.Sprintf("http://%s:%s", m.Host, m.Port)
	}
}

func (m *Mattermost) WsURL() string {
	if len(m.Port) == 0 {
		return fmt.Sprintf("ws://%s", m.Host)
	} else {
		return fmt.Sprintf("ws://%s:%s", m.Host, m.Port)
	}
}
