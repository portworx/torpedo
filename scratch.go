package main

import (
	"fmt"
)

var supportedDataServices = map[string]string{"cas": "Cassandra", "zk": "ZooKeeper", "kf": "Kafka", "rmq": "RabbitMQ", "pg": "PostgreSQL"}

func main() {
	for key := range supportedDataServices {
		fmt.Println("keys ", key)
	}

}
