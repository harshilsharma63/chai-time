package main

import (
	"net/http"
	"strings"

	"github.com/mattermost/mattermost-server/v6/model"
)

func (p *Plugin) ExecuteCommandConfig(userID, channelID, triggerId string) (*model.CommandResponse, *model.AppError) {
	can, appErr := p.CanConfigChannel(userID, channelID)
	if appErr != nil {
		return nil, appErr
	}

	if !can {
		return model.CommandResponseFromPlainText("You do not have permission to perform this operation."), nil
	}

	err := p.chai.OpenConfigDialog(channelID, triggerId)
	if err != nil {
		p.API.LogError(err.Error())
		p.API.LogError("Error occurred opening interactive dialog for setting config. Error: " + err.Error())
		return nil, model.NewAppError("ExecuteCommandConfig", "", nil, err.Error(), http.StatusInternalServerError)
	}

	return model.CommandResponseFromPlainText("Configure Chai Time in the open dialog"), nil
}

func (p *Plugin) CanConfigChannel(userID, channelID string) (bool, *model.AppError) {
	channelMembers, appErr := p.API.GetChannelMembersByIds(channelID, []string{userID})
	if appErr != nil || channelMembers == nil || len(channelMembers) != 1 {
		p.API.LogError("error occurred fetching user channel roles", "userID", userID, "channelID", channelMembers, "error", appErr)
		return false, appErr
	}

	p.API.LogDebug("user roles: " + (channelMembers)[0].Roles)
	return strings.Contains((channelMembers)[0].Roles, model.ChannelAdminRoleId), nil
}

func (p *Plugin) ExecuteCommandJoin(userID, channelID string) (*model.CommandResponse, *model.AppError) {
	enabledChannels, err := p.chai.GetEnabledChannels()
	if err != nil {
		return model.CommandResponseFromPlainText("Error occurred joining channel Chai Time."), nil
	}

	found := false
	for location := range enabledChannels {
		if location == channelID {
			found = true
			break
		}
	}

	if !found {
		return model.CommandResponseFromPlainText("Channel is not part of Chai Time. Make sure the channel you're trying to join has Chat Time enabled."), nil
	}

	err = p.chai.AddChannelMember(userID, channelID)
	if err != nil {
		return model.CommandResponseFromPlainText("Error occurred join channel Chai Time."), nil
	}

	return model.CommandResponseFromPlainText("You've successfully joined the channel Chai Time."), nil
}

func (p *Plugin) ExecuteCommandLeave(userID, channelId string) (*model.CommandResponse, *model.AppError) {
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

	err = p.chai.RemoveChannelMember(userID, channelId)
	if err != nil {
		return model.CommandResponseFromPlainText("Error occurred join channel Chai Time."), nil
	}

	return model.CommandResponseFromPlainText("You've successfully left the channel Chai Time."), nil
}
