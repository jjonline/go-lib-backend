package main

import (
	"fmt"
	"github.com/jjonline/go-lib-backend/qywx"
)

func main() {
	key := "**" // please set your key
	instance := qywx.New(key, true, nil)

	var err error

	// text
	err = instance.Text("hello world", nil, nil)
	fmt.Println(err)

	// markDown
	err = instance.Markdown("# title \n > hello world")
	fmt.Println(err)

	// news
	err = instance.News([]qywx.Article{{
		Title:       "图文title1",
		Description: "图文Description",
		URL:         "https://github.com/",
		PicUrl:      "https://blog.jjonline.cn/Upload/image/202212/20221217033855.jpeg",
	}})
	fmt.Println(err)

}
