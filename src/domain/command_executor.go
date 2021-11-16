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

import "context"

type CommandPayload interface {
	GetCommands() []Command
	GetGame() Game
}

type Command interface {
	GetCommand() string
	ShouldRunOnAll() bool
	GetServerID() int64
}

type CommandExecutor interface {
	PrepareInfractionCommands(ctx context.Context, infraction InfractionPayload, action string, serverID int64) (CommandPayload, error)
	QueueCommands(payload CommandPayload) error
	StartRunner(terminate chan uint8)
}

type CustomInfractionPayload struct {
	PlayerID   string
	Platform   string
	PlayerName string
	Type       string
	Duration   int64
	Reason     string
}

func (p *CustomInfractionPayload) GetPlayerID() string {
	return p.PlayerID
}

func (p *CustomInfractionPayload) GetPlatform() string {
	return p.Platform
}

func (p *CustomInfractionPayload) GetPlayerName() string {
	return p.PlayerName
}

func (p *CustomInfractionPayload) GetType() string {
	return p.Type
}

func (p *CustomInfractionPayload) GetDuration() int64 {
	return p.Duration
}

func (p *CustomInfractionPayload) GetReason() string {
	return p.Reason
}
