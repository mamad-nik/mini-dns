package archive

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func (c *Client) SearchByIP(ip string) (domain string, err error) {
	colls, err := c.DB.ListCollectionNames(context.TODO(), bson.D{})
	log.Println(colls)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	log.Println(ip)
	pipeline := mongo.Pipeline{
		bson.D{{
			Key: "$project", Value: bson.M{
				"sld": 1,
				"_id": 0,
				"convertedDoc": bson.A{
					bson.M{
						"$filter": bson.M{
							"input": bson.M{"$objectToArray": "$$ROOT"},
							"as":    "docElem",
							"cond":  bson.M{"$eq": bson.A{"$$docElem.v", ip}},
						},
					},
				},
			}},
		},
		bson.D{{
			Key: "$match", Value: bson.M{
				"convertedDoc": bson.M{
					"$ne": bson.A{},
				},
			},
		}},
	}

	for _, coll := range colls {
		cursor, err := c.DB.Collection(coll).Aggregate(context.TODO(), pipeline)
		if err != nil {
			log.Println(err)
			return "", err
		}
		var results []bson.M
		err = cursor.All(context.TODO(), &results)
		if err != nil {
			log.Println(err)
			return "", err
		}
		for _, res := range results {
			log.Println(res)
			sub, ok := res["convertedDoc"].(primitive.A)[0].(primitive.A)[0].(primitive.M)["k"].(string)
			if !ok {
				return "", fmt.Errorf("DB Problem")
			}
			sld, ok := res["sld"].(string)
			if !ok {
				return "", fmt.Errorf("DB Problem")
			}
			domain = reconstruct(coll, sld, sub)
		}
	}
	log.Println(err, domain)
	err = nil
	return
}

/*db.com.aggregate([
  {
    $project: {
      convertedDoc: {
        $map: {
          input: { $objectToArray: "$$ROOT" },
          as: "docElem",
          in: {
            k: "$$docElem.k",
            v: {
              $cond: {
                if: { $eq: ["$$docElem.k", "10.10.34.36"] },
                then: "$$docElem.v",
                else: null
              }
            }
          }
        }
      }
    }
  },
  {
    $match: {
      convertedDoc: {
        $elemMatch: { v: { $ne: null } } // Check for at least one element with non-null value
      }
    }
  }
])
*/
