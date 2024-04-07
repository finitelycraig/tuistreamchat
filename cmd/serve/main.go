package main

import (
	"github.com/finitelycraig/tuistreamchat/internal"
	"github.com/rs/zerolog"
	"os"
)

func main() {
	file, err := os.OpenFile(
		"ssh-dev.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0664,
	)
	if err != nil {
		panic(err)
	}

	logger := zerolog.New(file).With().Timestamp().Logger()

	defer file.Close()
	internal.RunOverSSH(logger)
}
