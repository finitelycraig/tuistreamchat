package data

import (
    "fmt"
	tea "github.com/charmbracelet/bubbletea"
	twitch "github.com/gempir/go-twitch-irc/v4"
    "github.com/rs/zerolog"
)

type NoteMsg string

type Channel struct  {
	Name			string
	Users			map[string]*User
    NumberOfUsers   int
	sync			chan tea.Msg
	twitchClient	*twitch.Client
    logger          zerolog.Logger
}

func NewChannel(name string, client *twitch.Client, s chan tea.Msg, logger zerolog.Logger) Channel {
	users := make(map[string]*User, 0)
	c := &Channel{
        Name:           name,
		Users: 			users,
		sync: 			s,
		twitchClient: 	client,
        logger:         logger,
	}
	return *c
}


func (c *Channel) AddHost(name string) {
	host := NewUser(name)
    host.MakeHost()
    c.Users[name] = &host
}

func (c *Channel) AddUser(name, style string, badges []string) {
    _, present := c.Users[name]
    if !present {
        //c.logger.Info().Msg(fmt.Sprintf("new user %s with style %s", name, style))
        user := NewUserWithStyle(name, style, badges)
        c.Users[name] = &user
    }
}

func (c *Channel) Listen(channel string) {
	c.twitchClient.OnPrivateMessage(func(message twitch.PrivateMessage) {
        badges := make([]string, 0, len(message.User.Badges))
        for k := range(message.User.Badges) {
            //c.logger.Info().Msg(k)
            badges = append(badges, k)
        }
        c.AddUser(message.User.Name, message.User.Color, badges)
        c.SendMessage(c.Users[message.User.Name].Say(message.Message))
	})
    c.logger.Info().Msg(fmt.Sprintf("Joining %s's channel", channel))
	c.twitchClient.Join(channel)
	go func() {
		err := c.twitchClient.Connect()
		if err != nil {
			panic(err)
		}
	}()
}

func (c *Channel) SendMessage(msg tea.Msg) {
    c.sync <- msg 
}

func (c *Channel) Say(message string) {
    c.twitchClient.Say(c.Name, message)
}

func (c *Channel) GetNumberOfUsers() int {
    users,_ := c.twitchClient.Userlist(c.Name)
    return len(users)
}
