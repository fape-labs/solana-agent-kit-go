package kit

import (
	"github.com/fape-labs/solana-agent-kit-go/pkg/solanaclient"
	"github.com/gagliardetto/solana-go"
)

type Agent interface {
	Signer() solana.PrivateKey
	RPC() *solanaclient.Client
}
