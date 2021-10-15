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

func (e *executor) RunBanCommand(ctx context.Context, payload *domain.PlayerCommandPayload, serverID int64, game domain.Game) error {
	// Get ban command from game settings
	gameSettings, err := e.gameService.GetGameSettings(game)
	if err != nil {
		e.logger.Error("Could not get game settings from game repo",
			zap.String("Game name", game.GetName()),
			zap.Error(err))
		return err
	}

	banCmd := gameSettings.BanCommandPattern

	// Replace placeholders inside banCmd with payload data
	banCmd = strings.ReplaceAll(banCmd, "{{PLAYER_ID}}", payload.PlayerID)
	banCmd = strings.ReplaceAll(banCmd, "{{PLAYER_NAME}}", payload.Name)
	banCmd = strings.ReplaceAll(banCmd, "{{DURATION}}", strconv.FormatInt(payload.Duration, 10))
	banCmd = strings.ReplaceAll(banCmd, "{{REASON}}", payload.Reason)

	// Ban the player for the correct remainder
	client := e.rconService.GetServerClient(serverID)
	res, err := client.ExecCommand(banCmd)
	if err != nil {
		e.logger.Error("Could not execute ban command",
			zap.String("Player ID", payload.PlayerID),
			zap.String("Platform", payload.Platform),
			zap.Int64("Server ID", serverID),
			zap.Error(err))
		return err
	}

	e.logger.Info("Banned player from server",
		zap.Int64("Server ID", serverID),
		zap.String("Player ID", payload.PlayerID),
		zap.String("Platform", payload.Platform),
		zap.String("Reason", payload.Reason),
		zap.String("Response From Server", res))

	return nil
}
