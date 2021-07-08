package public

import (
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	kratos "github.com/ory/kratos-client-go"
	"net/http"
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
	rData := RenderData{
		Action:   flow.Ui.GetAction(),
		Method:   flow.Ui.GetMethod(),
		UiNodes:  []Node{},
		Messages: []Message{},
	}

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
			if attrVal.String != nil {
				newNode.Value = *attrVal.String
			}

			attributes.UiNodeInputAttributes.GetRequired()
		}

		rData.UiNodes = append(rData.UiNodes, newNode)
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
