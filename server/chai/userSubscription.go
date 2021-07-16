package chai

import (
	"encoding/json"
	"errors"
)

const (
	channelSubscriptionsKeyPrefix = "channel_"
)

func (c *Chai) SaveUserSubscription(userID, channelID string) error {
	c.userSubscriptionLock.Lock()
	defer c.userSubscriptionLock.Unlock()

	locations, err := c.GetLocations()
	if err != nil {
		return err
	}

	found := false
	for location := range locations {
		if location == channelID {
			found = true
			break
		}
	}

	if !found {
		return errors.New("Channel is not part of Chai Time. Make sure the channel you're trying to join has Chat Time enabled.")
	}

	channelSubscription, err := c.GetChannelSubscriptions(channelID)
	if err != nil {
		return err
	}

	channelSubscription[userID] = true
	return c.SaveChannelSubscriptions(channelID, channelSubscription)
}

func (c *Chai) GetChannelSubscriptions(channelID string) (map[string]bool, error) {
	data, appErr := c.API.KVGet(channelSubscriptionsKeyPrefix + channelID)
	if appErr != nil {
		c.API.LogError("Error occurred fetching channel subscription data from KV store.", "channelID", channelID, "error", appErr.Error())
		return nil, errors.New(appErr.Error())
	}

	var channelSubscription map[string]bool
	if err := json.Unmarshal(data, &channelSubscription); err != nil {
		c.API.LogError("Error occurred unmarshalling channel subscription data.", "channelID", channelID, "error", err.Error())
		return nil, err
	}

	return channelSubscription, nil
}

func (c *Chai) SaveChannelSubscriptions(channelID string, subscriptions map[string]bool) error {
	data, err := json.Marshal(subscriptions)
	if err != nil {
		c.API.LogError("Error occurred marshalling channel subscription data.", "channelID", channelID, "error", err.Error())
		return err
	}

	if appErr := c.API.KVSet(channelSubscriptionsKeyPrefix + channelID, data); appErr != nil {
		c.API.LogError("Error occurred saving channel subscription data from KV store.", "channelID", channelID, "error", appErr.Error())
		return errors.New(appErr.Error())
	}

	return nil
}
