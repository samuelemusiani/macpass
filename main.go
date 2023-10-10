package main

import (
	"fmt"
	"net"
)

func main() {
	var input string

	fmt.Print("Enter a MAC address: ")
	fmt.Scan(&input)

	if _, err := net.ParseMAC(input); err != nil {
		fmt.Println("Mac is invalid")
	} else {
		fmt.Println("Mac is VALID")
	}
}
