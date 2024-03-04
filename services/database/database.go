package database

import (
	"context"
	"soa-bot/utils"

	"github.com/cloudflare/cfssl/log"
	"github.com/sarulabs/di"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDB struct {
	Db *mongo.Client
}

func InitDatabase(container di.Container)  (*MongoDB, error){
	cfg := container.Get(utils.DiConfig).(*utils.Config)
	uri := cfg.DB_URI

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)

	dbClient, err := mongo.Connect(context.Background(), opts)
	if err != nil {
		log.Error("Error connecting to MongoDB: %v", err)
		return nil, err
	}
	
	var result bson.M
	if err := dbClient.Database("admin").RunCommand(context.Background(), bson.D{{Key: "ping", Value: 1}}).Decode(&result); err != nil {
		log.Errorf("Error pinging MongoDB: %v", err)
		return nil, err
	}
	log.Info("Successfully connected to MongoDB!")

	return &MongoDB{dbClient}, nil
}

func (m *MongoDB) Close() error {
	return m.Db.Disconnect(context.Background())
}

func (m *MongoDB) CollectionExists(dbName string, collectionName string) (bool, error) {
	collection := m.Db.Database(dbName).Collection(collectionName)
	cursor, err := collection.Find(context.Background(), bson.D{})
	if err != nil {
		return false, err
	}
	defer cursor.Close(context.Background())
	return cursor.Next(context.Background()), nil
}

func (m *MongoDB) CreateCollection(dbName string, collectionName string) error {
	return m.Db.Database(dbName).CreateCollection(context.Background(), collectionName)
}

func (m *MongoDB) GetLatestItem(dbName string, collectionName string) (bson.M, error) {
	collection := m.Db.Database(dbName).Collection(collectionName)
	opts := options.FindOne().SetSort(bson.D{{Key: "time", Value: -1}})
	var result bson.M
	err := collection.FindOne(context.Background(), bson.D{}, opts).Decode(&result)

	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	
	return result, err
}

func (m *MongoDB) InsertItem(dbName string, collectionName string, item interface{}) error {
	collection := m.Db.Database(dbName).Collection(collectionName)
	_, err := collection.InsertOne(context.Background(), item)
	return err
}

func (m *MongoDB) GetCollectionsList(dbName string) ([]string, error) {
	collections, err := m.Db.Database(dbName).ListCollectionNames(context.Background(), bson.D{})
	return collections, err
}


func (m *MongoDB) GetItemByFilter(dbName string, collectionName string, key string, value string) (bson.M, error) {
	collection := m.Db.Database(dbName).Collection(collectionName)
	filter := bson.D{{Key: key, Value: value}}
	var result bson.M
	err := collection.FindOne(context.Background(), filter).Decode(&result)

	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	
	return result, err
}
