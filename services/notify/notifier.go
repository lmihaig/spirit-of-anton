package notify

import (
	"soa-bot/utils"

	"github.com/bwmarrin/discordgo"
	"github.com/sarulabs/di"
)

type Notifier struct {
	session *discordgo.Session
}

func InitNotifierService(container di.Container) (*Notifier, error) {
	session := container.Get(utils.DiDiscordSession).(*discordgo.Session)
	return &Notifier{session}, nil
}

func (n *Notifier) Notify(channelID string, message string) error {
	if channelID == "watch" {
		channelID = "1212468848972800011"
	}
	if len(message) > 2000 {
		message = message[:2000]
	}
	_, err := n.session.ChannelMessageSend(channelID, message)
	return err
}