package hook

import "io"

type LifecycleAdapter interface {
	ParsePayload(reader io.Reader) (*Env, error)
	FormatResponse(decision string, reason string) ([]byte, error)
	ChannelName() string
}
