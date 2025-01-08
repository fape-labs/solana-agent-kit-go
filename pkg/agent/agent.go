package agent

import (
	"github.com/fape-labs/solana-agent-kit-go/pkg/kit"
	"github.com/fape-labs/solana-agent-kit-go/pkg/solanaclient"
	"github.com/gagliardetto/solana-go"
)

var _ kit.Agent = (*Agent)(nil)

type Agent struct {
	privateKey solana.PrivateKey
	client     *solanaclient.Client
}

func New(pk solana.PrivateKey, client *solanaclient.Client) *Agent {
	return &Agent{
		privateKey: pk,
		client:     client,
	}
}

func (a *Agent) Signer() solana.PrivateKey {
	return a.privateKey
}

func (a *Agent) RPC() *solanaclient.Client {
	return a.client
}
