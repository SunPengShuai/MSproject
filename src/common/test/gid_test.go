package test

import (
	"fmt"
	"gid"
	snowflak "github.com/bwmarrin/snowflake"
	"os"
	"testing"
)

func TestGid(t *testing.T) {
	node, err := snowflak.NewNode(0)
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}

	for i := 0; i < 20; i++ {
		id := node.Generate()

		fmt.Printf("int64 ID: %d\n", id)
		fmt.Printf("string ID: %s\n", id)
		fmt.Printf("base2 ID: %s\n", id.Base2())
		fmt.Printf("base64 ID: %s\n", id.Base64())
		fmt.Printf("ID time: %d\n", id.Time())
		fmt.Printf("ID node: %d\n", id.Node())
		fmt.Printf("ID step: %d\n", id.Step())
		fmt.Println("--------------------------------")
	}
	var gidint gid.GID
	gidint = &gid.SnowFlakeGID{
		ID: node.Generate(),
	}

	gid, err := gidint.GetInt64()
	println(gid)

}
