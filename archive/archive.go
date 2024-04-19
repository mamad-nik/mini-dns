package archive

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"time"

	minidns "github.com/mamad-nik/mini-dns"
	"github.com/mamad-nik/mini-dns/agent"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	database = "mini-dns"
	set      = "$set"
	sld      = "sld"
)

type Record struct {
	ID  primitive.ObjectID `bson:"_id,omitempty"`
	Sld string             `bson:"sld"`
}

type Client struct {
	DB *mongo.Database
}

func (client Client) Exists(url []string) (bool, error) {
	coll := client.DB.Collection(url[0])

	filter := bson.D{{Key: sld, Value: url[1]}}
	result := coll.FindOne(context.TODO(), filter)

	var m bson.M
	if err := result.Decode(&m); err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, err
	}
	fmt.Println(m)
	return true, nil

}

func (client *Client) Update(url []string, IP string) {
	coll := client.DB.Collection(url[0])

	filter := bson.D{{Key: sld, Value: url[1]}}
	update := bson.D{{Key: set, Value: bson.D{{Key: url[2], Value: IP}}}}

	_, err := coll.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func (client *Client) Insert(url []string, IP string) error {
	coll := client.DB.Collection(url[0])
	r := Record{
		Sld: url[1],
	}

	_, err := coll.InsertOne(context.TODO(), r)
	if err != nil {
		return err
	}
	client.Update(url, IP)
	return nil
}

func (client *Client) Find(URI string) (string, error) {
	url := parser(URI)
	coll := client.DB.Collection(url[0])

	matchStage := bson.D{{Key: "$match", Value: bson.D{{Key: "sld", Value: url[1]}}}}
	unsetStage := bson.D{{Key: "$unset", Value: bson.A{"_id"}}}
	limitStage := bson.D{{Key: "$limit", Value: 1}}

	pipeline := mongo.Pipeline{matchStage, unsetStage, limitStage}
	cursor, err := coll.Aggregate(context.TODO(), pipeline)
	if err != nil {
		return "", err
	}

	var results []bson.M
	err = cursor.All(context.TODO(), &results)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", errors.New("Find: No SLD")
		}
		return "", err
	}

	s, ok := results[0][url[2]]
	if !ok {
		return "", errors.New("Find: No Record")
	}
	return s.(string), nil
}

func (client *Client) Upsert(URI string, IP string, ok bool) error {
	url := parser(URI)

	if ok {
		client.Update(url, IP)
	} else {
		client.Insert(url, IP)
	}
	return nil
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
func (client *Client) assistance(ch minidns.Request) {
	res, err := client.Find(ch.Domain)
	if err != nil {
		var exists bool
		if err.Error() == "Find: No SLD" {
			exists = false
		} else if err.Error() == "Find: No Record" {
			exists = true
		}
		newIP, err := agent.LookUp(ch.Domain)
		if err != nil {
			if dnserr, ok := err.(*net.DNSError); ok && dnserr.IsNotFound {
				ch.Err <- errors.New("no such host")
			} else if ok && dnserr.IsTimeout {
				ch.Err <- errors.New("timeout, try again please")
			}
		} else {
			res = newIP
			client.Upsert(ch.Domain, newIP, exists)
		}
	}
	ch.IP <- res
}
func Manage(mongoURI string, ip <-chan minidns.Request) {
	client := NewDB(mongoURI)
	updateTimer := time.NewTicker(time.Minute)
	defer updateTimer.Stop()

	for {
		select {
		case ch := <-ip:
			go client.assistance(ch)
		case <-updateTimer.C:
			client.Restore()
		}
	}
}
