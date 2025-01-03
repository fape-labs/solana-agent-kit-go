package pump

import (
	"context"
	"fmt"
	"github.com/fape-labs/solana-agent-kit-go/pkg/pumpfun/pumpidl"

	// General solana packages.
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	confirm "github.com/gagliardetto/solana-go/rpc/sendAndConfirmTransaction"
	"github.com/gagliardetto/solana-go/rpc/ws"

	// This package interacts with the Compute Budget program, allowing
	// to easily get instruction to set compute budget limit/price for example.
	cb "github.com/gagliardetto/solana-go/programs/compute-budget"
	// This package interacts with the Solana system program, allowing
	// to transfer solana for example.
	"github.com/gagliardetto/solana-go/programs/system"
	// This package interacts with the Token program, allowing
	// to create a token for example.
	"github.com/gagliardetto/solana-go/programs/token"
	// This package interacts with the Associated Token Account program
	// allowing to create/close an associated token account for example.
	associatedtokenaccount "github.com/gagliardetto/solana-go/programs/associated-token-account"
)

// Contains commonly used addresses with the pumpidl.fun program, that are not present
// in the generated code, from its IDL file.
var (
	// Global account address for pumpidl.fun
	globalPumpFunAddress = solana.MustPublicKeyFromBase58("4wTV1YmiEkRvAtNtsSGPtUrqRYQMe5SKy2uB4Jjaxnjf")
	// pumpidl.fun mint authority
	pumpFunMintAuthority = solana.MustPublicKeyFromBase58("TSLvdd1pWpHVjahSpsvCXUbgwsL3JAcvokwaKt1eokM")
	// pumpidl.fun event authority
	pumpFunEventAuthority = solana.MustPublicKeyFromBase58("Ce6TQqeHC9p8KetsN6JsjHK7UTZk7nasjjnr7XxXp9F1")
	// pumpidl.fun fee recipient
	pumpFunFeeRecipient = solana.MustPublicKeyFromBase58("CebN5WGQ4jvEPvsVU4EoHEpgzq1VV7AbicfhtW4xC9iM")
)

// SetDevnetMode sets the pumpidl.fun program addresses to the devnet addresses.
// It is important to call this function if you are using the devnet.
func SetDevnetMode() {
	// This is the address you want to use as pumpidl.fun fee recipient on devnet, otherwise, it
	// will not work, as the official pumpidl.fun fee recipient account is not initialized on devnet.
	// I know, using global variables is ugly, but passing this address around everywhere
	// (in BuyToken / SellToken), while it's actually a constant on mainnet is even uglier,
	// considering that there is no other difference.
	pumpFunFeeRecipient = solana.MustPublicKeyFromBase58("68yFSZxzLWJXkxxRGydZ63C6mHx1NLEDWmwN9Lb5yySg")
}

type BondingCurvePublicKeys struct {
	BondingCurve           solana.PublicKey
	AssociatedBondingCurve solana.PublicKey
}

// getBondingCurveAndAssociatedBondingCurve returns the bonding curve and associated bonding curve, in a structured format.
func getBondingCurveAndAssociatedBondingCurve(mint solana.PublicKey) (*BondingCurvePublicKeys, error) {
	// Derive bonding curve address.
	// define the seeds used to derive the PDA
	// getProgramDerivedAddress equivalent.
	seeds := [][]byte{
		[]byte("bonding-curve"),
		mint.Bytes(),
	}
	bondingCurve, _, err := solana.FindProgramAddress(seeds, pumpidl.ProgramID)
	if err != nil {
		return nil, fmt.Errorf("failed to derive bonding curve address: %w", err)
	}
	// Derive associated bonding curve address.
	associatedBondingCurve, _, err := solana.FindAssociatedTokenAddress(
		bondingCurve,
		mint,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to derive associated bonding curve address: %w", err)
	}
	return &BondingCurvePublicKeys{
		BondingCurve:           bondingCurve,
		AssociatedBondingCurve: associatedBondingCurve,
	}, nil
}

func getComputUnitPriceInstr(rpcClient *rpc.Client, user solana.PrivateKey) (*cb.SetComputeUnitPrice, error) {
	// create priority fee instructions
	out, err := rpcClient.GetRecentPrioritizationFees(context.TODO(), solana.PublicKeySlice{user.PublicKey(), pumpidl.ProgramID, pumpFunMintAuthority, globalPumpFunAddress, solana.TokenMetadataProgramID, system.ProgramID, token.ProgramID, associatedtokenaccount.ProgramID, solana.SysVarRentPubkey, pumpFunEventAuthority})
	if err != nil {
		return nil, fmt.Errorf("failed to get recent prioritization fees: %w", err)
	}
	var median uint64
	length := uint64(len(out))
	for _, fee := range out {
		median = fee.PrioritizationFee
	}
	median /= length
	cupInst := cb.NewSetComputeUnitPriceInstruction(median)
	return cupInst, nil
}

func CreateToken(rpcClient *rpc.Client, wsClient *ws.Client, user solana.PrivateKey, mint *solana.Wallet, name string, symbol string, uri string, buyAmountSol float64, percentage float64) (string, error) {
	bondingCurveData, err := getBondingCurveAndAssociatedBondingCurve(mint.PublicKey())
	if err != nil {
		return "", fmt.Errorf("failed to get bonding curve and associated bonding curve: %w", err)
	}
	// Get token metadata address
	metadata, _, err := solana.FindTokenMetadataAddress(mint.PublicKey())
	if err != nil {
		return "", fmt.Errorf("can't find token metadata address: %w", err)
	}

	// Default pumpidl.fun compute limit is 250k, so we set the same here.
	culInst := cb.NewSetComputeUnitLimitInstruction(uint32(250000))
	cupInst, err := getComputUnitPriceInstr(rpcClient, user)
	if err != nil {
		return "", fmt.Errorf("failed to get compute unit price instructions: %w", err)
	}
	// Create the pump fun instruction
	instr := pumpidl.NewCreateInstruction(
		name,
		symbol,
		uri,
		mint.PublicKey(),
		pumpFunMintAuthority,
		bondingCurveData.BondingCurve,
		bondingCurveData.AssociatedBondingCurve,
		globalPumpFunAddress,
		solana.TokenMetadataProgramID,
		metadata,
		user.PublicKey(),
		system.ProgramID,
		token.ProgramID,
		associatedtokenaccount.ProgramID,
		solana.SysVarRentPubkey,
		pumpFunEventAuthority,
		pumpidl.ProgramID,
	)
	instruction := instr.Build()
	// get recent block hash
	recent, err := rpcClient.GetLatestBlockhash(context.TODO(), rpc.CommitmentFinalized)
	if err != nil {
		return "", fmt.Errorf("error while getting recent block hash: %w", err)
	}
	instructions := []solana.Instruction{
		culInst.Build(),
		cupInst.Build(),
		instruction,
	}
	// get buy instructions
	if buyAmountSol > 0 {
		buyInstructions, err := getBuyInstructions(rpcClient, mint.PublicKey(), user.PublicKey(), SolToLamp(buyAmountSol), percentage)
		if err != nil {
			return "", fmt.Errorf("failed to get buy instructions: %w", err)
		}
		instructions = append(instructions, buyInstructions...)
	}
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
			if mint.PublicKey().Equals(key) {
				return &mint.PrivateKey
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
