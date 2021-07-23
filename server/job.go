package main

import (
	"fmt"
	"github.com/pkg/errors"
	"strconv"
	"strings"
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

	for channelId := range channels {
		channelConfig, err := p.chai.GetChannelConfig(channelId)
		if err != nil {
			// no need to log error here as it should be logged in called function.
			// Continue processing for other channels.
			continue
		}

		lastRun, err := p.getChannelLastRun(channelId)
		if err != nil {
			continue
		}

		if time.Now().Sub(*lastRun) < (time.Duration(channelConfig.Frequency) * 7 * 24 * time.Hour) {
			continue
		}

		if channelConfig.DayOfWeek != strings.ToLower(time.Now().Format("Monday")) {
			continue
		}

		pairings, err := p.chai.GetParing(channelId)
		if err != nil {
			continue
		}

		pairingPost, err := p.chai.GeneratePairingPost(p.getConfiguration().BotID, channelId, pairings)
		if err != nil {
			continue
		}

		if _, appErr :=p.API.CreatePost(pairingPost); appErr != nil {
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
