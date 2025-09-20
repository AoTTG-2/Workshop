package common

type Appender[V any] interface {
	Append(V)
}
