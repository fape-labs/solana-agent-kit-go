package solanaclient

type ChainID int

const ChainSolana ChainID = 101

type TokenMetadata struct {
	ChainId    ChainID         `json:"chainId"`
	Address    string          `json:"address"`
	Symbol     string          `json:"symbol"`
	Name       string          `json:"name"`
	Collection string          `json:"collection"`
	Ticker     string          `json:"ticker"`
	Decimals   int             `json:"decimals"`
	Image      string          `json:"image"`
	Tags       []string        `json:"tags"`
	Extensions TokenExtensions `json:"extensions"`
}

type TokenExtensions struct {
	CoingeckoID string `json:"coingeckoId"`
	Discord     string `json:"discord"`
	Medium      string `json:"medium"`
	Telegram    string `json:"telegram"`
	Twitter     string `json:"twitter"`
	Website     string `json:"website"`
}
