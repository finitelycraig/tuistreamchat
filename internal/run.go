package internal

import (
    "context"
    "fmt"
    "net"
	"os"
    "os/signal"
    "syscall"
    "time"

	tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/ssh"
    "github.com/charmbracelet/wish"
    "github.com/charmbracelet/wish/activeterm"
    "github.com/charmbracelet/wish/bubbletea"
	"github.com/finitelycraig/tuistreamchat/data"
	"github.com/rs/zerolog"
	"github.com/muesli/termenv"
)


func RunHosted(logger zerolog.Logger) {
	user := os.Getenv("TWITCHBOT")
    oauth := os.Getenv("TWITCHOAUTH")
	logger.Info().Msg(fmt.Sprintf("Starting a room with host %s", user))
	sync := make(chan tea.Msg)
    chat := data.NewChat(sync, logger)
    chat.SetUpChannel(user, oauth)
    p := tea.NewProgram(chat, tea.WithAltScreen())
    if _, err := p.Run(); err != nil {
       os.Exit(0) 
    }
}

func RunUnHosted(logger zerolog.Logger) {
	logger.Info().Msg(fmt.Sprintf("Starting a room without a host"))
	sync := make(chan tea.Msg)
    chat := data.NewAnonymousChat(sync, logger)
    p := tea.NewProgram(chat, tea.WithAltScreen())
    if _, err := p.Run(); err != nil {
       os.Exit(0) 
    }
}

const (
	host = "localhost"
	port = "2222"
)

type app struct {
	*ssh.Server
    logger zerolog.Logger
}

func newApp(log zerolog.Logger) *app {
	a := new(app)
	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),
		wish.WithHostKeyPath(".ssh/id_ed25519"),
		wish.WithMiddleware(
			bubbletea.MiddlewareWithProgramHandler(a.ProgramHandler, termenv.ANSI256),
			activeterm.Middleware(),
		),
	)
    a.logger = log
	if err != nil {
		log.Error().Err(err).Msg("Could not start new app")
	}

	a.Server = s
	return a
}

func (a *app) Start() {
	var err error
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	a.logger.Info().Msg(fmt.Sprintf("Starting SSH server host %s port %s", host, port))
	go func() {
		if err = a.ListenAndServe(); err != nil {
			a.logger.Error().Err(err).Msg("Could not start server in app.Start()")
			done <- nil
		}
	}()

	<-done
	a.logger.Info().Msg("Stopping SSH server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() { cancel() }()
	if err := a.Shutdown(ctx); err != nil {
		a.logger.Error().Msg("Could not stop server")
	}
}

func (a *app) ProgramHandler(s ssh.Session) *tea.Program {
	a.logger.Info().Msg(fmt.Sprintf("Starting a room without a host"))
	sync := make(chan tea.Msg)
    chat := data.NewAnonymousChat(sync, a.logger)
    programOptions := bubbletea.MakeOptions(s)
    programOptions = append(programOptions, tea.WithAltScreen())
	p := tea.NewProgram(chat, programOptions...)

	return p
}

func RunOverSSH(logger zerolog.Logger) {
	app := newApp(logger)
	app.Start()
}
