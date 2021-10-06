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

package watchdog

import (
	"Refractor/domain"
	"context"
	"go.uber.org/zap"
	"time"
)

func StartRCONServerWatchdog(rconService domain.RCONService, serverService domain.ServerService, log *zap.Logger) error {
	for {
		// Run every 15 seconds
		time.Sleep(time.Second * 5)

		rconClients := rconService.GetClients()
		allServerData, err := serverService.GetAllServerData()
		if err != nil {
			log.Error("Watchdog routine could not get all server data", zap.Error(err))
			return err
		}

		if len(rconClients) != len(allServerData) {
			for _, serverData := range allServerData {
				client := rconClients[serverData.ServerID]

				// Check if an RCON client exists for this server. If not, create one
				if client == nil {
					if !serverData.ReconnectInProgress {
						server, err := serverService.GetByID(context.TODO(), serverData.ServerID)
						if err != nil || server == nil {
							log.Error(
								"Watchdog routine could not get server data",
								zap.Int64("Server", serverData.ServerID),
								zap.Error(err),
							)
							continue
						}

						// Start reconnection routine
						log.Info("Reconnect routine not started. Starting...", zap.Int64("Server", server.ID))
						go rconService.StartReconnectRoutine(server, serverData)
						serverData.ReconnectInProgress = true
					}
				}
			}
		}
	}
}
