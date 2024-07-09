package main

import (
	"context"
	"fmt"
	"github.com/jjonline/go-lib-backend/guzzle"
)

func main() {
	// 实例化redis
	client := guzzle.New(nil, nil)

	// Get
	res, err := client.Get(context.TODO(), "https://www.baidu.com/", nil, nil)
	fmt.Printf("%#v\n", res)
	fmt.Println(err)

	// Get with Query
	res1, err1 := client.Get(context.TODO(), "https://www.baidu.com/?s=1", map[string]string{"key": "v", "wd": "test"}, nil)
	fmt.Printf("%#v\n", res1)
	fmt.Println(err1)

	// post
	res2, err2 := client.PostForm(context.TODO(), "https://www.baidu.com/?s=1", map[string]string{"key": "v", "wd": "test"}, nil)
	fmt.Printf("%#v\n", res2)
	fmt.Println(err2)

	// Delete
	res3, err3 := client.Delete(context.TODO(), "https://www.baidu.com/?s=1", map[string]string{"key": "v", "wd": "test"}, nil)
	fmt.Printf("%#v\n", res3)
	fmt.Println(err3)
}
