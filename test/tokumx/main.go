package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Person struct {
	Name  string
	Phone string
}

type Oplog struct {
	ID bson.ObjectId `bson:"_id"`
}

type logger struct{}

func (logger) Output(calldepth int, s string) error {
	log.Printf("%d: %s", calldepth, s)
	return nil
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("arguments: [tail|insert] [server]")
	}

	session, err := mgo.Dial(os.Args[2])
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)
	session.SetSafe(&mgo.Safe{WMode: "majority", J: true})

	switch os.Args[1] {
	case "tail":
		db := session.DB("local")
		c := db.C("oplog.rs")
		var lastId bson.ObjectId
		var result Oplog
		iter := c.Find(nil).Sort("$natural").Tail(1 * time.Minute)
		for {
			fmt.Println("++ tailing for 1 minute")
			for iter.Next(&result) {
				fmt.Println(result.ID)
				lastId = result.ID
			}
			if iter.Err() != nil {
				log.Printf("ERROR: %v\n", iter.Err())
				break
			}
			if iter.Timeout() {
				iter.Close()
				fmt.Println("++ timeout")
			} else if len(lastId) == 0 {
				fmt.Println("++ no id yet")
				time.Sleep(10 * time.Second)
				continue
			}
			query := c.Find(bson.M{"_id": bson.M{"$gt": lastId}})
			iter = query.Sort("$natural").Tail(5 * time.Second)
		}
		iter.Close()

	case "insert":
		mgo.SetLogger(logger{})
		mgo.SetDebug(true)
		db := session.DB("test")
		c := db.C("people")

		err = c.Insert(
			&Person{"Ale", "+55 53 8116 9639"},
			&Person{"Cla", "+55 53 8402 8510"},
		)
		if err != nil {
			log.Fatal("insert", err)
		}

		result := Person{}
		err = c.Find(bson.M{"name": "Ale"}).One(&result)
		if err != nil {
			log.Fatal("find", err)
		}

		fmt.Println("Phone:", result.Phone)
	default:
		log.Fatal("arguments: [tail|insert] [server]")
	}
}
