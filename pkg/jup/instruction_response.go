package jup

import "github.com/gagliardetto/solana-go"

type TxAccount struct {
	PublicKey  solana.PublicKey `json:"pubkey"`
	IsWritable bool             `json:"isWritable"`
	IsSigner   bool             `json:"isSigner"`
}

type RawInstruction struct {
	ProgramId    solana.PublicKey `json:"programId"`
	AccountsMeta []*TxAccount     `json:"accounts"`
	DataB        []byte           `json:"data"`
}

func (si *RawInstruction) ProgramID() solana.PublicKey {
	return si.ProgramId
}

func (si *RawInstruction) Accounts() []*solana.AccountMeta {
	m := make([]*solana.AccountMeta, len(si.AccountsMeta))
	for i := 0; i < len(si.AccountsMeta); i++ {
		m[i] = &solana.AccountMeta{
			PublicKey:  si.AccountsMeta[i].PublicKey,
			IsWritable: si.AccountsMeta[i].IsWritable,
			IsSigner:   si.AccountsMeta[i].IsSigner,
		}
	}
	return m
}

func (si *RawInstruction) Data() ([]byte, error) {
	return si.DataB, nil
}

type InstructionResponse struct {
	TokenLedgerInstruction      interface{}       `json:"tokenLedgerInstruction"`
	ComputeBudgetInstructions   []*RawInstruction `json:"computeBudgetInstructions"`
	SetupInstructions           []*RawInstruction `json:"setupInstructions"`
	SwapInstruction             *RawInstruction   `json:"swapInstruction"`
	CleanupInstruction          *RawInstruction   `json:"cleanupInstruction"`
	OtherInstructions           []interface{}     `json:"otherInstructions"`
	AddressLookupTableAddresses []string          `json:"addressLookupTableAddresses"`
	PrioritizationFeeLamports   int               `json:"prioritizationFeeLamports"`
	ComputeUnitLimit            int               `json:"computeUnitLimit"`
	PrioritizationType          struct {
		ComputeBudget struct {
			MicroLamports          int         `json:"microLamports"`
			EstimatedMicroLamports interface{} `json:"estimatedMicroLamports"`
		} `json:"computeBudget"`
	} `json:"prioritizationType"`
	SimulationSlot        int         `json:"simulationSlot"`
	DynamicSlippageReport interface{} `json:"dynamicSlippageReport"`
	SimulationError       struct {
		ErrorCode string `json:"errorCode"`
		Error     string `json:"error"`
	} `json:"simulationError"`
}
