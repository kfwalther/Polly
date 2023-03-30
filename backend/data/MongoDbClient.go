package data

import (
	"context"
	"log"

	"github.com/markcheno/go-quote"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Define our MongoDB client.
type MongoDbClient struct {
	mongoClient *mongo.Client
	pollyDb     *mongo.Database
}

// Constructor for a new MongoDbClient object.
func NewMongoDbClient() *MongoDbClient {
	var m MongoDbClient
	return &m
}

// Connect to our MongoDB and grab the Polly data collection.
func (mc *MongoDbClient) ConnectMongoDb() {
	// Set connection URL.
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	// Connect to MongoDB.
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatalf("Unable to connect to the MongoDB instance: %v", err)
	}
	mc.mongoClient = client
	// Check the connection.
	err = mc.mongoClient.Ping(context.Background(), nil)
	if err != nil {
		log.Fatalf("Unable to ping MongoDB instance: %v", err)
	}
	// Connect to the "polly-data" database.
	mc.pollyDb = mc.mongoClient.Database("polly-data")

	// EXAMPLES

	// // Print the stock price data
	// cursor, err := collection.Find(context.Background(), bson.M{})
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer cursor.Close(context.Background())

	// for cursor.Next(context.Background()) {
	// 	var result StockPrice
	// 	err := cursor.Decode(&result)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	fmt.Printf("%s: %.2f (%s)\n", result.Ticker, result.Price, result.Date.Format("2006-01-02"))
	// }
	// if err := cursor.Err(); err != nil {
	// 	log.Fatal(err)
	// }
	return
}

func (mc *MongoDbClient) StoreQuotes(q quote.Quote) {
	// Grab the stock history table.
	historyData := mc.pollyDb.Collection("stock-history")
	// Insert the stock price data into the collection.
	for i, _ := range q.Close {
		_, err := historyData.InsertOne(context.Background(),
			bson.M{"ticker": q.Symbol, "DateTime": q.Date[i], "price": q.Close[i]})
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (mc *MongoDbClient) DisconnectMongoDb() {
	// Disconnect from MongoDB.
	err := mc.mongoClient.Disconnect(context.Background())
	if err != nil {
		log.Fatalf("Error occurred while disconnecting from MongoDB instance: %v", err)
	}
}
