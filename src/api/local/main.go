package main

import (
	"fmt"
)


func dockerHandler() string {
	return "Docker API"
}

func main() {
	fmt.Println("Local API Called")

	fmt.Println(dockerHandler())
}
