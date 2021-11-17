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

package command_executor

import "Refractor/domain"

type infractionCommandPayload struct {
	Commands []domain.Command
	Game     domain.Game
}

func newInfractionCommandPayload(cmds []domain.Command, game domain.Game) domain.CommandPayload {
	return &infractionCommandPayload{cmds, game}
}

func (ic *infractionCommandPayload) GetCommands() []domain.Command {
	return ic.Commands
}

func (ic *infractionCommandPayload) GetGame() domain.Game {
	return ic.Game
}

type infractionCommand struct {
	Command  string
	RunOnAll bool
	ServerID int64
}

func (i *infractionCommand) GetCommand() string {
	return i.Command
}

func (i *infractionCommand) ShouldRunOnAll() bool {
	return i.RunOnAll
}

func (i *infractionCommand) GetServerID() int64 {
	return i.ServerID
}
