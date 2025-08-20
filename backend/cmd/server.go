package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/whatacotton/tarumi/cmd/app"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	app.Run()
}
