package data

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"math/rand"
)

// var Colours = [8]string{"#c6a0f6","#f5bde6","#ed8796","#a6da95","#eed49f","#8aadf4","#f5a97f","#91d7e3"}
var Colours = [12]string{"#a6cee3", "#1f78b4", "#b2df8a", "#33a02c", "#fb9a99", "#e31a1c", "#fdbf6f", "#ff7f00", "#cab2d6", "#6a3d9a", "#ffff99", "#b15928"}

var modBadge = "🟩"
var vipBadge = "🟪"
var broadcasterBadge = "🟥"

type User struct {
	Name   string
	Style  lipgloss.Style
	Badges string
	IsHost bool
}

func randomColour() string {
	return Colours[rand.Intn(len(Colours))]
}

func randomStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(randomColour()))
}

func styleFromHex(hex string) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(hex))
}

func NewUser(n string) User {
	return User{Name: n, Style: randomStyle(), IsHost: false}
}

func NewUserWithStyle(n, s string, badges []string) User {
	if s == "" {
		s = randomColour()
	}
	prefix := ""
	for _, badge := range badges {
		switch badge {
		case "moderator":
			prefix += modBadge
		case "vip":
			prefix += vipBadge
		}
	}
	if prefix != "" {
		prefix += " "
	}
	return User{Name: n, Style: styleFromHex(s), Badges: prefix, IsHost: false}
}

func (u *User) MakeHost() {
	u.IsHost = true
	u.Badges = broadcasterBadge + " "
}

func (u *User) Say(msg string) string {
	return fmt.Sprintf("%s %s", u.StylishName(), msg)
}

func (u *User) StylishName() string {
	//return u.Badges + u.Style.Render(u.Name + ": ")
	return u.Style.Render(u.Name + ":")
}
