package amqp091otel

import "runtime/debug"

var (
	version        = "v0.0.0"
	amqpLibVersion = "unknown"
)

func init() {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}
	for _, dep := range info.Deps {
		if dep.Path == amqpLibName {
			amqpLibVersion = dep.Version
		}
		if dep.Path == libName {
			version = dep.Version
		}
	}
}
