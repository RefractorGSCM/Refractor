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

import (
	"context"
	"github.com/guregu/null"
	"time"
)

type ChatReceiveBody struct {
	ServerID   int64  `json:"server_id"`
	PlayerID   string `json:"player_id"`
	Platform   string `json:"platform"`
	Name       string `json:"name"`
	Message    string `json:"message"`
	SentByUser bool   `json:"sent_by_user"`
}

type ChatMessage struct {
	MessageID  int64     `json:"id"`
	PlayerID   string    `json:"player_id"`
	Platform   string    `json:"platform"`
	ServerID   int64     `json:"server_id"`
	Message    string    `json:"message"`
	Flagged    bool      `json:"flagged"`
	CreatedAt  time.Time `json:"created_at"`
	ModifiedAt null.Time `json:"modified_at"`
	PlayerName string    `json:"player_name,omitempty"` // not a db field, must be populated manually.
}

type ChatRepo interface {
	Store(ctx context.Context, message *ChatMessage) error
	GetByID(ctx context.Context, id int64) (*ChatMessage, error)
}
