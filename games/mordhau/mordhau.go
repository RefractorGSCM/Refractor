package mordhau

import (
	"Refractor/domain"
	"regexp"
	"time"
)

type mordhau struct {
	config   *domain.GameConfig
	platform domain.Platform
}

func NewMordhauGame(platform domain.Platform) domain.Game {
	return &mordhau{
		config: &domain.GameConfig{
			UseRCON:                 true,
			AlivePingInterval:       time.Second * 30,
			EnableBroadcasts:        true,
			PlayListPollingInterval: time.Hour * 1,
			EnableChat:              true,
			BroadcastPatterns:       map[string]*regexp.Regexp{},
		},
		platform: platform,
	}
}

func (g *mordhau) GetName() string {
	return "Mordhau"
}

func (g *mordhau) GetConfig() *domain.GameConfig {
	return g.config
}

func (g *mordhau) GetPlatform() domain.Platform {
	return g.platform
}
