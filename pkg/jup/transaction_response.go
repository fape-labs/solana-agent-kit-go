package jup

type JupTxResponse struct {
	SwapTransaction           string `json:"swapTransaction"`
	LastValidBlockHeight      int    `json:"lastValidBlockHeight"`
	PrioritizationFeeLamports int    `json:"prioritizationFeeLamports"`
	ComputeUnitLimit          int    `json:"computeUnitLimit"`
	PrioritizationType        struct {
		ComputeBudget struct {
			MicroLamports          int         `json:"microLamports"`
			EstimatedMicroLamports interface{} `json:"estimatedMicroLamports"`
		} `json:"computeBudget"`
	} `json:"prioritizationType"`
	DynamicSlippageReport interface{}      `json:"dynamicSlippageReport"`
	SimulationError       *SimulationError `json:"simulationError"`
}
type SimulationError struct {
	ErrorCode string `json:"errorCode"`
	Error     string `json:"error"`
}
