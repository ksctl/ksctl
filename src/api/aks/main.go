package main

import (
	"fmt"
)


func aksHandler() string {
	return "Azure CLI API"
}

func main() {
	fmt.Println("Azure AKS API Called")

	fmt.Println(aksHandler())
}
