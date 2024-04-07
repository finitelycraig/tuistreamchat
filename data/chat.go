package data

import (
	"fmt"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	twitch "github.com/gempir/go-twitch-irc/v4"
	"github.com/rs/zerolog"
	"golang.org/x/term"
	"os"
	"strings"
)

type (
	errMsg error
)

// chat implements bubbletea
type Chat struct {
	channel              *Channel
	sync                 chan tea.Msg
	termHeight           int
	termWidth            int
	viewport             viewport.Model
	messages             []string
	messagesWidthWrapped []string
	textarea             textarea.Model
	err                  error
	anonymous            bool
	logger               zerolog.Logger
}

func NewChat(s chan tea.Msg, logger zerolog.Logger) *Chat {
	width, height, _ := term.GetSize(int(os.Stdout.Fd()))
	ta := textarea.New()
	ta.Placeholder = "Send a message... ctrl+c or Esc to exit"
	ta.Focus()
	ta.Prompt = "┃ "
	ta.CharLimit = 500
	ta.SetWidth(width - 5)
	ta.SetHeight(1)
	//remove cursor line styling
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.FocusedStyle.Prompt = lipgloss.NewStyle().Foreground(lipgloss.Color("#6441a5"))
	ta.ShowLineNumbers = false
	ta.KeyMap.InsertNewline.SetEnabled(false)

	vp := viewport.New(width-5, height-6)
	vp.Style = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#6441a5")).PaddingRight(0)
	//logger.Info().Msg(fmt.Sprintf("vp height: %d", vp.Height))
	//logger.Info().Msg(fmt.Sprintf("vp width: %d", vp.Width))

	chat := &Chat{
		channel:              nil,
		sync:                 s,
		termHeight:           height,
		termWidth:            width,
		viewport:             vp,
		messages:             []string{},
		messagesWidthWrapped: []string{},
		textarea:             ta,
		err:                  nil,
		anonymous:            false,
		logger:               logger,
	}
	return chat
}

func (c *Chat) SetUpChannel(user, oauth string) {
	var client *twitch.Client
	if oauth == "" {
		client = twitch.NewAnonymousClient()
	} else {
		client = twitch.NewClient(user, oauth)
	}
	channel := NewChannel(user, client, c.sync, c.logger)
	c.channel = &channel
	c.channel.AddHost(user)
	c.channel.Listen(user)
	colourBanner := ""
	for _, colour := range Colours {
		style := lipgloss.NewStyle().Foreground(lipgloss.Color(colour))
		colourBanner += style.Render("=")
	}
	c.viewport.SetContent(fmt.Sprintf("Welcome to %s's chat\n%s", c.channel.Name, colourBanner))
}

func NewAnonymousChat(s chan tea.Msg, logger zerolog.Logger) *Chat {
	width, height, _ := term.GetSize(int(os.Stdout.Fd()))

	ta := textarea.New()
	ta.Placeholder = "Send a message..."
	ta.Focus()
	ta.Prompt = "┃ "
	ta.CharLimit = 500
	ta.SetWidth(width - 5)
	ta.SetHeight(1)
	//remove cursor line styling
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.FocusedStyle.Prompt = lipgloss.NewStyle().Foreground(lipgloss.Color("#6441a5"))
	ta.ShowLineNumbers = false
	ta.KeyMap.InsertNewline.SetEnabled(false)

	vp := viewport.New(width-4, height-4)
	vp.Style = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#6441a5")).PaddingRight(2)
	//logger.Info().Msg(fmt.Sprintf("vp height: %d", vp.Height))
	//logger.Info().Msg(fmt.Sprintf("vp width: %d", vp.Width))

	chat := &Chat{
		channel:              nil,
		sync:                 s,
		viewport:             vp,
		messages:             []string{},
		messagesWidthWrapped: []string{},
		textarea:             ta,
		err:                  nil,
		anonymous:            true,
		logger:               logger,
	}
	return chat
}

func (c *Chat) Start() {
	fmt.Println("starting a chat")
	for {
		msg := <-c.sync
		fmt.Println(msg)
	}
}

func (c *Chat) Init() tea.Cmd {
	go c.Listen()
	return textarea.Blink
}

func (c *Chat) Listen() {
	for {
		msg := <-c.sync
		c.Update(msg)
	}
}

