package main

import (
	"encoding/json"
	"fmt"
	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-plugin-starter-template/server/chai"
	"github.com/mattermost/mattermost-server/v6/model"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/mattermost/mattermost-plugin-api/cluster"
	"github.com/mattermost/mattermost-server/v6/plugin"
)

const (
	jobInterval = 120 * time.Minute
)

var BotUserID string

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

	chai   *chai.Chai
	job    *cluster.Job
	client *pluginapi.Client
}

func (p *Plugin) OnActivate() error {
	if err := p.OnConfigurationChange(); err != nil {
		return err
	}

	p.initClient()
	if err := p.ensureBot(); err != nil {
		return err
	}

	if err := p.registerSlashCommands(); err != nil {
		return err
	}

	p.chai = &chai.Chai{
		API: p.API,
	}

	if err := p.run(); err != nil {
		p.API.LogError("Failed to schedule scheduled job.", "error", err.Error())
		return err
	}

	return nil
}

func (p *Plugin) initClient() {
	p.client = pluginapi.NewClient(p.API, p.Driver)
}

func (p *Plugin) run() error {
	if p.job != nil {
		if err := p.job.Close(); err != nil {
			return err
		}
	}

	job, err := cluster.Schedule(
		p.API,
		"ChaiTimeScheduler",
		cluster.MakeWaitForInterval(jobInterval),
		func() {
			if err := p.RunJob(); err != nil {
				p.API.LogError("Failed to Run job", "error", err.Error())
			}
		},
	)

	if err != nil {
		p.API.LogError(fmt.Sprintf("Unable to schedule job for standup reports. Error: {%s}", err.Error()))
		return err
	}

	p.job = job
	return nil
}

func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	p.API.LogError(args.Command)
	if args.Command == "/chai config" {
		return p.ExecuteCommandConfig(args.UserId, args.ChannelId, args.TriggerId)
	} else if args.Command == "/chai join" {
		return p.ExecuteCommandJoin(args.UserId, args.ChannelId)
	} else if args.Command == "/chai leave" {
		return p.ExecuteCommandLeave(args.UserId, args.ChannelId)
	} else if args.Command == "/chai test" {
		if err := p.RunJob(); err != nil {
			return model.CommandResponseFromPlainText(err.Error()), nil
			p.API.LogError("Failed to Run job", "error", err.Error())
		}

		return model.CommandResponseFromPlainText("Yo!"), nil
	}

	return nil, nil
}

func (p *Plugin) ensureBot() error {
	botID, err := p.client.Bot.EnsureBot(bot, pluginapi.ProfileImagePath("/assets/profile.gif"))
	if err != nil {
		p.API.LogError("Error occurred ensuring chaibot exists. Error: " + err.Error())
		return err
	}

	BotUserID = botID

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
			RoleID:  model.SystemUserRoleId,
			SubCommands: []*model.AutocompleteData{
				{
					Trigger: "config",
					RoleID:  model.SystemUserRoleId,
				},
				{
					Trigger: "join",
					RoleID:  model.SystemUserRoleId,
				},
				{
					Trigger: "leave",
					RoleID:  model.SystemUserRoleId,
				},
				{
					Trigger: "test",
					RoleID:  model.SystemUserRoleId,
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
		UserId:    request.UserId,
		Message:   "Chai Time config has been saved successfully.",
		ChannelId: request.ChannelId,
	})

	w.Write(body)
	w.WriteHeader(http.StatusOK)
}

// See https://developers.mattermost.com/extend/plugins/server/reference/
