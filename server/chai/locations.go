package chai

import "encoding/json"

const (
	locationsKey = "locations"
)

func (c *Chai) GetLocations() (map[string]bool, error) {
	c.locationsLock.Lock()
	defer c.locationsLock.Unlock()

	data, err := c.API.KVGet(locationsKey)
	if err != nil {
		c.API.LogError("Error occurred fetching enabled locations. Error: " + err.Error())
		return nil, err
	}

	var locations map[string]bool
	if err := json.Unmarshal(data, &locations); err != nil {
		c.API.LogError("Error occurred unmarshalling enabled locations data. Error: " + err.Error())
		return nil, err
	}

	return locations, nil
}

func (c *Chai) SaveLocations(locations map[string]bool) error {
	c.locationsLock.Lock()
	defer c.locationsLock.Unlock()

	data, err := json.Marshal(locations)
	if err != nil {
		c.API.LogError("Error occurred marshaling locations data for saving. Error: " + err.Error())
		return err
	}

	return c.API.KVSet(locationsKey, data)
}

func (c *Chai) AddLocation(location string) error {
	c.locationsLock.Lock()
	defer c.locationsLock.Unlock()

	locations, err := c.GetLocations()
	if err != nil {
		return err
	}

	locations[location] = true
	return c.SaveLocations(locations)
}
