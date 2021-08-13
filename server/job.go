package main

import (
	"fmt"
	"github.com/pkg/errors"
	"strconv"
	"time"
)

const (
	channelLastRunKeyPrefix = "last_run_"
)

func (p *Plugin) RunJob() error {
	channels, err := p.chai.GetEnabledChannels()
	if err != nil {
		return err
	}

	p.API.LogError(fmt.Sprintf("%v", channels))

	for channelId := range channels {
		p.API.LogDebug("Processing channel " + channelId)
		channelConfig, err := p.chai.GetChannelConfig(channelId)
		if err != nil {
			// no need to log error here as it should be logged in called function.
			// Continue processing for other channels.
			continue
		}

		p.API.LogError(fmt.Sprintf("%v", channelConfig))

		lastRun, err := p.getChannelLastRun(channelId)
		if err != nil {
			continue
		}

		if time.Now().Sub(*lastRun) >= (time.Hour * 24 * 7 * time.Duration(channelConfig.Frequency)) {
			p.API.LogDebug("Skipping pairings for channel " + channelId)
			continue
		}

		pairings, err := p.chai.GetParing(channelId)
		if err != nil {
			continue
		}

		config := p.getConfiguration()
		pairingPost, err := p.chai.GeneratePairingPost(BotUserID, channelId, pairings, config.HeaderMessage)
		if err != nil {
			continue
		}

		if _, appErr := p.API.CreatePost(pairingPost); appErr != nil {
			p.API.LogError("Error occurred creating pairing post for channel", "channelID", channelId, "error", appErr.Error())
			continue
		}

		if err := p.saveChannelLastRun(channelId, time.Now()); err != nil {
			continue
		}
	}

	return nil
}

func (p *Plugin) getChannelLastRun(channelID string) (*time.Time, error) {
	data, appErr := p.API.KVGet(channelLastRunKeyPrefix + channelID)
	if appErr != nil {
		p.API.LogError("Error occurred fetching channel's last run time from KV store.", "channelID", channelID, "error", appErr.Error())
		return nil, errors.New(appErr.Error())
	}

	if data == nil || len(data) == 0 {
		data = []byte("0")
	}

	seconds, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		p.API.LogError("Error occurred converting seconds data from string to int.", "data", string(data), "error", err.Error())
		return nil, err
	}

	lastRun := time.Unix(seconds, 0)
	return &lastRun, nil
}

func (p *Plugin) saveChannelLastRun(channelID string, run time.Time) error {
	data := fmt.Sprintf("%d", run.Unix())
	if appErr := p.API.KVSet(channelLastRunKeyPrefix+channelID, []byte(data)); appErr != nil {
		p.API.LogError("Error occurred saving channel last run to KV store", "channelID", channelID, "error", appErr.Error())
		return errors.New(appErr.Error())
	}

	return nil
}
