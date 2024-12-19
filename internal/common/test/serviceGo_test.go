package test

import (
	"fmt"
	"service"
	"testing"
)

func TestFindAvailableEndpoint(t *testing.T) {
	ips, ports, err := service.FindAvailableEndpoint(1, 2)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(ips, ports)
}
