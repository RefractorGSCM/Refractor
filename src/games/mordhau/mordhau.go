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

package mordhau

import (
	"Refractor/domain"
	"Refractor/pkg/broadcast"
	"encoding/binary"
	"github.com/refractorgscm/rcon/presets"
	"regexp"
	"time"
)

type mordhau struct {
	config            *domain.GameConfig
	platform          domain.Platform
	cmdOutputPatterns *domain.CommandOutputPatterns
}

func NewMordhauGame(platform domain.Platform) domain.Game {
	return &mordhau{
		config: &domain.GameConfig{
			UseRCON:                   true,
			AlivePingInterval:         time.Second * 10,
			EnableBroadcasts:          true,
			RCONInitCommands:          []string{"listen login", "listen chat"},
			PlayerListRefreshInterval: time.Minute * 10,
			EnableChat:                true,
			BroadcastPatterns: map[string]*regexp.Regexp{
				broadcast.TypeJoin: regexp.MustCompile("^Login: (?P<Date>[0-9\\.-]+): (?P<Name>.+) \\((?P<PlayerID>[0-9a-fA-F]+)\\) logged in$"),
				broadcast.TypeQuit: regexp.MustCompile("^Login: (?P<Date>[0-9\\.-]+): (?P<Name>.+) \\((?P<PlayerID>[0-9a-fA-F]+)\\) logged out$"),
				broadcast.TypeChat: regexp.MustCompile("^Chat: (?P<PlayerID>[0-9a-fA-F]+), (?P<Name>.+), \\((?P<Channel>.+)\\) (?P<Message>.+)$"),
				//broadcast.TypeBan: regexp.MustCompile("^Punishment: Admin (?P<AdminName>.+) \\((?P<AdminPlayerID>[0-9a-fA-F]+)\\) banned player (?P<PlayerID>[0-9a-fA-F]+) \\(Duration: (?P<Duration>\\d+), Reason: (?P<Reason>.+)\\)$"),
			},
			IgnoredBroadcastPatterns: []*regexp.Regexp{
				regexp.MustCompile("Keeping client alive for another [0-9]+ seconds"),
			},
			PermanentDurationValue: 99999999,
		},
		platform: platform,
		cmdOutputPatterns: &domain.CommandOutputPatterns{
			PlayerList: regexp.MustCompile("(?P<PlayerID>[0-9A-Z]+),\\s(?P<Name>[\\S ]+),\\s(?P<Ping>\\d{1,4})\\sms,\\steam\\s(?P<Team>[0-9-]+)"),
		},
	}
}

func (g *mordhau) GetName() string {
	return "Mordhau"
}

func (g *mordhau) GetConfig() *domain.GameConfig {
	return g.config
}

func (g *mordhau) GetPlatform() domain.Platform {
	return g.platform
}

func (g *mordhau) GetPlayerListCommand() string {
	return "PlayerList"
}

func (g *mordhau) GetCommandOutputPatterns() *domain.CommandOutputPatterns {
	return g.cmdOutputPatterns
}

func (g *mordhau) GetBroadcastCommand() string {
	return "Say %s"
}

func (g *mordhau) GetDefaultSettings() *domain.GameSettings {
	return &domain.GameSettings{
		Commands: &domain.GameCommandSettings{
			CreateInfractionCommands: &domain.InfractionCommands{
				Warn: []*domain.InfractionCommand{},
				Mute: []*domain.InfractionCommand{
					{
						Command:  "Mute {{PLAYER_ID}} {{DURATION}}",
						RunOnAll: true,
					},
				},
				Kick: []*domain.InfractionCommand{
					{
						Command:  "Kick {{PLAYER_ID}} {{REASON}}",
						RunOnAll: false,
					},
				},
				Ban: []*domain.InfractionCommand{
					{
						Command:  "Ban {{PLAYER_ID}} {{DURATION}} {{REASON}}",
						RunOnAll: true,
					},
				},
			},
			UpdateInfractionCommands: &domain.InfractionCommands{
				Warn: []*domain.InfractionCommand{},
				Mute: []*domain.InfractionCommand{
					{
						Command:  "Mute {{PLAYER_ID}} {{DURATION_REMAINING}}",
						RunOnAll: true,
					},
				},
				Kick: []*domain.InfractionCommand{},
				Ban: []*domain.InfractionCommand{
					{
						Command:  "Ban {{PLAYER_ID}} {{DURATION_REMAINING}} {{REASON}}",
						RunOnAll: true,
					},
				},
			},
			DeleteInfractionCommands: &domain.InfractionCommands{
				Warn: []*domain.InfractionCommand{},
				Mute: []*domain.InfractionCommand{
					{
						Command:  "Unmute {{PLAYER_ID}}",
						RunOnAll: true,
					},
				},
				Kick: []*domain.InfractionCommand{},
				Ban: []*domain.InfractionCommand{
					{
						Command:  "Unban {{PLAYER_ID}}",
						RunOnAll: true,
					},
				},
			},
			RepealInfractionCommands: &domain.InfractionCommands{
				Warn: []*domain.InfractionCommand{},
				Mute: []*domain.InfractionCommand{
					{
						Command:  "Unmute {{PLAYER_ID}}",
						RunOnAll: true,
					},
				},
				Kick: []*domain.InfractionCommand{},
				Ban: []*domain.InfractionCommand{
					{
						Command:  "Unban {{PLAYER_ID}}",
						RunOnAll: true,
					},
				},
			},
			SyncInfractionCommands: &domain.InfractionCommands{
				Mute: []*domain.InfractionCommand{
					{
						Command:  "Mute {{PLAYER_ID}} {{DURATION_REMAINING}}",
						RunOnAll: false,
					},
				},
				Ban: []*domain.InfractionCommand{
					{
						Command:  "Ban {{PLAYER_ID}} {{DURATION_REMAINING}} Refractor Ban Sync",
						RunOnAll: false,
					},
				},
			},
		},
		General: &domain.GeneralSettings{
			EnableBanSync:             true,
			EnableMuteSync:            true,
			PlayerInfractionThreshold: 10,
			PlayerInfractionTimespan:  4320, // 3 days
		},
	}
}

func (g *mordhau) GetRCONSettings() *domain.GameRCONSettings {
	return &domain.GameRCONSettings{
		// Mordhau has presets within the RefractorGSCM/RCON library so we can simply use those.
		RestrictedPacketIDs: presets.MordhauRestrictedPacketIDs,
		BroadcastChecker:    presets.MordhauBroadcastChecker,
		EndianMode:          binary.LittleEndian,
	}
}
