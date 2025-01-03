package MqApi

type MqApi interface {
	ListenAndGet()
	PushMsg()
}
