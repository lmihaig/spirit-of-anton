package scraping

import (
	"net/http"
	"soa-bot/services/database"
	"soa-bot/services/notify"
	"soa-bot/utils"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/cloudflare/cfssl/log"
	"github.com/sarulabs/di"
)


type ScrapedData struct {
	Time time.Time `bson:"time"`
	Data string `bson:"data"`
	Html string `bson:"html"`
}

type ScraperService struct {
	container di.Container
}

func InitScraperService(container di.Container) (*ScraperService, error) {
	db := container.Get(utils.DiDatabase).(*database.MongoDB)
	collections, err := db.GetCollectionsList("watch")
	if err != nil {
		log.Error("Error getting collections list", err)
		return nil, err
	}

	ss := ScraperService{container: container}

	for _, collection := range collections {

		lastItem, err := db.GetLatestItem("watch", collection)
		if lastItem == nil {
			continue
		}
		if err != nil {
			log.Error("Error getting latest item", err)
		}

		ss.NewScrapeWorker(collection, lastItem["html"].(string))
	}

	return &ss, nil
}


// database watch
// collection <URL>
// document ScraptedData 
func (s *ScraperService) NewScrapeWorker(url string, html string) {
	go func() {
		db := s.container.Get(utils.DiDatabase).(*database.MongoDB)
		notifier := s.container.Get(utils.DiNotifierService).(*notify.Notifier)
		log.Info("Scraping ", url, "for ", html)
		 for{ 
			scrapedData, err := scrape(url, html)
			if err != nil {
				log.Error("Error scraping", err)
			}
			lastItem, err := db.GetLatestItem("watch", url)
			if err != nil {
				log.Error("Error getting latest item", err)
			}
			if lastItem == nil || lastItem["data"] != scrapedData.Data {
				err = db.InsertItem("watch", url, scrapedData)
				if err != nil {
					log.Error("Error inserting item", err)
				}
				err = notifier.Notify("watch", scrapedData.Data)
				if err != nil {
					log.Error("Error notifying", err)
				}
			}
			
			time.Sleep(60 * time.Minute)
		 }
	}()
}


func scrape(url string, html string) (ScrapedData, error) {
	resp, err := http.Get(url)
	if err != nil {
		return ScrapedData{}, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return ScrapedData{}, err
	}

	scrapedData := doc.Find(html).First().Text()
	currentTime := time.Now()

	return ScrapedData{
		Time: currentTime,
		Data: scrapedData,
		Html: html,
	}, nil
}