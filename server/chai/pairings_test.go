package chai

import (
	"encoding/json"
	"github.com/mattermost/mattermost-server/v6/plugin/plugintest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestChai_getParing(t *testing.T) {
	channelID := "channel_1"

	mockAPI := &plugintest.API{}

	chai := &Chai{
		API: mockAPI,
	}

	members := map[string]bool{
		"userid_1": true,
		"userid_2": true,
		"userid_3": true,
		"userid_4": true,
		"userid_5": true,
		"userid_6": true,
		"userid_7": true,
	}
	membersJson, _ := json.Marshal(members)
	mockAPI.On("KVGet", "channel_channel_1").Return(membersJson, nil)

	history := channelHistory{
		"userid_1": {
			"userid_2": 2,
			"userid_5": 1,
		},
		"userid_2": {
			"userid_1": 2,
			"userid_3": 5,
		},
		"userid_3": {
			"userid_2": 5,
		},
		"userid_5": {
			"userid_1": 3,
			"userid_7": 10,
		},
		"userid_7": {
			"userid_5": 10,
		},
	}
	historyJson, _ := json.Marshal(history)
	mockAPI.On("KVGet", "history_channel_1").Return(historyJson, nil)
	mockAPI.On("KVSet", "history_channel_1", mock.Anything).Return(nil)

	pairing, err := chai.GetParing(channelID)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(pairing))
}
