package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/gorcon/rcon"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"go.stevenxie.me/covidcraft/rtcsig"
	"go.stevenxie.me/covidcraft/util/logutil"
	melody "gopkg.in/olahol/melody.v1"
)

const addr = ":8080"

func main() {
	if err := func() error {
		godotenv.Load()

		// Create logger.
		logger := log.NewJSONLogger(log.NewSyncWriter(os.Stdout))
		logger = level.NewInjector(logger, level.DebugValue())

		// Create game connection.
		var (
			gameAddr = "localhost:25575"
			gamePass = "minecraft"
		)
		if val := os.Getenv("RCON_ADDRESS"); val != "" {
			gameAddr = val
		}
		if val := os.Getenv("RCON_PASSWORD"); val != "" {
			gamePass = val
		}
		game, err := rcon.Dial(gameAddr, gamePass)
		if err != nil {
			return errors.Wrap(err, "dial")
		}

		// Create interchange.
		m := melody.New()
		interchange := rtcsig.NewInterchange(
			m,
			game,
			logutil.WithComponent(logger, "interchange"),
		)

		// Create handler.
		mux := http.NewServeMux()
		mux.Handle("/ws", interchange)

		// Create and run server.
		server := &http.Server{
			Addr:    addr,
			Handler: mux,
		}

		{
			l := log.With(logger, "addr", addr)
			l = level.Info(l)
			logutil.Log(l, "Listening for connections.")
		}
		return server.ListenAndServe()
	}(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %+v\n", err)
		os.Exit(1)
	}
}
