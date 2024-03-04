package commands

import (
	"soa-bot/services/scraping"
	"soa-bot/utils"

	"github.com/bwmarrin/discordgo"
	"github.com/cloudflare/cfssl/log"
	"github.com/sarulabs/di"
)

var WatchCommand = &discordgo.ApplicationCommand{
    Name:        "watch",
    Description: "Executes the `Watch()` function with provided URL and HTML element.",
    Options: []*discordgo.ApplicationCommandOption{
      {
        Type:        discordgo.ApplicationCommandOptionString,
        Name:        "url",
        Description: "The URL to watch.",
        Required:    true,
      },
	  {
        Type:        discordgo.ApplicationCommandOptionString,
        Name:        "html",
        Description: "The HTML to scrape",
        Required:    true,
      },
    },
	
}

func WatchHandler(s *discordgo.Session, i *discordgo.InteractionCreate, container di.Container){
	cmd_options := i.ApplicationCommandData().Options
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(cmd_options))
	for _, opt := range cmd_options {
		optionMap[opt.Name] = opt
	}

	url := optionMap["url"].Value.(string)
	html := optionMap["html"].Value.(string)

  scraperService := container.Get(utils.DiScraperService).(*scraping.ScraperService)
  scraperService.NewScrapeWorker(url, html)

  err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
    Type: discordgo.InteractionResponseChannelMessageWithSource,
    Data: &discordgo.InteractionResponseData{
      Content: "Scraping " + url + " for " + html,
    },
})
if err != nil {
  log.Error("Error responding to interaction:", err)
}
}

