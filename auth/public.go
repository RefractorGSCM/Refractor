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

package auth

import (
	"Refractor/pkg/conf"
	kratos "github.com/ory/kratos-client-go"
)

type publicHandlers struct {
	client *kratos.APIClient
	config *conf.Config
}

func NewPublicHandlers(client *kratos.APIClient, config *conf.Config) *publicHandlers {
	return &publicHandlers{
		client: client,
		config: config,
	}
}

type Node struct {
	Label    string
	Disabled bool
	Name     string
	Required bool
	Type     string
	Value    string
}

type Message struct {
	ID   int64
	Text string
	Type string
}

type RenderData struct {
	Action   string
	Method   string
	FlowID   string
	UiNodes  []Node
	Messages []Message
}
