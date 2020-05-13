package main

import (
	"github.com/micro/go-micro/v2/web"
)

func main() {
	connectToMongo()
	insertData()
	service := web.NewService(
		web.Name("foo"),
		web.Address(":8080"),
	)
	service.HandleFunc("/data", findDataHandler)
	service.Init()
	service.Run()
}
