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
			EnableChat:                false,
		},
		platform: platform,
		cmdOutputPatterns: &domain.CommandOutputPatterns{
			PlayerList: regexp.MustCompile("(?P<MCUUID>[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}):(?P<Name>[\\S]+)"),
		},
	}
}

func (g *minecraft) GetName() string {
	return "minecraft"
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
