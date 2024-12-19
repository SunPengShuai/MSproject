package service

import (
	"fmt"
	"testing"
)

func TestFindAvailableEndpoint(t *testing.T) {
	ips, ports, err := FindAvailableEndpoint(1, 2)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(ips, ports)
}
