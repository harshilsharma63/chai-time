package chai

import "github.com/mattermost/mattermost-server/v5/model"

func (c *Chai) OpenConfigDialog(triggerID string) error {
	c.API.OpenInteractiveDialog(model.OpenDialogRequest{
		TriggerId: triggerID,
		URL: "/plugins/com.harshilsharma63.chai-time/save_config",
		Dialog: model.Dialog{
			CallbackId: "chaiTimeSaveConfig",
			Title: "title",
			IntroductionText: "introduction",
			Elements: []model.DialogElement{
				{
					DisplayName: "Every",
					Name: "every",
					Type:
				},
			},
		},
	})
}
