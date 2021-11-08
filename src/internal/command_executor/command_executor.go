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

import (
	"Refractor/domain"
	"context"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"strconv"
	"strings"
)

type executor struct {
	rconService domain.RCONService
	gameService domain.GameService
	serverRepo  domain.ServerRepo
	logger      *zap.Logger
}

func NewCommandExecutor(rs domain.RCONService, gs domain.GameService, sr domain.ServerRepo, log *zap.Logger) domain.CommandExecutor {
	return &executor{
		rconService: rs,
		gameService: gs,
		serverRepo:  sr,
		logger:      log,
	}
}

func (e *executor) PrepareInfractionCommands(ctx context.Context, infraction domain.InfractionPayload, action string,
	serverID int64) (domain.CommandPayload, error) {

	// Make sure player name is set on infraction
	if infraction.GetPlayerName() == "" {
		return nil, errors.New("player name must be set")
	}

	// Get server
	server, err := e.serverRepo.GetByID(ctx, serverID)
	if err != nil {
		return nil, err
	}

	// Get server game
	game, err := e.gameService.GetGame(server.Game)
	if err != nil {
		return nil, err
	}

	// Get commands to run from game settings
	gameSettings, err := e.gameService.GetGameSettings(game)
	if err != nil {
		e.logger.Error("Could not get game settings from game repo",
			zap.String("Game name", game.GetName()),
			zap.Error(err))
		return nil, err
	}

	// Check if this game has commands set
	if gameSettings.Commands == nil {
		e.logger.Warn("Could not prepare infraction commands as no commands are set")
		return nil, domain.ErrNotFound
	}

	// Determine commands to prepare based on action and infraction type
	infrActionMap := gameSettings.Commands.InfractionActionMap()
	actMap := infrActionMap[action]
	if actMap == nil {
		return nil, errors.New("no infraction action type: " + action)
	}

	cmdMap := actMap.Map()
	cmds := cmdMap[infraction.GetType()]
	if cmds == nil {
		return nil, errors.New("no commands found for infraction type: " + infraction.GetType())
	}

	// Prepare the commands
	commands := make([]string, 0)

	// Parse and run the commands
	for _, cmd := range cmds {
		// Replace placeholders inside command with payload data
		runCmd := strings.ReplaceAll(cmd, "{{PLAYER_ID}}", infraction.GetPlayerID())
		runCmd = strings.ReplaceAll(runCmd, "{{PLATFORM}}", infraction.GetPlatform())
		runCmd = strings.ReplaceAll(runCmd, "{{PLAYER_NAME}}", infraction.GetPlayerName())
		runCmd = strings.ReplaceAll(runCmd, "{{DURATION}}", strconv.FormatInt(infraction.GetDuration(), 10))
		runCmd = strings.ReplaceAll(runCmd, "{{REASON}}", infraction.GetReason())

		commands = append(commands, runCmd)
	}

	return newInfractionCommand(commands, []int64{serverID}), nil
}

func (e *executor) RunCommands(payload domain.CommandPayload) error {
	for _, serverID := range payload.GetServerIDs() {
		client := e.rconService.GetServerClient(serverID)
		if client == nil {
			e.logger.Warn("Could not run commands on server. RCON client was nil.", zap.Int64("Server ID", serverID))
			continue
		}

		// TODO: Implement command queue or similar mechanism which can record commands which could not be executed
		// TODO: (e.g because client was nil) to be executed at a later time.

		for _, cmd := range payload.GetCommands() {
			if err := client.ExecCommandNoResponse(cmd); err != nil {
				e.logger.Error("Could not execute command on server",
					zap.String("Command", cmd),
					zap.Int64("Server ID", serverID),
					zap.Error(err))
				return err
			}

			e.logger.Info("Executed command on server",
				zap.String("Command", cmd),
				zap.Int64("Server ID", serverID))
		}
	}

	return nil
}
