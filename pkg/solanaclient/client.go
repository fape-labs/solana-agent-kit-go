package solanaclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gagliardetto/solana-go"
	sendandconfirmtransaction "github.com/gagliardetto/solana-go/rpc/sendAndConfirmTransaction"
	"strconv"
	"time"

	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
)

type Client struct {
	RPC    *rpc.Client
	WS     *ws.Client
	keyMap map[string]solana.PrivateKey
}

func NewClient(ctx context.Context, rpcEndpoint, wsEndpoint string) (*Client, error) {
	wsClient, err := ws.Connect(ctx, wsEndpoint)
	if err != nil {
		return nil, err
	}

	return &Client{
		RPC:    rpc.New(rpcEndpoint),
		WS:     wsClient,
		keyMap: make(map[string]solana.PrivateKey),
	}, nil
}

var ErrTokenAccountNotFound = errors.New("associated token account not found")

func (c *Client) AddSigner(key solana.PrivateKey) {
	c.keyMap[key.PublicKey().String()] = key
}

func (c *Client) DeriveAndCheckTokenAccount(owner solana.PublicKey, mint solana.PublicKey) (pubKey *solana.PublicKey, nonce uint8, err error) {

	addr, nonce, err := DeriveAssociatedTokenAddress(owner, mint)
	if err != nil {
		return nil, 0, err
	}

	getAccountsResult, err := c.RPC.GetTokenAccountsByOwner(context.TODO(), owner, &rpc.GetTokenAccountsConfig{
		Mint: &mint,
	}, &rpc.GetTokenAccountsOpts{
		Commitment: rpc.CommitmentFinalized,
		Encoding:   solana.EncodingJSONParsed,
	})

	if err != nil {
		return nil, 0, err
	}

	if len(getAccountsResult.Value) > 0 {
		return addr, nonce, nil
	}

	return addr, nonce, ErrTokenAccountNotFound
}

func (c *Client) GetAssociatedTokenAccount(owner, mint string) (*AssociatedTokenAccountResult, error) {
	ownerWallet := solana.MustPublicKeyFromBase58(owner)
	ticketMint := solana.MustPublicKeyFromBase58(mint)
	getAccountsResult, err := c.RPC.GetTokenAccountsByOwner(context.TODO(), ownerWallet, &rpc.GetTokenAccountsConfig{
		Mint: &ticketMint,
	}, &rpc.GetTokenAccountsOpts{
		Commitment: rpc.CommitmentFinalized,
		Encoding:   solana.EncodingJSONParsed,
	})

	if err != nil {
		return nil, err
	}

	if len(getAccountsResult.Value) > 0 {
		tokenAccount, err := rawDataToTokenAccount(getAccountsResult.Value[0].Pubkey.String(), getAccountsResult.Value[0].Account.Data.GetRawJSON())
		if err != nil {
			return nil, err
		}
		return tokenAccount, nil
	}

	return nil, ErrNoATAFound
}

func (c *Client) GetTokenBalanceOf(owner solana.PublicKey, mint solana.PublicKey) uint64 {
	address, _, err := DeriveAssociatedTokenAddress(owner, mint)
	if err != nil {
		return 0
	}

	getTokenAccountBalanceResult, err := c.RPC.GetTokenAccountBalance(context.Background(), *address, rpc.CommitmentFinalized)
	if err != nil {
		return 0
	}

	parseUint, err := strconv.ParseUint(getTokenAccountBalanceResult.Value.Amount, 10, 64)
	if err != nil {
		return 0
	}
	return parseUint
}

func rawDataToTokenAccount(pubKey string, message json.RawMessage) (*AssociatedTokenAccountResult, error) {
	tokenAccount := &AssociatedTokenAccountResult{}
	err := json.Unmarshal(message, tokenAccount)
	if err != nil {
		return nil, err
	}
	tokenAccount.PubKey = pubKey
	return tokenAccount, nil
}

func (c *Client) SignAndSendTx(b64 string) (*solana.Signature, error) {

	ctx := context.Background()
	tx, err := solana.TransactionFromBase64(b64)

	if err != nil {
		return nil, err
	}

	signatures, err := tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		p, ok := c.keyMap[key.String()]
		if ok {
			return &p
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	opts := rpc.TransactionOpts{
		SkipPreflight:       true,
		PreflightCommitment: rpc.CommitmentFinalized,
	}

	to := time.Second * 30

	signtr, err := sendandconfirmtransaction.SendAndConfirmTransactionWithOpts(
		ctx,
		c.RPC,
		c.WS,
		tx,
		opts,
		&to,
	)

	if err != nil {
		return &signatures[0], fmt.Errorf("failed tx ( %s ) to %v", signtr.String(), err)
	}

	return &signtr, nil
}
