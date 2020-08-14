package crawler

import (
	"fmt"
)

// Crawl will search for all the files inside the root
func Crawl(root string) {
	fmt.Println("Will craw into dir ", root)
}

//DirectoryGraph -
type DirectoryGraph struct {
}

//FileCrawlInfo - gathered
type FileCrawlInfo struct {
}

//RootCrawlGraph - Define the root graph crawl result
type RootCrawlGraph struct {
	directories []DirectoryGraph
	files       []FileCrawlInfo
}
