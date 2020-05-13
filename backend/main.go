package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	graphqlhandler "github.com/99designs/gqlgen/graphql/handler"
	"github.com/cockroachdb/errors"
	"github.com/gorcon/rcon"
	"github.com/joho/godotenv"

	"go.stevenxie.me/zoomcraft/backend/graphql"
	"go.stevenxie.me/zoomcraft/backend/graphql/graphqlutil"
	"go.stevenxie.me/zoomcraft/backend/minecraft"
	"go.stevenxie.me/zoomcraft/backend/util/logutil"
)

func main() {
	if err := func() error {
		godotenv.Load()

		// Create logger.
		logger := logutil.NewLogger(log.NewSyncWriter(os.Stdout))
		logger = logutil.WithComponent(logger, "backend")
		logger = level.NewInjector(logger, level.DebugValue())

		// Set log level.
		if !isTruthy(os.Getenv("BACKEND_DEBUG")) {
			logger = level.NewFilter(logger, level.AllowInfo())
		}

		// Create Minecraft client.
		var client *minecraft.Client
		if err := func() (err error) {
			var (
				addr = getEnv("RCON_ADDRESS", "localhost:25575")
				pass = getEnv("RCON_PASSWORD", "minecraft")
			)
			conn, err := rcon.Dial(addr, pass)
			if err != nil {
				return errors.Wrap(err, "dial server")
			}
			client = minecraft.NewClient(conn)
			return nil
		}(); err != nil {
			return errors.Wrap(err, "connect with RCON")
		}

		// Create services.
		var players minecraft.PlayerService
		if err := func() (err error) {
			logger := logutil.WithComponent(logger, "player_service")
			players = minecraft.NewPlayerService(client, logger)

			// Apply a cache layer to limit requests.
			cache := minecraft.PlayerServiceCache{MaxAge: 100 * time.Millisecond}
			players = cache.Apply(players)
			return nil
		}(); err != nil {
			return errors.Wrap(err, "create player service")
		}

		// Create executable schema.
		schema := graphql.NewExecutableSchema(graphql.Config{
			Resolvers: &graphql.Resolver{
				Players: players,
			},
		})

		// Create and configure handler.
		handler := graphqlhandler.NewDefaultServer(schema)
		handler.SetErrorPresenter(graphqlutil.PresentError)

		// Register HTTP routes.
		mux := http.NewServeMux()
		mux.Handle("/graphql", handler)
		mux.Handle("/graphiql", graphqlutil.ServeGraphiQL("./graphql"))

		// Create and run server.
		port := getEnv("BACKEND_PORT", "9090")
		server := &http.Server{
			Addr:    fmt.Sprintf(":%s", port),
			Handler: mux,
		}
		{
			l := log.With(logger, "port", port)
			l = level.Info(l)
			logutil.Log(l, "listening for connections")
		}
		return server.ListenAndServe()
	}(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %+v\n", err)
		os.Exit(1)
	}
}

func getEnv(key string, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

func isTruthy(v string) bool {
	v = strings.TrimSpace((v))
	switch v {
	case "true", "t", "1":
		return true
	default:
		return false
	}
}
