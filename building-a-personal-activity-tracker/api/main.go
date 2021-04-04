package main

import (
	"log"

	"github.com/drgarcia1986/floki/pkg/influxdb"
	"github.com/drgarcia1986/floki/pkg/server"
)

func main() {
	ic := influxdb.NewClient("http://rasp-1:8086", "floki")
	server := server.New(":8080", ic)
	log.Fatal(server.Run())
}
