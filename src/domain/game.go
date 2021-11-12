/*
 * This file is part of Refractor.
 *
 * Refractor is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package domain

import (
	"github.com/refractorgscm/rcon"
	"github.com/refractorgscm/rcon/endian"
	"regexp"
	"time"
)

var AllGames []string

// Game is the interface representing a game within Refractor.
type Game interface {
	GetName() string
	GetConfig() *GameConfig
	GetPlatform() Platform
	GetPlayerListCommand() string
	GetCommandOutputPatterns() *CommandOutputPatterns
	GetBroadcastCommand() string
	GetRCONSettings() *GameRCONSettings
	GetDefaultSettings() *GameSettings
}

// GameRCONSettings is a struct which holds RCON settings for a single game.
type GameRCONSettings struct {
	// RestrictedPacketIDs is an int32 slice which gets passed into a game's RCON client upon creation.
	// RCON works by sending data packets with a unique ID to the server, and the server responds with a data packet of
	// its own with a matching ID to the original request.
	//
	// RCON packet IDs are usually kept unique by simply incrementing a number every time a new packet is created.
	//
	// For some games, there are special (or reserved) packet IDs which mean specific things. Due to this, we may want
	// to make sure that the auto incrementing packet ID never becomes any of those restricted IDs to prevent confusion
	// between the game server and the RCON client.
	RestrictedPacketIDs []int32

	// BroadcastChecker is a function which will be called with each RCON packet sent by the game server as an argument.
	// It should check if the passed in packet is a broadcast using any relevant criteria. True is returned if the
	// checked packet is a broadcast, false otherwise.
	BroadcastChecker rcon.BroadcastMessageChecker

	// EndianMode represents the byte order used by a game's RCON implementation.
	EndianMode endian.Mode
}

type CommandOutputPatterns struct {
	PlayerList *regexp.Regexp // required
}

type GameConfig struct {
	UseRCON bool

	// AlivePingInterval is the interval on which alive pings are sent to the server to keep the RCON
	// connection alive. Set this to 0 to disable alive ping.
	AlivePingInterval time.Duration

	// EnableBroadcasts should be set to true if this game supports broadcasts. Broadcasts are real-time
	// notifications from the server of various events such as player join, player quit, etc.
	EnableBroadcasts bool

	// RCONInitCommands is a string slice containing commands which will be executed on the broadcast listener
	// socket when connection is established.
	RCONInitCommands []string

	// BroadcastPatterns is a map containing the regex patterns of various broadcast types. These are used to
	// parse data inside the broadcasts. If EnableBroadcasts is false, this can safely be set to nil or unset.
	BroadcastPatterns map[string]*regexp.Regexp

	// IgnoredBroadcastPatterns is a slice containing the regex patterns of messages which could come over the broadcast
	// handler which should be ignored.
	IgnoredBroadcastPatterns []*regexp.Regexp

	// EnableLiveChat enables live chat for this game if set to true.
	EnableChat bool

	// PlayerListPollingInterval is the interval at which the server is manually polled to fully update the player
	// list. This can be quite useful for games which support broadcasts where sometimes things can get out of sync.
	// For broadcast enabled games, setting this interval to 1 hour should be plenty sufficient. If the game's
	// broadcast system is very stable then you may not need this at all. If EnableBroadcasts is set to false, you
	// must set PlayListPollingInterval or else the player list will never be updated!
	PlayerListPollingInterval time.Duration

	// PlayerListRefreshInterval is the interval at which the server's player list is fully refreshed and pushed to all
	// necessary services and clients. This should be enabled for all games to mitigate server desyncs, but is especially
	// important for games which support broadcasts since desyncs are more likely.
	PlayerListRefreshInterval time.Duration
}

func (gc GameConfig) AlivePingEnabled() bool {
	return gc.AlivePingInterval != 0
}

func (gc GameConfig) PlayerListPollingEnabled() bool {
	return gc.PlayerListPollingInterval != 0
}

func (gc GameConfig) PlayerListRefreshEnabled() bool {
	return gc.PlayerListPollingInterval != 0
}

// InfractionDetectionEnabled returns true if an infraction broadcast pattern is set and broadcasts are enabled.
func (gc GameConfig) InfractionDetectionEnabled() bool {
	return gc.EnableBroadcasts && gc.BroadcastPatterns["INFRACTION"] != nil
}

type GameService interface {
	AddGame(game Game)
	GetAllGames() []Game
	GameExists(name string) bool
	GetGame(name string) (Game, error)
	GetGameSettings(game Game) (*GameSettings, error)
	SetGameSettings(game Game, settings *GameSettings) error
}

type GameCommandSettings struct {
	CreateInfractionCommands *InfractionCommands `json:"create"`
	UpdateInfractionCommands *InfractionCommands `json:"update"`
	DeleteInfractionCommands *InfractionCommands `json:"delete"`
	RepealInfractionCommands *InfractionCommands `json:"repeal"`
	SyncInfractionCommands   *InfractionCommands `json:"sync"`
}

type GeneralSettings struct {
	EnableBanSync  bool `json:"enable_ban_sync"`
	EnableMuteSync bool `json:"enable_mute_sync"`
}

type GameSettings struct {
	Commands *GameCommandSettings `json:"commands"`
	General  *GeneralSettings     `json:"general"`
}

func (gcs *GameCommandSettings) InfractionActionMap() map[string]*InfractionCommands {
	return map[string]*InfractionCommands{
		InfractionCommandCreate: gcs.CreateInfractionCommands,
		InfractionCommandUpdate: gcs.UpdateInfractionCommands,
		InfractionCommandDelete: gcs.DeleteInfractionCommands,
		InfractionCommandRepeal: gcs.RepealInfractionCommands,
		InfractionCommandSync:   gcs.SyncInfractionCommands,
	}
}

type GameRepo interface {
	GetSettings(game Game) (*GameSettings, error)
	SetSettings(game Game, settings *GameSettings) error
}
