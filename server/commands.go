package main

import (
	"fmt"
	"github.com/mattermost/mattermost-server/v5/model"
	"net/http"
)

func (p *Plugin) ExecuteCommandConfig(channelId, triggerId string) (*model.CommandResponse, *model.AppError) {
	p.API.LogInfo(fmt.Sprintf("%t", p.chai == nil))
	err := p.chai.OpenConfigDialog(channelId, triggerId)
	p.API.LogInfo(fmt.Sprintf("%t", p.API == nil))
	if err != nil {
		p.API.LogError(err.Error())
		p.API.LogError("Error occurred opening interactive dialog for setting config. Error: " + err.Error())
		return nil, model.NewAppError("ExecuteCommandConfig", "", nil, err.Error(), http.StatusInternalServerError)
	}

	return model.CommandResponseFromPlainText("Configure Chai Time in the open dialog"), nil
}

func (p *Plugin) ExecuteCommandJoin(userId, channelId string) (*model.CommandResponse, *model.AppError) {
	enabledChannels, err := p.chai.GetEnabledChannels()
	if err != nil {
		return model.CommandResponseFromPlainText("Error occurred joining channel Chai Time."), nil
	}

	found := false
	for location := range enabledChannels {
		if location == channelId {
			found = true
			break
		}
	}

	if !found {
		return model.CommandResponseFromPlainText("Channel is not part of Chai Time. Make sure the channel you're trying to join has Chat Time enabled."), nil
	}

	err = p.chai.AddChannelMember(userId, channelId)
	if err != nil {
		return model.CommandResponseFromPlainText("Error occurred join channel Chai Time."), nil
	}

	return model.CommandResponseFromPlainText("You've successfully joined the channel Chai Time."), nil
}

func (p *Plugin) ExecuteCommandLeave(userId, channelId string) (*model.CommandResponse, *model.AppError) {
	enabledChannels, err := p.chai.GetEnabledChannels()
	if err != nil {
		return model.CommandResponseFromPlainText("Error occurred leaving channel Chai Time."), nil
	}

	found := false
	for location := range enabledChannels {
		if location == channelId {
			found = true
			break
		}
	}

	if !found {
		return model.CommandResponseFromPlainText("Channel is not part of Chai Time."), nil
	}

	err = p.chai.RemoveChannelMember(userId, channelId)
	if err != nil {
		return model.CommandResponseFromPlainText("Error occurred join channel Chai Time."), nil
	}

	return model.CommandResponseFromPlainText("You've successfully left the channel Chai Time."), nil
}
