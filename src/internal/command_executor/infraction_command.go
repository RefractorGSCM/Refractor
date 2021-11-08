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

type infractionCommand struct {
	Commands []string
	Servers  []int64
}

func newInfractionCommand(cmds []string, servers []int64) domain.CommandPayload {
	return &infractionCommand{cmds, servers}
}

func (ic *infractionCommand) GetCommands() []string {
	return ic.Commands
}

func (ic *infractionCommand) GetServerIDs() []int64 {
	return ic.Servers
}
