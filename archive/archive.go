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
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	Sld          string             `bson:"sld"`
	LastModified time.Time          `bson:"lastmodified"`
}

type Client struct {
	DB *mongo.Database
}

func (client *Client) Subdomains(url []string) (map[string]string, error) {
	coll := client.DB.Collection(url[0])

	matchStage := bson.D{{Key: "$match", Value: bson.D{{Key: "sld", Value: url[1]}}}}
	unsetStage := bson.D{{Key: "$unset", Value: bson.A{"_id", "lastmodified"}}}

	cursor, err := coll.Aggregate(context.TODO(), mongo.Pipeline{matchStage, unsetStage})
	if err != nil {
		return nil, err
	}
	var results []bson.M
	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}
	res := make(map[string]string)

	for k, v := range results[0] {
		if k != sld {

			res[reconstruct(url[0], url[1], k)] = v.(string)
		}
	}

	return res, nil
}

func (client *Client) AddFields(url []string, IP string) {
	coll := client.DB.Collection(url[0])

	filter := bson.D{{Key: sld, Value: url[1]}}
	update := bson.D{{Key: set, Value: bson.D{{Key: url[2], Value: IP}, {Key: "lastmodified", Value: time.Now()}}}}

	_, err := coll.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func (client *Client) Insert(url []string, IP string) error {
	coll := client.DB.Collection(url[0])
	r := Record{
		Sld:          url[1],
		LastModified: time.Now(),
	}

	_, err := coll.InsertOne(context.TODO(), r)
	if err != nil {
		return err
	}
	client.AddFields(url, IP)
	return nil
}

func (client *Client) Find(URI string) (string, error) {
	url, err := parser(URI)
	if err != nil {
		return "", err
	}

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
	if len(results) == 0 {
		return "", errors.New("Find: No SLD")

	}
	r := results[0]["lastmodified"].(primitive.DateTime)
	if r.Time().Compare(time.Now().Add(-1*time.Minute)) < 0 {
		log.Println("Old data")
		return "", errors.New("Find: No Record or Old")
	}
	s, ok := results[0][url[2]]
	if !ok {
		return "", errors.New("Find: No Record or Old")
	}

	return s.(string), nil
}

func (client *Client) Upsert(URI string, IP string, ok bool) error {
	url, err := parser(URI)
	if err != nil {
		return err
	}
	if ok {
		client.AddFields(url, IP)
	} else {
		client.Insert(url, IP)
	}
	return nil
}

func NewClient(mongoURI string) Client {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Panic(err)
	}
	return Client{
		DB: client.Database(database),
	}
}

func (client *Client) Meta() int32 {
	colls, err := client.DB.ListCollectionNames(context.TODO(), bson.D{})
	if err != nil {
		fmt.Println(err)
		return 0
	}
	projectStage := bson.D{{Key: "$project", Value: bson.D{{Key: "fields", Value: bson.D{{Key: "$objectToArray", Value: "$$ROOT"}}}}}}
	unwindStage := bson.D{{Key: "$unwind", Value: "$fields"}}
	groupStage := bson.D{{Key: "$group", Value: bson.D{{Key: "_id", Value: nil}, {Key: "fields", Value: bson.D{{Key: "$addToSet", Value: "$fields.k"}}}}}}
	projectCountStage := bson.D{{Key: "$project", Value: bson.D{{Key: "_id", Value: 0}, {Key: "count", Value: bson.D{{Key: "$size", Value: "$fields"}}}}}}

	// Aggregate pipeline
	pipeline := mongo.Pipeline{projectStage, unwindStage, groupStage, projectCountStage}

	for _, c := range colls {
		cursor, err := client.DB.Collection(c).Aggregate(context.TODO(), pipeline)
		if err != nil {
			log.Println(err)
			return 0
		}
		var results []bson.M
		if err = cursor.All(context.TODO(), &results); err != nil {
			log.Println(err)
			return 0
		}
		if len(results) > 0 {
			return results[0]["count"].(int32)

		}
	}
	return 0
}

func (client *Client) assistance(ch minidns.Request) {

	if ch.ReqType == "ip" {
		res, err := client.Find(ch.Requset)
		if err != nil {
			if err.Error() == "invalid url" {
				ch.Err <- err
			}
			var exists bool
			if err.Error() == "Find: No SLD" {
				exists = false
			} else if err.Error() == "Find: No Record or Old" {
				exists = true
			}
			newIP, err := agent.LookUp(ch.Requset)
			if err != nil {
				if dnserr, ok := err.(*net.DNSError); ok && dnserr.IsNotFound {
					ch.Err <- errors.New("no such host")
				} else if ok && dnserr.IsTimeout {
					ch.Err <- errors.New("timeout, try again please")
				}
			} else {
				res = newIP
				client.Upsert(ch.Requset, newIP, exists)
			}
		}
		fmt.Println(res)
		ch.Response <- res
	} else if ch.ReqType == "domain" {
		res, err := client.SearchByIP(ch.Requset)
		if err != nil {
			ch.Err <- err
			return
		}
		log.Println("ip", res)
		ch.Response <- res
	}

}

func (c *Client) returnAll() (map[string]string, error) {
	colls, err := c.DB.ListCollectionNames(context.TODO(), bson.D{})
	log.Println(colls)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	m := make(map[string]string)
	for _, coll := range colls {
		cursor, err := c.DB.Collection(coll).Aggregate(context.TODO(), mongo.Pipeline{bson.D{{
			Key: "$unset", Value: bson.A{"_id", "lastmodified"},
		}}})
		if err != nil {
			log.Println("asdasd", err)
			return nil, err
		}
		var res []bson.M
		if err = cursor.All(context.TODO(), &res); err != nil {
			log.Println("dfgdfg", err)
			return nil, err
		}

		for _, v := range res {
			s := v[sld].(string)
			for sub, ip := range v {
				if sub != s {
					if i, ok := ip.(string); ok {
						m[reconstruct(coll, s, sub)] = i
					}
				}
			}

		}
	}
	r := c.Meta()
	if r != 0 {
		m["number of records"] = fmt.Sprint(r)
	}
	return m, nil
}

func (c *Client) handleMulti(mr minidns.MultiRequest) {
	if mr.ReqType == "sub" {
		url, err := parser(mr.Requset)
		if err != nil {
			mr.Err <- err
			return
		}
		res, err := c.Subdomains(url)
		if err != nil {
			mr.Err <- err
			return
		}
		mr.Response <- res
	} else if mr.ReqType == "all" {
		fmt.Printf("\"hello\": %v\n", "hello")
		m, err := c.returnAll()
		if err != nil {
			log.Println(err)
			return
		}
		mr.Response <- m
	}
}

func Manage(mongoURI string, singleReq <-chan minidns.Request, MultiReq <-chan minidns.MultiRequest) {
	client := NewClient(mongoURI)
	updateTimer := time.NewTicker(time.Minute * 5)
	defer updateTimer.Stop()

	for {
		select {
		case mr := <-MultiReq:
			go client.handleMulti(mr)
		case ch := <-singleReq:
			go client.assistance(ch)
		case <-updateTimer.C:
			client.Update()
		}
	}
}
