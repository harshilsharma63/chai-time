package chai

import (
	"encoding/json"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/pkg/errors"
	"golang.org/x/exp/rand"
	"math"
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

		minPairings := math.MaxInt64
		minPairingMembers := []string{}

		for otherMember := range history {
			if member == otherMember || processedMembers[otherMember] == true {
				continue
			}

			numPairings := memberHistory[otherMember]
			if numPairings > minPairings {
				continue
			} else if numPairings == minPairings {
				minPairingMembers = append(minPairingMembers, otherMember)
			} else {
				minPairings = numPairings
				minPairingMembers = []string{otherMember}
			}

		}

		rand.Seed(uint64(time.Now().UnixNano()))

		if numMembersLeft == 3 {
			processedMembers[member] = true
			processedMembers[minPairingMembers[0]] = true
			processedMembers[minPairingMembers[1]] = true

			pairings = append(pairings, []string{
				member,
				minPairingMembers[0],
				minPairingMembers[1],
			})

			history[member][minPairingMembers[0]] = history[member][minPairingMembers[0]] + 1
			history[member][minPairingMembers[1]] = history[member][minPairingMembers[1]] + 1

			history[minPairingMembers[0]][member] = history[minPairingMembers[0]][member] + 1
			history[minPairingMembers[0]][minPairingMembers[1]] = history[minPairingMembers[0]][minPairingMembers[1]] + 1

			history[minPairingMembers[1]][member] = history[minPairingMembers[1]][member] + 1
			history[minPairingMembers[1]][minPairingMembers[0]] = history[minPairingMembers[1]][minPairingMembers[0]] + 1

			numMembersLeft -= 3
		} else {
			processedMembers[member] = true
			pairedMember := minPairingMembers[rand.Intn(len(minPairingMembers))]
			processedMembers[pairedMember] = true

			pairings = append(pairings, []string{
				member,
				pairedMember,
			})

			history[member][pairedMember] = history[member][pairedMember] + 1
			history[pairedMember][member] = history[pairedMember][member] + 1

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

func (c *Chai) GeneratePairingPost(authorID, channelId string, pairing [][]string) (*model.Post, error) {
	post := &model.Post{
		UserId:    authorID,
		ChannelId: channelId,
		Message: "#### :coffee: New remote coffee time pairings for July 8 - 22\n" +
			"Don't forget to reach out to your coffee buddy to arrange for a day/time to meet! :grin: \n" +
			"**Please note pairings are randomized**, if your pair is repeated you can talk to each other again or pass for this round :blue_heart:\n\n",
	}

	pairingTable := make([]string, len(pairing)+1)
	pairingTable[1] = "|---|---|"

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
		if i == 0 {
			pairingTable[i] = rowString
		} else {
			pairingTable[i+1] = rowString
		}
	}

	post.Message += strings.Join(pairingTable, "\n")
	return post, nil
}
