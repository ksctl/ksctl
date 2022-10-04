package main

import (
	"fmt"
)


func eksHandler() string {
	return "AWS CLI API"
}

func main() {
	fmt.Println("AWS EKS API Called")

	fmt.Println(eksHandler())
}
