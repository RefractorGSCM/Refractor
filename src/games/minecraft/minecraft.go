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

package minecraft

import (
	"Refractor/domain"
	"github.com/refractorgscm/rcon/endian"
	"regexp"
	"time"
)

type minecraft struct {
	config            *domain.GameConfig
	platform          domain.Platform
	cmdOutputPatterns *domain.CommandOutputPatterns
}

func NewMinecraftGame(platform domain.Platform) domain.Game {
	return &minecraft{
		config: &domain.GameConfig{
			UseRCON:                   true,
			AlivePingInterval:         time.Minute * 2,
			EnableBroadcasts:          false,
			PlayerListPollingInterval: time.Second * 5,
			PlayerListRefreshInterval: time.Minute * 40,
			EnableChat:                false,
			PermanentDurationValue:    999999999999,
		},
		platform: platform,
		cmdOutputPatterns: &domain.CommandOutputPatterns{
			PlayerList: regexp.MustCompile("(?P<PlayerID>[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}):(?P<Name>[\\S]+)"),
		},
	}
}

func (g *minecraft) GetName() string {
	return "Minecraft"
}

func (g *minecraft) GetConfig() *domain.GameConfig {
	return g.config
}

func (g *minecraft) GetPlatform() domain.Platform {
	return g.platform
}

func (g *minecraft) GetPlayerListCommand() string {
	return "PlayerList"
}

func (g *minecraft) GetCommandOutputPatterns() *domain.CommandOutputPatterns {
	return g.cmdOutputPatterns
}

func (g *minecraft) GetBroadcastCommand() string {
	return "say %s"
}

func (g *minecraft) GetDefaultSettings() *domain.GameSettings {
	return &domain.GameSettings{
		Commands: &domain.GameCommandSettings{
			CreateInfractionCommands: &domain.InfractionCommands{
				Warn: []string{},
				Mute: []string{},
				Kick: []string{"kick {{PLAYER_NAME}} {{REASON}}"},
				Ban:  []string{"ban {{PLAYER_NAME}} {{REASON}}"},
			},
			UpdateInfractionCommands: &domain.InfractionCommands{
				Warn: []string{},
				Mute: []string{},
				Kick: []string{"kick {{PLAYER_NAME}} {{REASON}}"},
				Ban:  []string{"ban {{PLAYER_NAME}} {{REASON}}"},
			},
			DeleteInfractionCommands: &domain.InfractionCommands{
				Warn: []string{},
				Mute: []string{},
				Kick: []string{},
				Ban:  []string{"pardon {{PLAYER_NAME}}"},
			},
			RepealInfractionCommands: &domain.InfractionCommands{
				Warn: []string{},
				Mute: []string{},
				Kick: []string{},
				Ban:  []string{"pardon {{PLAYER_NAME}}"},
			},
			SyncInfractionCommands: &domain.InfractionCommands{
				Ban:  []string{"ban {{PLAYER_NAME}} Refractor Ban Sync"},
				Mute: []string{},
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

func (g *minecraft) GetRCONSettings() *domain.GameRCONSettings {
	return &domain.GameRCONSettings{
		RestrictedPacketIDs: nil, // nil since minecraft has no relevant restricted RCON packet IDs.
		BroadcastChecker:    nil, // nil since minecraft doesn't support broadcasts.
		EndianMode:          endian.Little,
	}
}
