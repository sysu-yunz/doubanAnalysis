package db

import (
	"context"
	"github.com/sysu-yunz/doubanAnalysis/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
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
