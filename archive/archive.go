package archive

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	database = "mini-dns"
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

type Client struct {
	DB *mongo.Database
}

func NewSubDomain(domain, ip string) Subdomain {
	return Subdomain{
		Domain: domain,
		IP:     ip,
	}
}

func (client Client) Update(url []string, IP string) {

}

func (client *Client) Insert(URI string, IP string) {
	url := parser(URI)

	coll := client.DB.Collection(url[0])
	r := Record{
		Sld: url[1],
		Subdomains: []Subdomain{
			NewSubDomain(url[2], IP),
		},
	}

	result, err := coll.InsertOne(context.TODO(), r)
	if err != nil {
		log.Println("failed to insert record:", err)
	}
	fmt.Println(result)
}

func (client Client) Find(URI string) {
	url := parser(URI)
	coll := client.DB.Collection(url[0])

	projection := bson.D{{Key: "subdomains.$", Value: 1}, {Key: "_id", Value: 0}, {Key: "subdomain.ip", Value: 1}}
	opts := options.Find().SetProjection(projection)

	var value string
	if len(url) < 3 {
		value = ""
	} else {
		value = url[2]
	}

	fmt.Println(url)
	query := bson.D{{Key: "subdomains.domain", Value: value}}
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
	if len(results) == 0 {
		fmt.Println("no matches")
		return
	}
	fmt.Println(results[0].Subdomains[0].IP)

}

func NewDB(mongoURI string) Client {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Panic(err)
	}
	return Client{
		DB: client.Database(database),
	}

}
