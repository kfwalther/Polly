package data

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Define our MongoDB client.
type MongoDbClient struct {
	mongoClient *mongo.Client
}

// Constructor for a new MongoDbClient object.
func NewMongoDbClient() *MongoDbClient {
	var m MongoDbClient
	return &m
}

// Connect to our MongoDB and grab the Polly data collection.
func (mc *MongoDbClient) ConnectMongoDb() *mongo.Collection {
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
	// Connect to the "stock-history" collection in the "polly-data" database.
	stockHistoryCollection := mc.mongoClient.Database("polly-data").Collection("stock-history")
	return stockHistoryCollection

	// EXAMPLES
	// // Insert the stock price data into the collection
	// for _, price := range prices {
	// 	_, err := collection.InsertOne(context.Background(), bson.M{"ticker": price.Ticker, "date": price.Date, "price": price.Price})
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// }

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

}

func (mc *MongoDbClient) DisconnectMongoDb() {
	// Disconnect from MongoDB.
	err := mc.mongoClient.Disconnect(context.Background())
	if err != nil {
		log.Fatalf("Error occurred while disconnecting from MongoDB instance: %v", err)
	}
}
