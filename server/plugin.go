package main

import (
	"encoding/json"
	"github.com/mattermost/mattermost-plugin-starter-template/server/chai"
	"github.com/mattermost/mattermost-server/v5/model"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"

	"github.com/mattermost/mattermost-server/v5/plugin"
)

var bot = &model.Bot{
	Username:    "chaibot",
	DisplayName: "Chai Bot",
	Description: "Getting staff to kow each other, one chai at a time",
}

// Plugin implements the interface expected by the Mattermost server to communicate between the server and plugin processes.
type Plugin struct {
	plugin.MattermostPlugin

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *configuration

	chai *chai.Chai
}

func (p *Plugin) OnActivate() error {
	if err := p.OnConfigurationChange(); err != nil {
		return err
	}

	if err := p.ensureBot(); err != nil {
		return err
	}

	if err := p.registerSlashCommands(); err != nil {
		return err
	}

	p.chai = &chai.Chai{
		API: p.API,
	}

	return nil
}

func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	if args.Command == "/chai config" {
		return p.ExecuteCommandConfig(args.ChannelId, args.TriggerId)
	}

	return nil, nil
}

func (p *Plugin) ensureBot() error {
	botID, err := p.Helpers.EnsureBot(bot, plugin.ProfileImagePath("/assets/profile.gif"))
	if err != nil {
		p.API.LogError("Error occurred ensuring chaibot exists. Error: " + err.Error())
		return err
	}

	config := p.getConfiguration()
	config.BotID = botID
	p.setConfiguration(config)

	return nil
}

func (p *Plugin) registerSlashCommands() error {
	return p.API.RegisterCommand(&model.Command{
		Trigger:      "chai",
		IconURL:      "/assets/profile.gif",
		AutoComplete: true,
		DisplayName:  "display name",
		Description:  "description",
		AutocompleteData: &model.AutocompleteData{
			Trigger: "chai",
			RoleID:  model.SYSTEM_USER_ROLE_ID,
			SubCommands: []*model.AutocompleteData{
				{
					Trigger: "config",
					RoleID:  model.SYSTEM_USER_ROLE_ID,
				},
			},
		},
	})
}

// ServeHTTP demonstrates a plugin that handles HTTP requests by greeting the world.
func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	p.API.LogInfo(path)

	if path == "/saveConfig" {
		p.handleSaveConfig(c, w, r)
	}
}

func (p *Plugin) handleSaveConfig(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	var request *model.SubmitDialogRequest
	if err := json.Unmarshal(body, &request); err != nil {
		p.API.LogError("Error occurred unmarshalling save config API request body", "error", err.Error())
		http.Error(w, "Error occurred unmarshalling save config API request body", http.StatusInternalServerError)
		return
	}

	frequency, err := strconv.Atoi(request.Submission["frequency"].(string))
	if err != nil {
		p.API.LogError("Error occurred converting frequency string to number.", "frequencyString", request.Submission["frequency"], "error", err.Error())
		http.Error(w, "Error occurred converting frequency string to number", http.StatusInternalServerError)
		return
	}

	config := chai.Config{
		ChannelID: request.ChannelId,
		Frequency: frequency,
		DayOfWeek: request.Submission["dayOfWeek"].(string),
	}

	if err := p.chai.SaveConfig(config); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	p.API.SendEphemeralPost(request.UserId, &model.Post{
		UserId: request.UserId,
		Message: "Chai Time config has been saved successfully.",
		ChannelId: request.ChannelId,
	})

	w.Write(body)
	w.WriteHeader(http.StatusOK)
}

// See https://developers.mattermost.com/extend/plugins/server/reference/
