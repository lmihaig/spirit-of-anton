package commands

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type WatchService struct {
	db *mongo.Database
	dg *discordgo.Session
}

type ScrapedData struct {
	Time time.Time `bson:"time"`
	Data string    `bson:"data"`
}

func NewWatchService(dg *discordgo.Session) (*WatchService, error) {
	const (
		uri = "mongodb://localhost:27017/"
	)
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)

	dbClient, err := mongo.Connect(context.Background(), opts)
	if err != nil {
		return nil, err
	}

	var result bson.M
	if err := dbClient.Database("admin").RunCommand(context.Background(), bson.D{{Key: "ping", Value: 1}}).Decode(&result); err != nil {
		return nil, err
	}
	fmt.Println("Successfully connected to MongoDB!")

	return &WatchService{db: dbClient.Database("watch"), dg: dg}, nil
}

func (ws *WatchService) Add(url string) {
	if !ws.collectionExists(url) {
		ws.db.CreateCollection(context.Background(), url)
	}
	ws.updateCollection(url)
}

func (ws *WatchService) collectionExists(collectionName string) bool {
	collections, err := ws.db.ListCollectionNames(context.Background(), bson.D{})
	if err != nil {
		log.Fatalf("Failed to list collections: %v", err)
	}
	for _, col := range collections {
		if col == collectionName {
			return true
		}
	}
	return false
}

func (ws *WatchService) updateCollection(collectionName string) error {
	newData, err := scrape(collectionName)
	if err != nil {
		return err
	}

	opts := options.FindOne().SetSort(map[string]int{"time": -1})

	var lastData ScrapedData
	err = ws.db.Collection(collectionName).FindOne(context.Background(), bson.D{}, opts).Decode(&lastData)
	if err == mongo.ErrNoDocuments || lastData.Data != newData.Data {
		_, err = ws.db.Collection(collectionName).InsertOne(context.Background(), newData)
		if err != nil {
			return err
		}
		ws.dg.ChannelMessageSend("827661729391050802", newData.Data)
		fmt.Println("New data appended to the collection.")
	} else {
		fmt.Println("Data is the same, no append needed.")
	}

	return nil
}

func (ws *WatchService) Run() {
	collections, err := ws.db.ListCollectionNames(context.Background(), bson.D{})
	if err != nil {
		log.Fatalf("Failed to list collections: %v", err)
	}
	for _, col := range collections {
		ws.updateCollection(col)
	}
}

func scrape(url string) (ScrapedData, error) {
	resp, err := http.Get(url)
	if err != nil {
		return ScrapedData{}, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return ScrapedData{}, err
	}

	scrapedData := doc.Find("main").First().Text()
	currentTime := time.Now()

	return ScrapedData{
		Time: currentTime,
		Data: scrapedData,
	}, nil
}
