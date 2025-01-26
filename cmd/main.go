package main

import (
	"fmt"

	"github.com/Arash-Afshar/linkchecker"
	"github.com/alecthomas/kong"
)

func prettyPrint(links []linkchecker.Link) {
	fmt.Println("V L URL")
	for _, link := range links {
		fmt.Println(link.String())
	}
}

func main() {
	var config linkchecker.Config
	ctx := kong.Parse(&config)
	if ctx.Error != nil {
		fmt.Printf("Error parsing arguments: %v\n", ctx.Error)
		return
	}

	if err := linkchecker.ValidateConfig(&config); err != nil {
		fmt.Printf("Error validating configuration: %v\n", err)
		return
	}

	links, err := linkchecker.CheckLinks(&config)
	if err != nil {
		fmt.Printf("Error checking links: %v\n", err)
		return
	}
	prettyPrint(links)
}