func (c *Chat) linesUsed() int {
	formattingOffset := 25
	lines := 0
	for _, message := range c.messages {
		c.logger.Info().Msg(message)
		if len(message)-formattingOffset < c.viewport.Width-1 {
			lines += 1
		} else {
			c.logger.Info().Msg(fmt.Sprintf("len(message)=%d | len(message) / c.viewport.Width)=%d", len(message), len(message)/c.viewport.Width))
			lines += (len(message) / c.viewport.Width) + 1
		}
	}
	return lines
}

func (c *Chat) resizeMessages() {
	c.messagesWidthWrapped = []string{}
	for _, m := range c.messages {
		c.concatMessage(m)
	}
}

func (c *Chat) wrapAfterResize() {
	c.messagesWidthWrapped = []string{}
	for _, msg := range c.messages {
		c.messagesWidthWrapped = append(c.messagesWidthWrapped, strings.Split(c.wordWrap(msg), "\n")...)
	}
}

func (c *Chat) wordWrap(msg string) string {
	//split each message into individual lines if they're longer than the width
	words := strings.Fields(msg)
	if len(words) == 0 {
		return ""
	}
	wrapped := words[0]
	lineWidth := c.viewport.Width - 2 // for the sides
	spaceLeft := lineWidth - len(wrapped) + 15
	for _, word := range words[1:] {
		if len(word)+1 > spaceLeft {
			if len(word) > lineWidth {
				wordLength := 0
				for i, c := range word {
					if i%(lineWidth) == 0 {
						wrapped += "\n"
						wordLength = 0
					}
					wrapped += string(c)
					wordLength++
				}
				spaceLeft = lineWidth - wordLength
			} else {
				wrapped += "\n" + word
				spaceLeft = lineWidth - len(word)
			}
		} else {
			wrapped += " " + word
			spaceLeft -= 1 + len(word)
		}
	}
	return wrapped
}

func (c *Chat) concatMessage(msg string) {
	c.messages = append(c.messages, msg)
	c.messagesWidthWrapped = append(c.messagesWidthWrapped, strings.Split(c.wordWrap(msg), "\n")...)
	//c.logger.Info().Msg(fmt.Sprintf("displaying %d lines of messages", len(c.messagesWidthWrapped)))
	//for _, msg := range c.messagesWidthWrapped {
	//	c.logger.Info().Msg(fmt.Sprintf("%s", msg))
	//}
}

func (c *Chat) formattedMessages() string {
	allMessages := strings.Join(c.messagesWidthWrapped, "\n")
	return allMessages
}

func (c *Chat) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	//c.logger.Info().Msg(fmt.Sprintf("Update with msg %s", msg))
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	c.textarea, tiCmd = c.textarea.Update(msg)
	c.viewport, vpCmd = c.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if c.anonymous {
			c.viewport.Width = msg.Width - 4
			c.viewport.Height = msg.Height - 4
			c.wrapAfterResize()
			c.viewport.SetContent(c.formattedMessages())
			c.viewport.GotoBottom()
		}
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return c, tea.Quit
		case tea.KeyF5:
			c.messages = []string{}
			c.viewport.SetContent(c.formattedMessages())
			c.viewport.GotoBottom()
		case tea.KeyBackspace:
			if c.channel != nil {
				if len(c.messages) > 0 {
					c.messages = c.messages[:len(c.messages)-1]
					c.wrapAfterResize()
					c.viewport.SetContent(c.formattedMessages())
					c.viewport.GotoBottom()
				}
			}
		case tea.KeyEnter:
			if c.channel == nil {
				c.SetUpChannel(c.textarea.Value(), "")
				c.textarea.Reset()
			} else if !c.anonymous {
				message := c.textarea.Value()
				c.concatMessage(c.channel.Users[c.channel.Name].StylishName() + message)
				c.viewport.SetContent(c.formattedMessages())
				c.textarea.Reset()
				c.viewport.GotoBottom()
				c.channel.Say(message)
			}
		}
	case string:
		c.concatMessage(string(msg))
		c.viewport.SetContent(c.formattedMessages())
		c.viewport.GotoBottom()
	case errMsg:
		c.err = msg
		return c, nil
	}
	return c, tea.Batch(tiCmd, vpCmd)
}

func (c *Chat) View() string {
	if c.channel == nil {
		c.textarea.Placeholder = "Enter a channel name to join"
		return fmt.Sprintf(
			"%s",
			c.textarea.View()) + "\n\n"
	} else if !c.anonymous {
		return fmt.Sprintf(
			"%s\n\n%s",
			c.viewport.View(),
			c.textarea.View(),
		) + "\n\n"
	} else {
		return fmt.Sprintf(
			"%s",
			c.viewport.View(),
		) + "\n\n"
	}
}
