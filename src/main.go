package main

import (
	"fmt"

	"github.com/nonanick/impatience/crawler"
)

func main() {

	var port uint16 = 8080

	fmt.Println("Impatience server, will run on ", port)
	fmt.Printf("Port is of type %T\n", port)
	crawler.Crawl("../public")
}
