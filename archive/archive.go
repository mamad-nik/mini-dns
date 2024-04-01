package archive

import (
	"context"
	"fmt"
	"log"
	"slices"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Subdomain struct {
	Domain string `bson:"domain"`
	IP     string `bson:"ip"`
}

type Record struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	Sld        string             `bson:"sld"`
	Subdomains []Subdomain        `bson:"subdomains"`
}

type DB struct {
	Client *mongo.Client
}

func (db DB) insert(url []string, IP string) {
	//	coll := db.Client.Database("db").Collection(url[0])

}

func parser(url string) (s []string) {
	s = strings.Split(url, ".")
	sub := len(s) - 2

	str := strings.Join(s[:sub], ".")
	s = slices.Delete(s, 0, sub)

	slices.Reverse(s)
	s = append(s, str)
	//s = slices.Delete(s, 0, 1)
	return
}

func (db *DB) Find(URI string) {
	url := parser(URI)
	coll := db.Client.Database("mini-dns").Collection(url[0])

	projection := bson.D{{Key: "subdomains.ip.$", Value: 1}, {Key: "_id", Value: 0}, {Key: "subdomain.ip", Value: 1}}
	opts := options.Find().SetProjection(projection)

	query := bson.D{{Key: "subdomains.domain", Value: "dns"}}
	curser, err := coll.Find(context.TODO(), query, opts)
	if err != nil {
		fmt.Println(err)
		return
	}

	var results []Record
	if err = curser.All(context.TODO(), &results); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(results[0].Subdomains[0].IP)

}

func NewDB(mongoURI string) DB {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Panic(err)
	}
	return DB{
		Client: client,
	}

}
