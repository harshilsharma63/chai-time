package chai

import (
	"encoding/json"
	"fmt"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/pkg/errors"
	"golang.org/x/exp/rand"
	"sort"
	"strings"
	"time"
)

const (
	channelHistoryKeyPrefix = "history_"
)

type channelHistory = map[string]map[string]int

func (c *Chai) GetParing(channelID string) ([][]string, error) {
	members, err := c.getChannelMembers(channelID)
	if err != nil {
		return nil, err
	}

	//memberList := c.getMemberList(members)
	//processedMembers := c.getProcessedMembersTemplate(members)
	processedMembers := map[string]bool{}
	numMembersLeft := len(members)

	history, err := c.getChannelHistory(channelID)
	if err != nil {
		return nil, err
	}

	history = c.hydrateHistory(members, history)

	pairings := [][]string{}

	for member := range history {
		if processedMembers[member] == true {
			continue
		}

		memberHistory := history[member]

		minPairings := map[int]bool{}
		minPairingMembers := map[int][]string{}

		for otherMember := range history {
			if member == otherMember || processedMembers[otherMember] == true {
				continue
			}

			numPairings := memberHistory[otherMember]

			minPairings[numPairings] = true
			minPairingMembers[numPairings] = append(minPairingMembers[numPairings], otherMember)

			//if numPairings > minPairings {
			//	continue
			//} else if numPairings == minPairings {
			//	minPairingMembers = append(minPairingMembers, otherMember)
			//} else {
			//	minPairings = numPairings
			//	minPairingMembers = []string{otherMember}
			//}
		}

		var sortedMinPairingCounts []int
		for key := range minPairings {
			sortedMinPairingCounts = append(sortedMinPairingCounts, key)
		}
		sort.Ints(sortedMinPairingCounts)

		rand.Seed(uint64(time.Now().UnixNano()))

		if numMembersLeft == 3 {
			length := len(minPairingMembers[sortedMinPairingCounts[0]])
			index := rand.Intn(length)
			memberA := minPairingMembers[sortedMinPairingCounts[0]][index]

			// remove selected member from array
			minPairingMembers[sortedMinPairingCounts[0]][index] = minPairingMembers[sortedMinPairingCounts[0]][length-1]
			minPairingMembers[sortedMinPairingCounts[0]] = minPairingMembers[sortedMinPairingCounts[0]][:length-1]

			x := 0
			if len(minPairingMembers[sortedMinPairingCounts[0]]) == 0 {
				x = 1
			}

			memberB := minPairingMembers[sortedMinPairingCounts[x]][rand.Intn(len(minPairingMembers[sortedMinPairingCounts[x]]))]

			processedMembers[member] = true
			processedMembers[memberA] = true
			processedMembers[memberB] = true

			pairings = append(pairings, []string{
				member,
				memberA,
				memberB,
			})

			history[member][memberA] = history[member][memberA] + 1
			history[member][memberB] = history[member][memberB] + 1

			history[memberA][member] = history[memberA][member] + 1
			history[memberA][memberB] = history[memberA][memberB] + 1

			history[memberB][member] = history[memberB][member] + 1
			history[memberB][memberA] = history[memberB][memberA] + 1

			numMembersLeft -= 3
		} else {
			otherMember := minPairingMembers[sortedMinPairingCounts[0]][rand.Intn(len(minPairingMembers[sortedMinPairingCounts[0]]))]

			processedMembers[member] = true
			processedMembers[otherMember] = true

			pairings = append(pairings, []string{
				member,
				otherMember,
			})

			history[member][otherMember] = history[member][otherMember] + 1
			history[otherMember][member] = history[otherMember][member] + 1

			numMembersLeft -= 2
		}
	}

	if err := c.saveChannelHistory(channelID, history); err != nil {
		return nil, err
	}

	return pairings, nil
}

func (c *Chai) getChannelHistory(channelID string) (channelHistory, error) {
	data, appErr := c.API.KVGet(channelHistoryKeyPrefix + channelID)
	if appErr != nil {
		c.API.LogError("Error occurred fetching channel history from KV store.", "channelID", channelID, "error", appErr.Error())
		return nil, errors.New(appErr.Error())
	}

	if data == nil || len(data) == 0 {
		data = []byte("{}")
	}

	var history map[string]map[string]int
	if err := json.Unmarshal(data, &history); err != nil {
		c.API.LogError("Error occurred unmarshalling channel history data.", "channelID", channelID, "error", err.Error())
		return nil, err
	}

	return history, nil
}

func (c *Chai) saveChannelHistory(channelID string, history channelHistory) error {
	data, err := json.Marshal(history)
	if err != nil {
		c.API.LogError("Error occurred marshalling channel history data.", "channelID", channelID, "error", err.Error())
		return err
	}

	if appErr := c.API.KVSet(channelHistoryKeyPrefix+channelID, data); appErr != nil {
		c.API.LogError("Error occurred saving channel history to KV store.", "channelID", channelID, "error", appErr.Error())
		return errors.New(appErr.Error())
	}

	return nil
}

func (c *Chai) hydrateHistory(members map[string]bool, history channelHistory) channelHistory {
	for member := range members {
		if history[member] == nil {
			history[member] = map[string]int{}
		}
	}

	return history
}

func (c *Chai) getMemberList(members map[string]bool) []string {
	membersList := make([]string, len(members))
	i := 0

	for member := range members {
		membersList[i] = member
		i++
	}

	return membersList
}

func (c *Chai) getProcessedMembersTemplate(members map[string]bool) map[string]bool {
	template := map[string]bool{}
	for member := range members {
		template[member] = false
	}

	return template
}

func (c *Chai) GeneratePairingPost(authorID, channelId string, pairing [][]string, headerMessage string) (*model.Post, error) {
	channelConfig, err := c.GetConfig(channelId)
	if err != nil {
		return nil, err
	}

	startDate := time.Now()
	// add channelConfig.Frequency number of weeks to start date
	endDate := startDate.Add(time.Hour * 24 * 7 * time.Duration(channelConfig.Frequency))

	message := fmt.Sprintf("#### :coffee: New remote coffee time pairings for %s - %s\n", startDate.Format("January 2"), endDate.Format("January 2"))
	if len(headerMessage) > 0 {
		message += headerMessage + "\n\n"
	}

	post := &model.Post{
		UserId:    authorID,
		ChannelId: channelId,
		Message:   message,
	}

	pairingTable := make([]string, len(pairing)+2)
	pairingTable[0] = "|Chai Time|"
	pairingTable[1] = "|---|---|---|"

	for i, pairing := range pairing {
		row := []string{}
		for _, userId := range pairing {
			user, appErr := c.API.GetUser(userId)
			if appErr != nil {
				c.API.LogError("Error occurred fetching user while generating pairing post.", "channelId", channelId, "userId", userId, "error", appErr.Error())
				return nil, errors.New(appErr.Error())
			}

			row = append(row, "@"+user.Username)
		}

		rowString := "|" + strings.Join(row, "|") + "|"
		pairingTable[i+2] = rowString
	}

	post.Message += strings.Join(pairingTable, "\n")
	return post, nil
}
