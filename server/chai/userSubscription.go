package chai

import (
	"encoding/json"
	"errors"
)

const (
	channelSubscriptionsKeyPrefix = "channel_"
)

func (c *Chai) AddChannelMember(userID, channelID string) error {
	c.userSubscriptionLock.Lock()
	defer c.userSubscriptionLock.Unlock()

	channelSubscription, err := c.getChannelMembers(channelID)
	if err != nil {
		return err
	}

	channelSubscription[userID] = true
	return c.saveChannelMembers(channelID, channelSubscription)
}

func (c *Chai) RemoveChannelMember(userID, channelID string) error {
	c.userSubscriptionLock.Lock()
	defer c.userSubscriptionLock.Unlock()

	channelSubscription, err := c.getChannelMembers(channelID)
	if err != nil {
		return err
	}

	delete(channelSubscription, userID)
	return c.saveChannelMembers(channelID, channelSubscription)
}

func (c *Chai) saveChannelMembers(channelID string, subscriptions map[string]bool) error {
	data, err := json.Marshal(subscriptions)
	if err != nil {
		c.API.LogError("Error occurred marshalling channel subscription data.", "channelID", channelID, "error", err.Error())
		return err
	}

	if appErr := c.API.KVSet(channelSubscriptionsKeyPrefix+channelID, data); appErr != nil {
		c.API.LogError("Error occurred saving channel subscription data from KV store.", "channelID", channelID, "error", appErr.Error())
		return errors.New(appErr.Error())
	}

	return nil
}

func (c *Chai) getChannelMembers(channelID string) (map[string]bool, error) {
	data, appErr := c.API.KVGet(channelSubscriptionsKeyPrefix + channelID)
	if appErr != nil {
		c.API.LogError("Error occurred fetching channel members data from KV store.", "channelID", channelID, "error", appErr.Error())
		return nil, errors.New(appErr.Error())
	}

	if data == nil || len(data) == 0 {
		data = []byte("{}")
	}

	var channelMembers map[string]bool
	if err := json.Unmarshal(data, &channelMembers); err != nil {
		c.API.LogError("Error occurred unmarshalling channel members data.", "channelID", channelID, "error", err.Error())
		return nil, err
	}

	return channelMembers, nil
}

func (c *Chai) GetChannelConfig(channelID string) (*Config, error) {
	data, appErr := c.API.KVGet(channelSubscriptionsKeyPrefix + channelID)
	if appErr != nil {
		c.API.LogError("Error occurred fetching channel subscription data from KV store.", "channelID", channelID, "error", appErr.Error())
		return nil, errors.New(appErr.Error())
	}

	var channelSubscription *Config
	if err := json.Unmarshal(data, &channelSubscription); err != nil {
		c.API.LogError("Error occurred unmarshalling channel subscription data.", "channelID", channelID, "error", err.Error())
		return nil, err
	}

	return channelSubscription, nil
}
