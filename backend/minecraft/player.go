package minecraft

import "context"

// A Player describes a player entity in Minecraft.
type Player struct {
	Username    string      `json:"username"`
	Position    Coordinates `json:"position"`
	Orientation Orientation `json:"orientation"`
}

// A PlayerService can get information about the Players on a server.
type PlayerService interface {
	Get(ctx context.Context, username string) (*Player, error)
	List(ctx context.Context) ([]*Player, error)
}
