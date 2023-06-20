package main

import (
	"fmt"
	"github.com/jjonline/go-lib-backend/helper"
)

func main() {
	fmt.Println(part.RootPath())
	fmt.Println(part.Path("/test"))
	fmt.Println(part.Path("test"))
	fmt.Println(part.Path("test/"))
	fmt.Println(part.Path("./test/"))
}
