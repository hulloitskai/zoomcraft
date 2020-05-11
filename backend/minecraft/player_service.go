package minecraft

import (
	"context"
	stderrors "errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"github.com/cockroachdb/errors"

	"go.stevenxie.me/covidcraft/backend/util/logutil"
)

type playerService struct {
	client *Client
	logger log.Logger
}

// NewPlayerService creates a PlayerService.
func NewPlayerService(c *Client, logger log.Logger) PlayerService {
	return &playerService{
		client: c,
		logger: level.NewInjector(logger, level.DebugValue()),
	}
}

func (svc *playerService) List(ctx context.Context) (_ []*Player, err error) {
	defer func(start time.Time) {
		l := log.With(svc.logger, "took", time.Since(start))
		logutil.Trace(l, "List", err)
	}(time.Now())

	usernames, err := svc.listUsernames()
	if err != nil {
		return nil, errors.Wrap(err, "list usernames")
	}
	{
		l := log.With(svc.logger, "usernames", usernames)
		logutil.Log(l, "Discovered %d players.", len(usernames))
	}

	players := make([]*Player, 0, len(usernames))
	for _, u := range usernames {
		player, err := svc.Get(ctx, u)
		if err != nil {
			if err == ErrNotFound {
				continue
			}
			return nil, errors.Wrapf(err, "get position for '%s'", u)
		}
		players = append(players, player)
	}
	return players, nil
}

func (svc *playerService) Get(_ context.Context, username string) (_ *Player, err error) {
	logger := log.With(svc.logger, "username", username)
	defer func(start time.Time) {
		l := log.With(logger, "took", time.Since(start))
		logutil.Trace(l, "Get", err)
	}(time.Now())
	pos, err := svc.getPosition(username)
	if err != nil {
		return nil, errors.Wrap(err, "position")
	}
	rot, err := svc.getRotation(username)
	if err != nil {
		return nil, errors.Wrap(err, "rotation")
	}
	return &Player{
		Username:    username,
		Position:    *pos,
		Orientation: *rot,
	}, nil
}

func (svc *playerService) getPosition(username string) (*Coordinates, error) {
	cmd := fmt.Sprintf("data get entity %s Pos", username)
	out, err := svc.client.Execute(cmd)
	if err != nil {
		return nil, errors.Wrap(err, "execute command")
	}
	if out == "No entity was found" { // player disconnected
		return nil, ErrNotFound
	}

	// Parse output.
	out = out[strings.LastIndexByte(out, ':')+2:]
	out = strings.Trim(out, "[]")

	// Parse output parts into pos.
	var (
		pos   = new(Coordinates)
		parts = []*float64{&pos.X, &pos.Y, &pos.Z}
	)
	for i, part := range strings.Split(out, ", ") {
		f, err := strconv.ParseFloat(part[:len(part)-1], 64)
		if err != nil {
			return nil, errors.Wrap(err, "parse coordinate part")
		}
		*(parts[i]) = f
	}
	return pos, nil
}

func (svc *playerService) getRotation(username string) (*Orientation, error) {
	cmd := fmt.Sprintf("data get entity %s Rotation", username)
	out, err := svc.client.Execute(cmd)
	if err != nil {
		return nil, errors.Wrap(err, "execute command")
	}
	if out == "No entity was found" { // player disconnected
		return nil, ErrNotFound
	}

	// Parse output.
	out = out[strings.LastIndexByte(out, ':')+2:]
	out = strings.Trim(out, "[]")

	// Parse output parts into orient.
	var (
		orient = new(Orientation)
		parts  = []*float32{&orient.X, &orient.Y}
	)
	for i, part := range strings.Split(out, ", ") {
		f, err := strconv.ParseFloat(part[:len(part)-1], 32)
		if err != nil {
			return nil, errors.Wrap(err, "parse rotation part")
		}
		*(parts[i]) = float32(f)
	}
	return orient, nil
}

func (svc *playerService) listUsernames() ([]string, error) {
	out, err := svc.client.Execute("list")
	if err != nil {
		return nil, err
	}
	out = out[strings.LastIndexByte(out, ':')+2:]
	if out == "" {
		return nil, nil
	}
	return strings.Split(out, ","), nil
}

// ErrNotFound is returned when an entity could not be found.
var ErrNotFound = stderrors.New("minecraft: not found")
