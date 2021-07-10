package domain

import (
	"regexp"
	"time"
)

// Game is the interface representing a game within Refractor.
type Game interface {
	GetName() string
	GetConfig() *GameConfig
	GetPlatform() Platform
}

type GameConfig struct {
	UseRCON bool

	// AlivePingInterval is the interval on which alive pings are sent to the server to keep the RCON
	// connection alive. Set this to 0 to disable alive ping.
	AlivePingInterval time.Duration

	// EnableBroadcasts should be set to true if this game supports broadcasts. Broadcasts are real-time
	// notifications from the server of various events such as player join, player quit, etc.
	EnableBroadcasts bool

	// BroadcastPatterns is a map containing the regex patterns of various broadcast types. These are used to
	// parse data inside the broadcasts. If EnableBroadcasts is false, this can safely be set to nil or unset.
	BroadcastPatterns map[string]*regexp.Regexp

	// EnableLiveChat enables live chat for this game if set to true.
	EnableChat bool

	// PlayerListPollingInterval is the interval at which the server is manually polled to fully update the player
	// list. This can be quite useful for games which support broadcasts where sometimes things can get out of sync.
	// For broadcast enabled games, setting this interval to 1 hour should be plenty sufficient. If the game's
	// broadcast system is very stable then you may not need this at all. If EnableBroadcasts is set to false, you
	// must set PlayListPollingInterval or else the player list will never be updated!
	PlayListPollingInterval time.Duration
}

type GameService interface {
	AddGame(game Game)
	GetAllGames() []Game
	GameExists(name string) bool
	GetGame(name string) (Game, error)
}
