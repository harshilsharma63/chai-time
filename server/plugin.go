package main

import (
	"fmt"
	"github.com/mattermost/mattermost-plugin-starter-template/server/chai"
	"github.com/mattermost/mattermost-server/v5/model"
	"net/http"
	"sync"

	"github.com/mattermost/mattermost-server/v5/plugin"
)

var bot = &model.Bot{
	Username: "chaibot",
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

	Chai *chai.Chai
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

	p.Chai = &chai.Chai{
		API: p.API,
	}

	return nil
}

func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	p.API.LogInfo("########################################################")
	p.API.LogInfo(fmt.Sprintf("-%s-", args.Command))
	p.API.LogInfo("########################################################")
	if args.Command == "/chai config" {

	}

	return nil,  nil
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
		Trigger: "chai",
		IconURL: "/assets/profile.gif",
		AutoComplete: true,
		DisplayName: "display name",
		Description: "description",
		AutocompleteData: &model.AutocompleteData{
			Trigger: "chai",
			RoleID:      model.SYSTEM_USER_ROLE_ID,
			SubCommands: []*model.AutocompleteData{
				{
					Trigger: "config",
					RoleID:   model.SYSTEM_USER_ROLE_ID,
				},
			},
		},
	})
}

// ServeHTTP demonstrates a plugin that handles HTTP requests by greeting the world.
func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello, world!")
}

// See https://developers.mattermost.com/extend/plugins/server/reference/
