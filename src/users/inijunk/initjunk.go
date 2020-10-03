package main

import (
	"context"
	"log"

	"uberMessenger/src/common"
	"uberMessenger/src/users"
)

func main() {
	ctx := context.TODO()
	client, err := common.NewClient()
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	dao,err := users.NewDAO(ctx, client)
	if err != nil {
		log.Fatal(err)
	}

	if err:=dao.InitJunk(ctx); err!=nil {
		log.Fatal(err)
	}
}