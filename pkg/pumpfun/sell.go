package pump

import (
	"context"
	"fmt"
	"github.com/fape-labs/solana-agent-kit-go/pkg/pumpfun/pumpidl"
	"math/big"
	"strconv"

	"github.com/gagliardetto/solana-go"
	associatedtokenaccount "github.com/gagliardetto/solana-go/programs/associated-token-account"
	cb "github.com/gagliardetto/solana-go/programs/compute-budget"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	confirm "github.com/gagliardetto/solana-go/rpc/sendAndConfirmTransaction"
	"github.com/gagliardetto/solana-go/rpc/ws"
)

func SellToken(rpcClient *rpc.Client, wsClient *ws.Client, user solana.PrivateKey, mint solana.PublicKey, sellTokenAmount uint64, percentage float64, all bool) (string, error) {
	// create priority fee instructions
	culInst := cb.NewSetComputeUnitLimitInstruction(uint32(250000))
	cupInst := cb.NewSetComputeUnitPriceInstruction(uint64(10000))
	instructions := []solana.Instruction{
		culInst.Build(),
		cupInst.Build(),
	}
	// get sell instructions
	sellInstructions, err := getSellInstructions(rpcClient, user, mint, sellTokenAmount, percentage, all)
	if err != nil {
		return "", fmt.Errorf("failed to get sell instructions: %w", err)
	}
	instructions = append(instructions, sellInstructions)
	// get recent block hash
	recent, err := rpcClient.GetLatestBlockhash(context.TODO(), rpc.CommitmentFinalized)
	if err != nil {
		return "", fmt.Errorf("error while getting recent block hash: %w", err)
	}
	// create new transaction
	tx, err := solana.NewTransaction(
		instructions,
		recent.Value.Blockhash,
		solana.TransactionPayer(user.PublicKey()),
	)
	if err != nil {
		return "", fmt.Errorf("error while creating new transaction: %w", err)
	}
	_, err = tx.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			if user.PublicKey().Equals(key) {
				return &user
			}
			return nil
		},
	)
	if err != nil {
		return "", fmt.Errorf("can't sign transaction: %w", err)
	}

	// Send transaction, and wait for confirmation:
	sig, err := confirm.SendAndConfirmTransaction(
		context.TODO(),
		rpcClient,
		wsClient,
		tx,
	)
	if err != nil {
		return "", fmt.Errorf("can't send and confirm new transaction: %w", err)
	}
	return sig.String(), nil
}

// getSellInstructions is a function that returns the pumpidl.fun instructions to sell the token
func getSellInstructions(rpcClient *rpc.Client, user solana.PrivateKey, mint solana.PublicKey, sellTokenAmount uint64, percentage float64, all bool) (*pumpidl.Instruction, error) {
	ata, _, err := solana.FindAssociatedTokenAddress(
		user.PublicKey(),
		mint,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to derive associated token account: %w", err)
	}
	if all {
		tokenAccounts, err := rpcClient.GetTokenAccountBalance(context.TODO(), ata, rpc.CommitmentConfirmed)
		if err != nil {
			return nil, fmt.Errorf("can't get amount of token in balance: %w", err)
		}
		amount, err := strconv.Atoi(tokenAccounts.Value.Amount)
		if err != nil {
			return nil, fmt.Errorf("can't convert token amount to integer: %w", err)
		}
		sellTokenAmount = uint64(amount)
	}
	bondingCurveData, err := getBondingCurveAndAssociatedBondingCurve(mint)
	if err != nil {
		return nil, fmt.Errorf("can't get bonding curve data: %w", err)
	}
	bondingCurve, err := fetchBondingCurve(rpcClient, bondingCurveData.BondingCurve)
	if err != nil {
		return nil, fmt.Errorf("can't fetch bonding curve: %w", err)
	}
	minSolOutput := calculateSellQuote(sellTokenAmount, bondingCurve, percentage)
	minSolOutput = big.NewInt(0)
	sellInstr := pumpidl.NewSellInstruction(
		sellTokenAmount,
		minSolOutput.Uint64(),
		globalPumpFunAddress,
		pumpFunFeeRecipient,
		mint,
		bondingCurveData.BondingCurve,
		bondingCurveData.AssociatedBondingCurve,
		ata,
		user.PublicKey(),
		system.ProgramID,
		associatedtokenaccount.ProgramID,
		token.ProgramID,
		pumpFunEventAuthority,
		pumpidl.ProgramID,
	)
	sell, err := sellInstr.ValidateAndBuild()
	if err != nil {
		return nil, fmt.Errorf("can't validate and build sell instruction: %w", err)
	}
	return sell, nil
}

// calculateSellQuote calculates how many SOL should be received for selling a specific amount of tokens, given a specific amount of token, bonding curve data, and percentage.
// tokenAmount is the amount of token you want to sell
// bondingCurve is the bonding curve data, that will help to calculate the number of sol to get
// percentage is the slippage, 0.98 means 2% slippage
func calculateSellQuote(tokenAmount uint64, bondingCurve *BondingCurveData, percentage float64) *big.Int {
	amount := big.NewInt(int64(tokenAmount))

	// Clone bonding curve data to avoid mutations
	virtualSolReserves := new(big.Int).Set(bondingCurve.VirtualSolReserves)
	virtualTokenReserves := new(big.Int).Set(bondingCurve.VirtualTokenReserves)

	// Compute the new virtual reserves
	x := new(big.Int).Mul(virtualSolReserves, amount)
	y := new(big.Int).Add(virtualTokenReserves, amount)
	a := new(big.Int).Div(x, y)
	percentageMultiplier := big.NewFloat(percentage)
	sol := new(big.Float).SetInt(a)
	number := new(big.Float).Mul(sol, percentageMultiplier)
	final, _ := number.Int(nil)
	return final
}
