package main

import (
	"fmt"
	"github.com/mattermost/mattermost-server/v5/model"
	"net/http"
)

func (p *Plugin) ExecuteCommandConfig(channelId, triggerId string) (*model.CommandResponse, *model.AppError) {
	p.API.LogInfo("AAAAAA")
	p.API.LogInfo(fmt.Sprintf("%t", p.chai == nil))
	err := p.chai.OpenConfigDialog(channelId, triggerId)
	p.API.LogInfo("BBBBBBBB")
	p.API.LogInfo(fmt.Sprintf("%t", p.API == nil))
	if err != nil {
		p.API.LogInfo("CCCCCC")
		p.API.LogError(err.Error())
		p.API.LogInfo("CCCCCC")
		p.API.LogError("Error occurred opening interactive dialog for setting config. Error: " + err.Error())
		return nil, model.NewAppError("ExecuteCommandConfig", "", nil, err.Error(), http.StatusInternalServerError)
	}

	return model.CommandResponseFromPlainText("Configure Chai Time in the open dialog"), nil
}
