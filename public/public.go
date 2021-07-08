package public

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
