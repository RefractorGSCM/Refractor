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
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	kratos "github.com/ory/kratos-client-go"
	"net/http"
	"strings"
)

func (h *publicHandlers) SettingsHandler(c echo.Context) error {
	flowID := c.QueryParam("flow")

	if flowID == "" {
		return c.Redirect(http.StatusTemporaryRedirect,
			fmt.Sprintf("%s/self-service/settings/browser", h.config.KratosPublic))
	}

	settingsURL := fmt.Sprintf("%s/self-service/settings/flows?id=%s", h.config.KratosPublic, flowID)

	req, err := http.NewRequest("GET", settingsURL, nil)
	if err != nil {
		return err
	}

	for _, cookie := range c.Cookies() {
		req.AddCookie(cookie)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		if res != nil && (res.StatusCode == http.StatusGone || res.StatusCode == http.StatusNotFound) {
			// the flow is invalid or is no longer valid
			return c.Redirect(http.StatusTemporaryRedirect,
				fmt.Sprintf("%s/self-service/settings/browser", h.config.KratosPublic))
		}

		return err
	}

	flow := kratos.SettingsFlow{}
	if err := json.NewDecoder(res.Body).Decode(&flow); err != nil {
		return err
	}

	//data, err := json.MarshalIndent(flow, "", "  ")
	//if err != nil {
	//	return err
	//}
	//
	//l, _ := zap.NewDevelopment()
	//l.Debug(string(data))

	// pass the flow data along to the renderer for display
	type rDataGroup struct {
		ProfileData        RenderData
		PasswordData       RenderData
		Messages           []Message
		ShowProfile        bool
		Success            bool
		SuccessRedirectURL string
		BackRedirectURL    string
	}

	rData := rDataGroup{
		ProfileData: RenderData{
			Action:  flow.Ui.GetAction(),
			Method:  flow.Ui.GetMethod(),
			UiNodes: []Node{},
		},
		PasswordData: RenderData{
			Action:  flow.Ui.GetAction(),
			Method:  flow.Ui.GetMethod(),
			UiNodes: []Node{},
		},
		Messages:           []Message{},
		ShowProfile:        true,
		Success:            false,
		SuccessRedirectURL: "http://127.0.0.1:3000", // TODO: don't hardcode these values
		BackRedirectURL:    "http://127.0.0.1:3000",
	}

	// If this flow was initialized by an account recovery, we do not want to show the profile update form
	// to avoid confusing the user.
	if strings.Contains(flow.RequestUrl, "/self-service/recovery") {
		rData.ShowProfile = false
	}

	// If the user set their password, then the flow should be marked as complete so we update the render data success
	// variable to adjust rendering accordingly.
	if flow.State == "success" {
		rData.Success = true
	}

	submitsSeen := 0

	for _, node := range flow.Ui.Nodes {
		newNode := Node{}

		if meta, ok := node.GetMetaOk(); ok {
			if label, ok := meta.GetLabelOk(); ok {
				newNode.Label = label.Text
			}
		}

		if attributes, ok := node.GetAttributesOk(); ok {
			newNode.Disabled = attributes.UiNodeInputAttributes.Disabled
			newNode.Name = attributes.UiNodeInputAttributes.Name
			newNode.Required = attributes.UiNodeInputAttributes.GetRequired()
			newNode.Type = attributes.UiNodeInputAttributes.Type

			attrVal := attributes.UiNodeInputAttributes.GetValue()
			if val, ok := attrVal.(string); ok {
				newNode.Value = val
			}

			attributes.UiNodeInputAttributes.GetRequired()
		}

		// Since the settings flow has two points of submission, we want to put the UI nodes in the correct form
		// based on where they belong. The following if/else if logic is determining where they belong.
		if newNode.Name == "traits.email" || newNode.Name == "traits.username" {
			// if this node belongs to the profile form, add it to profile data
			rData.ProfileData.UiNodes = append(rData.ProfileData.UiNodes, newNode)
		} else if newNode.Name == "csrf_token" {
			// if this node is the CSRF token, add it to both
			rData.ProfileData.UiNodes = append(rData.ProfileData.UiNodes, newNode)
			rData.PasswordData.UiNodes = append(rData.PasswordData.UiNodes, newNode)
		} else if newNode.Name == "password" {
			// if this node is the password, add it to the password form
			rData.PasswordData.UiNodes = append(rData.PasswordData.UiNodes, newNode)
		}

		if newNode.Type == "submit" {
			if submitsSeen == 0 {
				// if this is the first submit, put under profile data
				rData.ProfileData.UiNodes = append(rData.ProfileData.UiNodes, newNode)
			} else {
				// otherwise, put it under password data
				rData.PasswordData.UiNodes = append(rData.PasswordData.UiNodes, newNode)
			}

			submitsSeen++
		}
	}

	for _, message := range flow.Ui.Messages {
		newMessage := Message{
			ID:   message.Id,
			Text: message.Text,
			Type: message.Type,
		}

		rData.Messages = append(rData.Messages, newMessage)
	}

	return c.Render(http.StatusOK, "settings", rData)
}
