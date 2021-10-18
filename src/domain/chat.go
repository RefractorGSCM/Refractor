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
	Name       string    `json:"name,omitempty"` // not a db field, must be populated manually.
}

type ChatRepo interface {
	Store(ctx context.Context, message *ChatMessage) error
	GetByID(ctx context.Context, id int64) (*ChatMessage, error)
	GetRecentByServer(ctx context.Context, serverID int64, count int) ([]*ChatMessage, error)
	Search(ctx context.Context, args FindArgs, limit, offset int) (int, []*ChatMessage, error)
	GetFlaggedMessages(ctx context.Context, count int, serverIDs []int64, random bool) ([]*ChatMessage, error)
	GetFlaggedMessageCount(ctx context.Context) (int, error)
	Update(ctx context.Context, id int64, args UpdateArgs) (*ChatMessage, error)
}

type ChatService interface {
	Store(c context.Context, message *ChatMessage) error
	GetRecentByServer(c context.Context, serverID int64, count int) ([]*ChatMessage, error)
	GetFlaggedMessages(c context.Context, count int, random bool) ([]*ChatMessage, error)
	HandleChatReceive(body *ChatReceiveBody, serverID int64, game Game)
	HandleUserSendChat(body *ChatSendBody)
	GetFlaggedMessageCount(c context.Context) (int, error)
	UnflagMessage(c context.Context, id int64) error
}

type FlaggedWord struct {
	ID   int64  `json:"id"`
	Word string `json:"word"`
}

type FlaggedWordRepo interface {
	Store(ctx context.Context, word *FlaggedWord) error
	GetAll(ctx context.Context) ([]*FlaggedWord, error)
	Update(ctx context.Context, id int64, newWord string) (*FlaggedWord, error)
	Delete(ctx context.Context, id int64) error
}

type FlaggedWordService interface {
	Store(c context.Context, word *FlaggedWord) error
	GetAll(c context.Context) ([]*FlaggedWord, error)
	Update(c context.Context, id int64, newWord string) (*FlaggedWord, error)
	Delete(c context.Context, id int64) error
	MessageContainsFlaggedWord(c context.Context, message string) (bool, error)
}
