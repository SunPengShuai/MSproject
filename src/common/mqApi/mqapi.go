package mqApi

type MsgType int

const (
	SimpleMsg MsgType = iota
	StorageUpdate
)

type MqMsg struct {
	MsgType MsgType     `json:"type"`
	Data    interface{} `json:"data"`
}
type MqApi interface {
	RecvMsg(qname string) (interface{}, error)
	SendMsg(msg MqMsg, routingKey string) error
}
