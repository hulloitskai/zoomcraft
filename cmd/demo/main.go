package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cockroachdb/errors"

	"github.com/gorcon/rcon"
	"github.com/joho/godotenv"
)

func main() {
	if err := func() error {
		godotenv.Load()

		var (
			addr = os.Getenv("RCON_ADDRESS")
			pass = os.Getenv("RCON_PASSWORD")
		)
		conn, err := rcon.Dial(addr, pass)
		if err != nil {
			return errors.Wrap(err, "dial")
		}

		listOutput, err := conn.Execute("list")
		if err != nil {
			return errors.Wrap(err, "execute command")
		}
		listOutput = listOutput[strings.LastIndexByte(listOutput, ':')+2:]
		players := strings.Split(listOutput, ", ")

		t := time.NewTicker(300 * time.Millisecond)
		for range t.C {
			for _, player := range players {
				res, err := conn.Execute(fmt.Sprintf("data get entity %s Pos", player))
				if err != nil {
					return errors.Wrap(err, "execute command")
				}
				fmt.Printf("Output: %s\n", res)
			}
		}

		return nil
	}(); err != nil {
		panic(err)
	}
}
