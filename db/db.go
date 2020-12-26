package db

import (
	"context"
	"github.com/sysu-yunz/doubanAnalysis/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"time"
)

type MgoC struct {
	*mongo.Client
}

func NewDB(pwd string) *MgoC {
	//uri := "mongodb+srv://chengqian" + ":" + pwd + "@cluster0-01hyt.azure.mongodb.net/fund?retryWrites=true&w=majority"
	uri := "mongodb://127.0.0.1:27017"
	ctx := context.TODO()
	clientOptions := options.Client().ApplyURI(uri)

	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		log.Error("New Client %+v", err)
	}

	err = client.Connect(ctx)
	if err != nil {
		log.Error("Connect %+v", err)
	}

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Error("Ping %+v", err)
	}

	log.Debug("Connected to MongoDB!")

	return &MgoC{client}
}

func (c *MgoC) InsertMoviesBasic(ms []Movie) {
	col := c.Database("douban").Collection("movie")
	var msi []interface{}
	for _, t := range ms{
		msi = append(msi, t)
	}
	res, err := col.InsertMany(context.TODO(), msi)

	if err != nil {
		log.Error("Inserting movie %+v ", err)
	}

	log.Debug("Inserted movie %+v ", res)
}

func (c *MgoC) GetMovies() *mongo.Cursor {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	col := c.Database("douban").Collection("movie")
	cur, err := col.Find(ctx, bson.M{"date":bson.M{
		"$lt": "2013-10-31",
	}})
	if err != nil {
		log.Error("Finding watches %+v", err)
	}

	return cur

}

func (c *MgoC) GetAllMovies() []Movie {

	var ms []Movie

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	col := c.Database("douban").Collection("movie")
	cur, err := col.Find(ctx, bson.M{})
	if err != nil {
		log.Error("Finding watches %+v", err)
	}

	for cur.Next(ctx) {
		var result Movie
		err := cur.Decode(&result)
		if err != nil { log.Error("Decode watch %+v", err) }
		ms = append(ms, result)
	}

	return ms
}

func (c *MgoC) UpdateMovieRT(m Movie) {
	filter := bson.D{{"subject", m.Subject}}
	update := bson.M{"$set": bson.M{"ep": m.Ep, "runtime": m.RunTime}}

	col := c.Database("douban").Collection("movie")

	updateResult, err := col.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		log.Fatal("%s", err)
	}

	log.Info("Matched %v documents and updated %v documents.\n", updateResult.MatchedCount, updateResult.ModifiedCount)
}