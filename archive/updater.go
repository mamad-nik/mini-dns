package archive

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mamad-nik/mini-dns/agent"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type docs []bson.M

func lookUp(url string) (ip string, err error) {
	ip, err = agent.LookUp(url)
	if err != nil {
		log.Println("Update: LookUp: ", err)
		return "", err
	}
	return ip, nil

}

func (c *Client) Update() {
	colls, err := c.DB.ListCollectionNames(context.TODO(), bson.D{})
	if err != nil {
		fmt.Println(err)
		return
	}
	log.Println("Update: started")

	matchStage := bson.D{
		{Key: "$match", Value: bson.D{
			{Key: "lastmodified", Value: bson.D{
				{Key: "$lte", Value: time.Now().Add(-1 * time.Minute)},
			}},
		}},
	}
	unsetStage := bson.D{{Key: "$unset", Value: bson.A{"_id", "lastmodified"}}}
	sortStage := bson.D{{Key: "$sort", Value: bson.D{{Key: "lastmodified", Value: 1}}}}
	for _, coll := range colls {
		cursor, err := c.DB.Collection(coll).Aggregate(context.TODO(), mongo.Pipeline{matchStage, unsetStage, sortStage})
		if err != nil {
			log.Println("Update: fetch: ", err)
			return
		}
		var results docs
		if err = cursor.All(context.TODO(), &results); err != nil {
			fmt.Println("Update: unmarshall: ", err)
			return
		}
		for _, result := range results {
			log.Println(result)
			go func(r primitive.M, co string) {
				for k := range r {
					if v, ok := r[sld].(string); ok {
						url := reconstruct(co, v, k)
						ip, err := lookUp(url)
						if err == nil {
							c.AddFields(parser(url), ip)
							log.Println(url, " -> ", ip)
						}
					}
				}
			}(result, coll)
		}
	}
	log.Println("Update: done")

}
