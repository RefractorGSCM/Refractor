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
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"strconv"
	"strings"
)

type executor struct {
	rconService domain.RCONService
	gameService domain.GameService
	logger      *zap.Logger
}

func NewCommandExecutor(rs domain.RCONService, gs domain.GameService, log *zap.Logger) domain.CommandExecutor {
	return &executor{
		rconService: rs,
		gameService: gs,
		logger:      log,
	}
}

func (e *executor) RunInfractionCommands(infrType, action string, payload *domain.PlayerCommandPayload, serverID int64, game domain.Game) error {
	// Get commands to run from game settings
	gameSettings, err := e.gameService.GetGameSettings(game)
	if err != nil {
		e.logger.Error("Could not get game settings from game repo",
			zap.String("Game name", game.GetName()),
			zap.Error(err))
		return err
	}

	if gameSettings.Commands == nil {
		e.logger.Warn("Could not run infraction commands as no commands are set")
		return nil
	}

	infrActionMap := gameSettings.Commands.InfractionActionMap()
	actMap := infrActionMap[action]
	if actMap == nil {
		return errors.New("no infraction action type: " + action)
	}

	cmdMap := actMap.Map()
	cmds := cmdMap[infrType]
	if cmds == nil {
		return errors.New("no commands found for infraction type: " + infrType)
	}

	// Run the command
	client := e.rconService.GetServerClient(serverID)

	// Parse and run the commands
	for _, cmd := range cmds {
		// Replace placeholders inside command with payload data
		runCmd := strings.ReplaceAll(cmd, "{{PLAYER_ID}}", payload.PlayerID)
		runCmd = strings.ReplaceAll(runCmd, "{{PLAYER_NAME}}", payload.Name)
		runCmd = strings.ReplaceAll(runCmd, "{{DURATION}}", strconv.FormatInt(payload.Duration, 10))
		runCmd = strings.ReplaceAll(runCmd, "{{REASON}}", payload.Reason)

		res, err := client.ExecCommand(runCmd)
		if err != nil {
			e.logger.Error("Could not execute ban command",
				zap.String("Player ID", payload.PlayerID),
				zap.String("Platform", payload.Platform),
				zap.Int64("Server ID", serverID),
				zap.Error(err))
			return err
		}

		e.logger.Info("Ran Infraction Command",
			zap.String("Command", runCmd),
			zap.Int64("Server ID", serverID),
			zap.String("Player ID", payload.PlayerID),
			zap.String("Platform", payload.Platform),
			zap.String("Reason", payload.Reason),
			zap.String("Response From Server", res))
	}

	return nil
}
