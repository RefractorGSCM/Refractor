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

package websocket

import (
	"Refractor/domain"
	"encoding/json"
	"github.com/gobwas/ws/wsutil"
	"go.uber.org/zap"
	"io"
	"net"
)

type ChatSendHandler func(body *SendChatBody)

type Client struct {
	ID              int64
	UserID          string
	Conn            net.Conn
	Pool            *Pool
	ChatSendHandler ChatSendHandler
	logger          *zap.Logger
}

var nextClientID int64 = 0

func NewClient(userID string, conn net.Conn, pool *Pool, csh ChatSendHandler, log *zap.Logger) *Client {
	nextClientID++

	return &Client{
		ID:              nextClientID,
		UserID:          userID,
		Conn:            conn,
		Pool:            pool,
		ChatSendHandler: csh,
		logger:          log,
	}
}

const expectedExitCode = 1001

type SendChatBody struct {
	ServerID int64  `json:"server_id"`
	Message  string `json:"message"`
	UserID   string
}

func (c *Client) Read() {
	defer func() {
		c.Pool.Unregister <- c
		_ = c.Conn.Close()
	}()

	for {
		msgBytes, _, err := wsutil.ReadClientData(c.Conn)
		if err != nil {
			// EOF is an expected error which occurs when a disconnection happens, so we return since this client
			// connection has been closed.
			if err == io.EOF {
				return
			}

			// If the error is of type wsutil.ClosedError and the error code was the expected Going Away code (1001)
			// then we return since this is also an expected and unconsequential error.
			if wsErr, ok := err.(wsutil.ClosedError); ok && wsErr.Code == expectedExitCode {
				return
			}

			c.logger.Error(
				"Could not read message from client",
				zap.Int64("Client ID", c.ID),
				zap.String("User ID", c.UserID),
				zap.Error(err),
			)
			continue
		}

		var msg *domain.WebsocketMessage
		if err = json.Unmarshal(msgBytes, &msg); err != nil {
			c.logger.Error(
				"Could not unmarshal message from client",
				zap.Int64("Client ID", c.ID),
				zap.String("User ID", c.UserID),
				zap.String("Message", string(msgBytes)),
				zap.Error(err),
			)

			continue
		}

		if msg.Type == "ping" {
			reply := &domain.WebsocketMessage{
				Type: "pong",
				Body: "",
			}

			msgBytes, err := json.Marshal(reply)
			if err != nil {
				c.logger.Error(
					"Could not marshal ping reply message",
					zap.Int64("Client ID", c.ID),
					zap.String("User ID", c.UserID),
					zap.Error(err),
				)
				continue
			}

			// Send pong message
			if err := wsutil.WriteServerText(c.Conn, msgBytes); err != nil {
				c.logger.Error(
					"Could not send ping reply message",
					zap.Int64("Client ID", c.ID),
					zap.String("User ID", c.UserID),
					zap.Error(err),
				)
				continue
			}

			// Continue as there's no need to log a ping message
			continue
		}

		if msg.Type == "chat" {
			data, err := json.Marshal(msg.Body)
			if err != nil {
				c.logger.Error(
					"Could not marshal chat message",
					zap.Int64("Client ID", c.ID),
					zap.String("User ID", c.UserID),
					zap.Error(err),
				)
				continue
			}

			body := &SendChatBody{}
			if err := json.Unmarshal(data, body); err != nil {
				c.logger.Error(
					"Could not unmarshal chat send body",
					zap.Int64("Client ID", c.ID),
					zap.String("User ID", c.UserID),
					zap.Error(err),
				)
				continue
			}

			body.UserID = c.UserID

			c.ChatSendHandler(body)
		}

		c.logger.Info(
			"Received message from client",
			zap.Int64("Client ID", c.ID),
			zap.String("User ID", c.UserID),
			zap.Any("Message", msg),
			zap.Error(err),
		)
	}
}
