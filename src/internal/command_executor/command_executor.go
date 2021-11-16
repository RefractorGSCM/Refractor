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
	rconService    domain.RCONService
	gameService    domain.GameService
	serverRepo     domain.ServerRepo
	playerNameRepo domain.PlayerNameRepo
	logger         *zap.Logger
	queue          chan *queuedCommand
}

type queuedCommand struct {
	cmd      string
	serverID int64
}

func NewCommandExecutor(rs domain.RCONService, gs domain.GameService, sr domain.ServerRepo, pnr domain.PlayerNameRepo,
	log *zap.Logger) domain.CommandExecutor {
	return &executor{
		rconService:    rs,
		gameService:    gs,
		serverRepo:     sr,
		playerNameRepo: pnr,
		logger:         log,
		queue:          make(chan *queuedCommand, 100),
	}
}

func (e *executor) PrepareInfractionCommands(ctx context.Context, infraction domain.InfractionPayload, action string,
	serverID int64) (domain.CommandPayload, error) {

	// If player name is not set, fetch it
	playerName := infraction.GetPlayerName()
	if playerName == "" {
		name, _, err := e.playerNameRepo.GetNames(ctx, infraction.GetPlayerID(), infraction.GetPlatform())
		if err != nil {
			e.logger.Error("Could not get player name",
				zap.String("Player ID", infraction.GetPlayerID()),
				zap.String("Platform", infraction.GetPlatform()),
				zap.Error(err))
			return nil, err
		}

		playerName = name
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
	commands := make([]domain.Command, 0)

	// Parse and run the commands
	for _, cmd := range cmds {
		// Replace placeholders inside command with payload data
		runCmd := strings.ReplaceAll(cmd.Command, "{{PLAYER_ID}}", infraction.GetPlayerID())
		runCmd = strings.ReplaceAll(runCmd, "{{PLATFORM}}", infraction.GetPlatform())
		runCmd = strings.ReplaceAll(runCmd, "{{PLAYER_NAME}}", playerName)
		runCmd = strings.ReplaceAll(runCmd, "{{DURATION}}", strconv.FormatInt(infraction.GetDuration(), 10))
		runCmd = strings.ReplaceAll(runCmd, "{{REASON}}", infraction.GetReason())

		commands = append(commands, &infractionCommand{
			Command:  runCmd,
			RunOnAll: cmd.RunOnAll,
			ServerID: serverID,
		})
	}

	return newInfractionCommandPayload(commands, game), nil
}

func (e *executor) QueueCommands(payload domain.CommandPayload) error {
	game := payload.GetGame()
	cmds := payload.GetCommands()

	// Get servers of this game
	serversOfGame, err := e.serverRepo.GetByGame(context.TODO(), game.GetName())
	if err != nil {
		e.logger.Error("Command executor could not get servers by game", zap.String("Game",
			game.GetName()),
			zap.Error(err))
		return err
	}

	for _, cmd := range cmds {
		if !cmd.ShouldRunOnAll() {
			// Only run on the specified server
			e.queue <- &queuedCommand{
				cmd:      cmd.GetCommand(),
				serverID: cmd.GetServerID(),
			}
			continue
		}

		// Queue on all servers running this game
		for _, server := range serversOfGame {
			if server.Deactivated {
				// do not run on deactivated servers
				continue
			}

			// Add command to queue
			e.queue <- &queuedCommand{
				cmd:      cmd.GetCommand(),
				serverID: server.ID,
			}
		}
	}

	return nil
}

// StartRunner is a runner routine which reads from the command queue and executes the commands within
// it on the correct server. This should only be used when the command response doesn't matter because
// the response is not returned out of the runner.
func (e *executor) StartRunner(terminate chan uint8) {
	for {
		select {
		case <-terminate:
			e.logger.Info("Terminating command runner routine")
			break
		case queuedCmd := <-e.queue:
			client := e.rconService.GetServerClient(queuedCmd.serverID)
			if client == nil {
				e.logger.Warn("Could not run commands on server. RCON client was nil.", zap.Int64("Server ID", queuedCmd.serverID))
				break
			}

			e.logger.Info("Running command", zap.String("cmd", queuedCmd.cmd))
			if _, err := client.ExecCommand(queuedCmd.cmd); err != nil {
				e.logger.Error("Could not execute command on server",
					zap.String("Command", queuedCmd.cmd),
					zap.Int64("Server ID", queuedCmd.serverID),
					zap.Error(err))
				break
			}

			e.logger.Info("Executed command on server",
				zap.String("Command", queuedCmd.cmd),
				zap.Int64("Server ID", queuedCmd.serverID))
		}
	}
}
