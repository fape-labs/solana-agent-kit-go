package solanaclient

import (
	"errors"
	"github.com/gagliardetto/solana-go"
	"math"
	"strconv"

	"github.com/shopspring/decimal"
)

var ErrNoATAFound = errors.New("no ATA found")

type AssociatedTokenAccountResult struct {
	PubKey string `json:"-"`
	Parsed struct {
		Info struct {
			IsNative    bool   `json:"isNative"`
			Mint        string `json:"mint"`
			Owner       string `json:"owner"`
			State       string `json:"state"`
			TokenAmount struct {
				Amount         string  `json:"amount"`
				Decimals       int     `json:"decimals"`
				UiAmount       float64 `json:"uiAmount"`
				UiAmountString string  `json:"uiAmountString"`
			} `json:"tokenAmount"`
		} `json:"info"`
		Type string `json:"type"`
	} `json:"parsed"`
	Program string `json:"program"`
	Space   int    `json:"space"`
}

func (ata *AssociatedTokenAccountResult) InexactFloatAmount() float64 {
	return ata.DecimalAmount().InexactFloat64()
}

func (ata *AssociatedTokenAccountResult) DecimalAmount() decimal.Decimal {
	return decimal.RequireFromString(ata.Parsed.Info.TokenAmount.Amount).Div(decimal.NewFromFloat(math.Pow(10, float64(ata.Parsed.Info.TokenAmount.Decimals))))
}

func (ata *AssociatedTokenAccountResult) Uint64Amount() uint64 {
	parseUint, err := strconv.ParseUint(ata.Parsed.Info.TokenAmount.Amount, 10, 64)
	if err != nil {
		return 0
	}
	return parseUint
}

func DeriveAssociatedTokenAddress(owner solana.PublicKey, mint solana.PublicKey) (pubKey *solana.PublicKey, nonce uint8, err error) {
	addr, nonce, err := solana.FindProgramAddress([][]byte{owner.Bytes(), solana.TokenProgramID.Bytes(), mint.Bytes()}, solana.SPLAssociatedTokenAccountProgramID)
	if err != nil {
		return nil, 0, err
	}
	return &addr, nonce, nil
}
