package chai

import (
	"encoding/json"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/pkg/errors"
	"strconv"
)

const (
	configKeyPrefix = "config_"
)

type Config struct {
	ChannelID string
	Frequency int
	DayOfWeek string
}

func (c *Chai) OpenConfigDialog(channelID, triggerID string) error {
	config, err := c.GetConfig(channelID)
	if err != nil {
		return err
	}

	// TODO set default value here
	dialog := model.OpenDialogRequest{
		TriggerId: triggerID,
		URL:       "/plugins/chaitime/saveConfig",
		Dialog: model.Dialog{
			CallbackId:       "chaiTimeSaveConfig",
			Title:            "title",
			IntroductionText: "introduction",
			Elements: []model.DialogElement{
				{
					DisplayName: "Every",
					Name:        "frequency",
					Type:        "select",
					Options: []*model.PostActionOptions{
						{
							Text:  "Every week",
							Value: "1",
						},
						{
							Text:  "Every 2 week",
							Value: "2",
						},
						{
							Text:  "Every 3 weeks",
							Value: "3",
						},
						{
							Text:  "Every 4 weeks",
							Value: "4",
						},
					},
					Default: "4",
				},
				{
					DisplayName: "On",
					Name:        "dayOfWeek",
					Type:        "select",
					Options: []*model.PostActionOptions{
						{
							Text:  "Sunday",
							Value: "sunday",
						},
						{
							Text:  "Monday",
							Value: "monday",
						},
						{
							Text:  "Tuesday",
							Value: "tuesday",
						},
						{
							Text:  "Wednesday",
							Value: "wednesday",
						},
						{
							Text:  "Thursday",
							Value: "thursday",
						},
						{
							Text:  "Friday",
							Value: "friday",
						},
						{
							Text:  "Saturday",
							Value: "saturday",
						},
					},
					Default: "monday",
				},
			},
		},
	}

	// fill default values if available
	if config != nil {
		dialog.Dialog.Elements[0].Default = strconv.Itoa(config.Frequency)
		dialog.Dialog.Elements[1].Default = config.DayOfWeek
	}

	if appErr := c.API.OpenInteractiveDialog(dialog); appErr != nil {
		return errors.New(appErr.Error())
	}

	return nil
}

func (c *Chai) SaveConfig(config Config) error {
	c.configLock.Lock()
	defer c.configLock.Unlock()

	data, err := json.Marshal(config)
	if err != nil {
		c.API.LogError("Error occurred marshaling channel config object. Error: " + err.Error())
		return err
	}

	if appErr := c.API.KVSet(configKeyPrefix + config.ChannelID, data); appErr != nil {
		c.API.LogError("Error occurred saving channel config to KV store.", "channelID", config.ChannelID, "error", appErr.Error())
		return err
	}

	return nil
}

func (c *Chai) GetConfig(channelID string) (*Config, error) {
	data, appErr := c.API.KVGet(configKeyPrefix + channelID)
	if appErr != nil {
		c.API.LogError("Error occurred fetching channel config from KV store.", "channelID", channelID, "error", appErr.Error())
		return nil, errors.New(appErr.Error())
	}

	if len(data) == 0 {
		return nil, nil
	}

	var config *Config
	if err := json.Unmarshal(data, &config); err != nil {
		c.API.LogError("Error occurred unmarshalling channel config.", "channelID", channelID, "error", err.Error())
		return nil, err
	}

	return config, nil
}
