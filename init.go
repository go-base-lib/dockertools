package dockertools

import "github.com/docker/docker/client"

type ClientGet func() (*client.Client, error)

var (
	DefaultLocalClientGet ClientGet = func() (*client.Client, error) {
		return client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	}

	defaultApiTools *ApiTools
)

func InitLocal() {
	defaultApiTools = NewApiTools(DefaultLocalClientGet)
}

func Init(clientGet ClientGet) {
	defaultApiTools = NewApiTools(clientGet)
}
