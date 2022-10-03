package snowid

import (
	"github.com/bwmarrin/snowflake"
)

var node *snowflake.Node

const nodeID int64 = 1

func Init() (err error) {
	node, err = snowflake.NewNode(nodeID)
	return
}

// GenID 生成ID时会上锁，确保不重复
func GenID() string {
	return node.Generate().String()
}
