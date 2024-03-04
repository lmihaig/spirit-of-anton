package main

import (
	"fmt"
	"os"
	"os/signal"
	"soa-bot/services/bot"
	"soa-bot/services/database"
	"soa-bot/services/notify"
	"soa-bot/services/scraping"
	"soa-bot/utils"

	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/cloudflare/cfssl/log"
	"github.com/sarulabs/di"
)

func main() {
    diBuilder, _ := di.NewBuilder()

    diBuilder.Add(di.Def{
		Name: utils.DiConfig,
		Build: func(ctn di.Container) (interface{}, error) {
			return utils.InitConfig(), nil
		},
	})

    diBuilder.Add(di.Def{
        Name: utils.DiDiscordSession,
        Build: func(ctn di.Container) (interface{}, error) {
			return discordgo.New("")
		},
        Close: func(obj interface{}) error {
			session := obj.(*discordgo.Session)
			log.Info("Shutting down bot session...")
			session.Close()
			return nil
        },
    })

    diBuilder.Add(di.Def{
		Name: utils.DiDatabase,
		Build: func(ctn di.Container) (interface{}, error) {
			return database.InitDatabase(ctn)
		},
		Close: func(obj interface{}) error {
			database := obj.(*database.MongoDB)
			log.Info("Shutting down database connection...")
			database.Close()
			return nil
		},
	})


    diBuilder.Add(di.Def{
        Name: utils.DiNotifierService,
        Build: func(ctn di.Container) (interface{}, error) {
            return notify.InitNotifierService(ctn)
        },
    })

    diBuilder.Add(di.Def{
        Name: utils.DiWokoService,
        Build: func(ctn di.Container) (interface{}, error) {
            return scraping.InitWokoService(ctn)
        },
    })

    diBuilder.Add(di.Def{
        Name: utils.DiScraperService,
        Build: func(ctn di.Container) (interface{}, error) {
            return scraping.InitScraperService(ctn)
        },
    })

    ctn := diBuilder.Build()
    defer ctn.DeleteWithSubContainers()

    ctn.Get(utils.DiDiscordSession)
    release := bot.InitDiscordBot(ctn)
    defer release()

    ctn.Get(utils.DiScraperService)
    ctn.Get(utils.DiWokoService)

    fmt.Println("Started event loop.  Press CTRL-C to exit.")
    sc := make(chan os.Signal, 1)
    signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
    <-sc
}