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
			AlivePingInterval:         time.Second * 30,
			EnableBroadcasts:          true,
			PlayerListPollingInterval: time.Hour * 1,
			EnableChat:                true,
			BroadcastPatterns:         map[string]*regexp.Regexp{},
		},
		platform: platform,
		cmdOutputPatterns: &domain.CommandOutputPatterns{
			PlayerList: regexp.MustCompile("(?P<PlayerID>[0-9A-Z]+),\\\\s(?P<Name>[\\\\S ]+),\\\\s(?P<Ping>\\\\d{1,4})\\\\sms,\\\\steam\\\\s(?P<Team>[0-9-]+)"),
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
