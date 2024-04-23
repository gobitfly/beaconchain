package utils

import (
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gtuk/discordwebhook"
)

// Send message to discord
func SendMessage(content string, config *types.InternalAlertDiscord) {
	if len(config.DiscordWebhookUrl) > 0 {
		err := discordwebhook.SendMessage(config.DiscordWebhookUrl, discordwebhook.Message{Username: &config.DiscordUserName, Content: &content, AvatarUrl: &config.AvatarURL})
		if err != nil {
			log.Error(err, "error sending message to discord", 0, map[string]interface{}{"content": content, "webhookUrl": config.DiscordWebhookUrl, "username": config.DiscordUserName})
		}
	}
}
