package playfab

import "Refractor/domain"

type playfab struct{}

func NewPlayfabPlatform() domain.Platform {
	return &playfab{}
}

func (p *playfab) GetName() string {
	return "PlayFab"
}

func (p *playfab) GetPlayerIDField() string {
	return "PlayFabID"
}
