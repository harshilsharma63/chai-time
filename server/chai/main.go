package chai

import (
	"github.com/mattermost/mattermost-server/v5/plugin"
	"sync"
)

type Chai struct {
	API plugin.API

	locationsLock        sync.RWMutex
	configLock           sync.RWMutex
	userSubscriptionLock sync.RWMutex
}
