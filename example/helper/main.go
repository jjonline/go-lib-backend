package main

import (
	"fmt"
	"github.com/jjonline/go-lib-backend/helper"
)

func main() {
	fmt.Println(helper.RootPath())
	fmt.Println(helper.Path("/test"))
	fmt.Println(helper.Path("test"))
	fmt.Println(helper.Path("test/"))
	fmt.Println(helper.Path("./test/"))
}
