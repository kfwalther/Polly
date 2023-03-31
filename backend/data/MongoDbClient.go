package data

import (
	"context"
	"log"
	"time"

	"github.com/markcheno/go-quote"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Define our MongoDB client.
type MongoDbClient struct {
	mongoClient  *mongo.Client
	pollyDb      *mongo.Database
	stockHistory *mongo.Collection
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
	return
}

func (mc *MongoDbClient) TickerExists(ticker string) bool {
	names, err := mc.pollyDb.ListCollectionNames(context.Background(), bson.D{{"options.capped", true}})
	if err != nil {
		log.Printf("ERROR: Failed to get MongoDB collection names: %v", err)
		return false
	}

	for _, name := range names {
		if name == ticker {
			log.Printf("The collection %s exists!", ticker)
			return true
		}
	}
	return false
	// filter := bson.M{"ticker": ticker}
	// cursor, err := mc.pollyDb.Collection(ticker).Find(context.Background(), filter)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer cursor.Close(context.Background())

	// // Check if the cursor has any results
	// if cursor.Next(context.Background()) {
	// 	fmt.Printf("Found data for ticker symbol %s\n", ticker)
	// 	return true
	// } else {
	// 	fmt.Printf("Did not find data for ticker symbol %s\n", ticker)
	// 	return false
	// }
}

func (mc *MongoDbClient) GetLatestQuote(ticker string) time.Time {
	// Setup the filter and sorting options.
	filter := bson.M{"ticker": ticker}
	options := options.FindOne().SetSort(bson.D{{"DateTime", -1}})

	// Lookup the most recent document in the DB for this ticker.
	var result bson.M
	err := mc.pollyDb.Collection(ticker).FindOne(context.Background(), filter, options).Decode(&result)
	if err != nil {
		log.Fatal(err)
	}
	// Get the datetime of the most recent document.
	dateTime, ok := result["DateTime"].(primitive.DateTime)
	if !ok {
		log.Fatal("DateTime field from MongoDB data is not the expected type (primitive.DateTime).")
	}
	return dateTime.Time()
}

func (mc *MongoDbClient) StoreQuote(q quote.Quote) {
	// Grab the corresponding stock history collection from the DB.
	historyData := mc.pollyDb.Collection(q.Symbol)
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
