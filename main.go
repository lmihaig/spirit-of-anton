package main

import (
	"bot/commands"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)



var dg *discordgo.Session
var GuildID = ""
var watchService *commands.WatchService

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file:", err)
	}
	Token := os.Getenv("DISCORD_TOKEN")
	
	if Token == "" {
		log.Fatal("DISCORD_TOKEN not found in .env file")
	}

	dg, err = discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	watchService, err = commands.NewWatchService(dg)
	if err != nil {
		log.Fatal("Error starting Watch Service:", err)
	}
}


var commandDescriptions = []*discordgo.ApplicationCommand{
  {
    Name:        "ping",
    Description: "Checks if the bot is online.",
  },
  {
    Name:        "watch",
    Description: "Executes the `Watch()` function with provided URL and time.",
    Options: []*discordgo.ApplicationCommandOption{
      {
        Type:        discordgo.ApplicationCommandOptionString,
        Name:        "url",
        Description: "The URL to watch.",
        Required:    true,
      },
    },
  },
}

var commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
  "ping": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
    err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
      Type: discordgo.InteractionResponseChannelMessageWithSource,
      Data: &discordgo.InteractionResponseData{
        Content: "Pong!",
      },
    })
    if err != nil {
      fmt.Println("Error responding to interaction:", err)
    }
  },

  "watch": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	cmd_options := i.ApplicationCommandData().Options
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(cmd_options))
	for _, opt := range cmd_options {
		optionMap[opt.Name] = opt
	}

	url := optionMap["url"].Value.(string)
	watchService.Add(url, i.GuildID)

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
      Type: discordgo.InteractionResponseChannelMessageWithSource,
      Data: &discordgo.InteractionResponseData{
        Content: "Added succesfully",
      },
    })

    if err != nil {
      fmt.Println("Error responding to interaction:", err)
    }
  },
  }

func init() {
	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})
}

func main() {
	err := dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	log.Println("Adding commands...")
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commandDescriptions))
	for i, v := range commandDescriptions {
		cmd, err := dg.ApplicationCommandCreate(dg.State.User.ID, GuildID, v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}

	go watchService.Run()

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	defer dg.Close()
}
