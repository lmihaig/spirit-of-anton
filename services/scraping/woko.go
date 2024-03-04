package scraping

import (
	"net/http"
	"soa-bot/services/database"
	"soa-bot/services/notify"
	"soa-bot/utils"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/cloudflare/cfssl/log"
	"github.com/sarulabs/di"
)

type WokoService struct {
	container di.Container
}

func InitWokoService(container di.Container) (*WokoService, error) {
	wokoService := WokoService{container: container}

	wokoService.RunWokoService()

	return &wokoService, nil
}

func (woko *WokoService) RunWokoService (){
	go func() {
		db := woko.container.Get(utils.DiDatabase).(*database.MongoDB)
		notifier := woko.container.Get(utils.DiNotifierService).(*notify.Notifier)
		for {
			scrapedData, err := scrapeWoko()
			if err != nil {
				log.Error("Error scraping", err)
			}
			for _, data := range scrapedData {
				item, err := db.GetItemByFilter("woko", "woko", "roomID", data.RoomID)
				if err != nil {
					log.Error("Error getting item by filter", err)
				}
				if item == nil {
					err := db.InsertItem("woko", "woko", data)
					if err != nil {
						log.Error("Error inserting item", err)
					}
					err = notifier.Notify("watch", data.ToString())
					if err != nil {
						log.Error("Error notifying", err)
					}
				}
			}
			time.Sleep(60 * time.Minute)
		}
		
	}()
}

type WokoData struct {
	RoomID string `bson:"roomID"`
	Title string `bson:"title"`
	Tenant bool `bson:"tenant"`
	Date string `bson:"date"`
	Price string `bson:"price"`
	Address	string `bson:"address"`
	Link string `bson:"link"`
}

func (data WokoData) ToString() string {
	return "Title: " + data.Title + "\n" +
		"Tenant: " + strconv.FormatBool(data.Tenant) + "\n" +
		"Date: " + data.Date + "\n" +
		"Price: " + data.Price + "\n" +
		"Address: " + data.Address + "\n" +
		"Link: " + data.Link

}

func scrapeWoko() ([]WokoData, error) {
	resp, err := http.Get("https://www.woko.ch/en/zimmer-in-zuerich")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var roomList []WokoData
	doc.Find(".inserat").Each(func(i int, s *goquery.Selection) {
        link, _ := s.Find("a").Attr("href")
        link = "https://www.woko.ch" + link

		roomID := link[strings.LastIndex(link, "/")+1:]
		title := s.Find(".titel h3").Text()
		tenantText := s.Find("table tr:nth-child(1) td:nth-child(1)").Text()
		tenant := strings.Contains(strings.ToLower(tenantText), "tenant")
		date := s.Find(".titel span").Text()
		price := s.Find(".preis").Text()
		address := s.Find("table tr:nth-child(2) td:nth-child(2)").Text()

		room := WokoData{
			RoomID:  roomID,
			Title:   title,
			Tenant:  tenant,
			Date:    date,
			Price:   price,
			Address: address,
			Link:    link,
		}

		roomList = append(roomList, room)
	})

	return roomList, nil
}