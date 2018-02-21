package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
	"github.com/onrik/ethrpc"
)

type clients struct {
	Redis *redis.Client
	Eth   *ethrpc.EthRPC
}

type block struct {
	BlockNumber int
	BlockHash   string
}

func main() {
	c := clients{
		Redis: redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		}),
		Eth: ethrpc.New("http://127.0.0.1:8545"),
	}

	go c.grabber()

	router := mux.NewRouter()
	router.HandleFunc("/api/blocks", c.getBlocks).Methods("GET")

	log.Fatal(http.ListenAndServe(":8000", router))
}

func (c *clients) grabber() {
	client := c.Eth

	for i := 0; ; i++ {
		block, err := client.EthGetBlockByNumber(i, true)

		if err != nil {
			panic(err)
		}

		if block.Hash == "" {
			break
		}

		err = c.Redis.Set(strconv.Itoa(block.Number), block.Hash, 0).Err()
		if err != nil {
			panic(err)
		}
	}
}

func (c *clients) getBlocks(w http.ResponseWriter, r *http.Request) {
	blocks := []block{}

	for i := 0; ; i++ {
		hash, err := c.Redis.Get(strconv.Itoa(i)).Result()
		if err == redis.Nil {
			break
		} else if err != nil {
			panic(err)
		} else {
			blocks = append(blocks, block{i, hash})
		}
	}

	json.NewEncoder(w).Encode(blocks)
}
