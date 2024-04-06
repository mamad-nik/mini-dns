package archive

import (
	"context"
	"fmt"
	"log"

	"github.com/mamad-nik/mini-dns/agent"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type docs []bson.M

func (c *Client) Restore() {
	colls, err := c.DB.ListCollectionNames(context.TODO(), bson.D{})
	if err != nil {
		fmt.Println(err)
		return
	}
	log.Println("Update: started")

	opts := options.Find().SetProjection(bson.D{{Key: "_id", Value: false}})

	for _, v := range colls {
		curser, err := c.DB.Collection(v).Find(context.TODO(), bson.D{}, opts)
		if err != nil {
			log.Println("Update: fetch: ", err)
			return
		}
		var results docs
		if err = curser.All(context.TODO(), &results); err != nil {
			fmt.Println("Update: unmarshall: ", err)
			return
		}
		for _, v1 := range results {
			sld := v1["sld"].(string)
			for i := range v1 {
				if i != "sld" {
					uri := ""
					if i == "-val" {
						uri = sld + "." + v
					} else {
						uri = i + "." + sld + "." + v
					}

					ip, err := agent.LookUp(uri)
					if err != nil {
						log.Println("Update: LookUp: ", err)
						continue
					}
					log.Printf("Update: %s -> %s\n", uri, ip)
					c.Update(parser(uri), ip)
				}
			}

		}
	}
	log.Println("Update: done")
}
