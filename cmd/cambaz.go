package main

import (
	"github.com/brsyuksel/cambaz"
	"log"
)

func main() {
	if err := cambaz.Main(); err != nil {
		log.Fatalln(err)
	}
}
