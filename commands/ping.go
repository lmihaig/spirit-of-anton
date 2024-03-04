package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/cloudflare/cfssl/log"
	"github.com/sarulabs/di"
)

var PingCommand = &discordgo.ApplicationCommand{
    Name:        "ping",
    Description: "Checks if the bot is online.",
}

func PingHandler(s *discordgo.Session, i *discordgo.InteractionCreate, container di.Container) {
    err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
      Type: discordgo.InteractionResponseChannelMessageWithSource,
      Data: &discordgo.InteractionResponseData{
        Content: "Pong!",
      },
    })
    if err != nil {
      log.Error("Error responding to interaction:", err)
    }
}
