package main

import (
	"fmt"
	"github.com/jjonline/go-lib-backend/defense"
	"github.com/jjonline/go-lib-backend/example/defense/client"
	"time"
)

func main()  {
	// 实例化redis
	redisClient := client.NewRedis()

	// 实例化defense
	sDefense := defense.New(redisClient, 10 * time.Minute, 5)

	// try trigger
	for i := 0; i <= 10; i++ {
		if sDefense.Defense("try_defense_key") !=nil {
			fmt.Println("defense status in")
		} else {
			fmt.Println("not ini defense status")
		}
	}
}
