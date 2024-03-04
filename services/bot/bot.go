package bot

import (
	"fmt"
	"log"
	"soa-bot/commands"
	"soa-bot/utils"

	"github.com/bwmarrin/discordgo"
	"github.com/sarulabs/di"
)



type CommandInfo struct {
    Handler func(s *discordgo.Session, i *discordgo.InteractionCreate, container di.Container)
    Command *discordgo.ApplicationCommand
}

var commandsMap = map[string]CommandInfo{
    commands.PingCommand.Name: {Handler: commands.PingHandler, Command: commands.PingCommand},
    commands.WatchCommand.Name: {Handler: commands.WatchHandler, Command: commands.WatchCommand},
}


func InitDiscordBot(container di.Container) (release func()){
	release = func() {}

	session := container.Get(utils.DiDiscordSession).(*discordgo.Session)
	cfg := container.Get(utils.DiConfig).(*utils.Config)

	session.Token = "Bot " + cfg.Discord_token
    session.StateEnabled = false


	session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
        if command, ok := commandsMap[i.ApplicationCommandData().Name]; ok {
            command.Handler(s, i, container)
        } else {
            fmt.Println("Unknown command received:", i.ApplicationCommandData().Name)
        }
    })


    err := session.Open()
    if err != nil {
        fmt.Println("error opening connection,", err)
        return
    }

    // session.ApplicationCommandBulkOverwrite("1212330926567329844", " ", commandList)
	for _, v := range commandsMap {
		_, err := session.ApplicationCommandCreate(session.State.User.ID, "", v.Command)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Command.Name, err)
		}
	}

	return
}
