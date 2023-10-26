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
	databaseName string
	mongoClient  *mongo.Client
	ctx          context.Context
	pollyDb      *mongo.Database
	stockHistory *mongo.Collection
}

// Define a record structure to temporarily house the DB data.
type TempQuote struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	Ticker   string             `bson:"ticker,omitempty"`
	DateTime time.Time          `bson:"DateTime,omitempty"`
	Price    float64            `bson:"price,omitempty"`
}

// Constructor for a new MongoDbClient object.
func NewMongoDbClient() *MongoDbClient {
	var m MongoDbClient
	return &m
}

// Connect to our MongoDB and grab the Polly data collection.
func (mc *MongoDbClient) ConnectMongoDb(connectionUri string, dbName string) {
	mc.databaseName = dbName
	// Set connection URL.
	clientOptions := options.Client().ApplyURI(connectionUri)
	// Save our database context.
	mc.ctx = context.Background()

	// Connect to MongoDB.
	client, err := mongo.Connect(mc.ctx, clientOptions)
	if err != nil {
		log.Fatalf("Unable to connect to the MongoDB instance: %v", err)
	}
	mc.mongoClient = client
	// Check the connection.
	err = mc.mongoClient.Ping(mc.ctx, nil)
	if err != nil {
		log.Fatalf("Unable to ping MongoDB instance: %v", err)
	}
	// Connect to the database.
	mc.pollyDb = mc.mongoClient.Database(mc.databaseName)
	return
}

func (mc *MongoDbClient) TickerExists(ticker string) bool {
	// Get the complete list of collections in the database.
	names, err := mc.pollyDb.ListCollectionNames(mc.ctx, bson.M{"type": "collection"})
	if err != nil {
		log.Printf("ERROR: Failed to get MongoDB collection names: %v", err)
		return false
	}
	// Look for the collection matching the given ticker.
	for _, name := range names {
		if name == ticker {
			return true
		}
	}
	return false
}

func (mc *MongoDbClient) GetLatestQuote(ticker string) time.Time {
	// Setup the filter and sorting options.
	filter := bson.M{"ticker": ticker}
	options := options.FindOne().SetSort(bson.D{{"DateTime", -1}})

	// Lookup the most recent document in the DB for this ticker.
	var result bson.M
	err := mc.pollyDb.Collection(ticker).FindOne(mc.ctx, filter, options).Decode(&result)
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

func (mc *MongoDbClient) GetTickerData(ticker string) quote.Quote {
	var data quote.Quote
	data.Symbol = ticker
	// Grab the corresponding stock history collection from the DB.
	cursor, err := mc.pollyDb.Collection(ticker).Find(mc.ctx, bson.M{})
	if err != nil {
		log.Fatalf("GetTickerData: Failed to find collection for ticker %s: %v", ticker, err)
	}
	defer cursor.Close(mc.ctx)

	// Iterate through the records, saving each.
	for cursor.Next(mc.ctx) {
		var q TempQuote
		err := cursor.Decode(&q)
		if err != nil {
			log.Fatalf("GetTickerData: Failed to decode the record: %v: %v", q, err)
		}
		data.Date = append(data.Date, q.DateTime)
		data.Close = append(data.Close, q.Price)
	}
	return data
}

func (mc *MongoDbClient) StoreTickerData(q quote.Quote) {
	// Grab the corresponding stock history collection from the DB.
	historyData := mc.pollyDb.Collection(q.Symbol)
	// Insert the stock price data into the collection.
	for i, _ := range q.Close {
		_, err := historyData.InsertOne(mc.ctx,
			bson.M{"ticker": q.Symbol, "DateTime": q.Date[i], "price": q.Close[i]})
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (mc *MongoDbClient) DisconnectMongoDb() {
	// Disconnect from MongoDB.
	err := mc.mongoClient.Disconnect(mc.ctx)
	if err != nil {
		log.Fatalf("Error occurred while disconnecting from MongoDB instance: %v", err)
	}
}
