package pump

import "github.com/fape-labs/solana-agent-kit-go/pkg/kit"

type Client struct {
	agent kit.Agent
}

func NewClient(agent kit.Agent) *Client {
	return &Client{
		agent: agent,
	}
}
