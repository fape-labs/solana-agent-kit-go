package jup

import (
	"encoding/json"
	"fmt"
	"github.com/fape-labs/solana-agent-kit-go/pkg/solanaclient"
	"github.com/gagliardetto/solana-go"
	"github.com/valyala/fasthttp"
)

type JupClient struct {
	BaseUrl string
}

func (jc *JupClient) Swap(client *solanaclient.Client, signer solana.PrivateKey, inputMint solana.PublicKey, outputMint solana.PublicKey, amount uint64, slippageBps int) (*solana.Signature, error) {
	qr, err := jc.GetQuote(inputMint, outputMint, amount, slippageBps)
	if err != nil {
		return nil, err
	}

	sr, err := jc.GetSwapTx(signer, qr)
	if err != nil {
		return nil, err
	}

	return client.SignAndSendTx(sr.SwapTransaction)
}

func (jc *JupClient) GetQuote(inputMint solana.PublicKey, outputMint solana.PublicKey, amount uint64, slippageBps int) (*QuoteReseponse, error) {
	req := fasthttp.AcquireRequest()
	req.SetRequestURI(fmt.Sprintf("%s/quote?inputMint=%s&outputMint=%s&amount=%d&slippageBps=%d&onlyDirectRoutes=false", jc.BaseUrl, inputMint.String(), outputMint.String(), amount, slippageBps))
	resp := fasthttp.AcquireResponse()

	defer fasthttp.ReleaseResponse(resp)

	err := fasthttp.Do(req, resp)
	if err != nil {
		return nil, err
	}
	fasthttp.ReleaseRequest(req)

	//fmt.Println(string(resp.Body()))

	juqQuoteResp := &QuoteReseponse{}
	err = json.Unmarshal(resp.Body(), juqQuoteResp)
	if err != nil {
		return nil, err
	}
	defer fasthttp.ReleaseResponse(resp)

	return juqQuoteResp, nil
}

func (jc *JupClient) GetSwapTx(signer solana.PrivateKey, q *QuoteReseponse) (*JupTxResponse, error) {
	swapReqObj := map[string]any{
		"quoteResponse":             q,
		"userPublicKey":             signer.PublicKey().String(),
		"wrapAndUnwrapSol":          true,
		"dynamicComputeUnitLimit":   true,
		"prioritizationFeeLamports": 20000,
		"useSharedAccounts":         false,
	}

	swapReq := fasthttp.AcquireRequest()
	swapReq.SetRequestURI(fmt.Sprintf("%s/swap", jc.BaseUrl))
	swapReq.Header.SetContentType("application/json")
	swapReq.Header.SetMethod(fasthttp.MethodPost)
	httpBody, err := json.Marshal(swapReqObj)
	swapReq.SetBodyRaw(httpBody)

	swapResp := fasthttp.AcquireResponse()
	err = fasthttp.Do(swapReq, swapResp)
	if err != nil {
		return nil, err
	}

	r := &JupTxResponse{}
	err = json.Unmarshal(swapResp.Body(), r)
	if err != nil {
		return nil, err
	}
	if r.SwapTransaction == "" {
		return nil, fmt.Errorf("failed to get swap tx from jup %s", string(swapResp.Body()))
	}

	return r, nil
}

func (jc *JupClient) GetSwapIx(signer solana.PrivateKey, q *QuoteReseponse) (*InstructionResponse, error) {
	//
	swapReqObj := map[string]any{
		"quoteResponse":             q,
		"userPublicKey":             signer.PublicKey().String(),
		"wrapAndUnwrapSol":          true,
		"dynamicComputeUnitLimit":   true,
		"prioritizationFeeLamports": 20000,
	}

	swapReq := fasthttp.AcquireRequest()
	swapReq.SetRequestURI(fmt.Sprintf("%/swap-instructions", jc.BaseUrl))
	swapReq.Header.SetContentType("application/json")
	swapReq.Header.SetMethod(fasthttp.MethodPost)
	httpBody, err := json.Marshal(swapReqObj)
	swapReq.SetBodyRaw(httpBody)

	swapResp := fasthttp.AcquireResponse()
	err = fasthttp.Do(swapReq, swapResp)
	if err != nil {
		return nil, err
	}

	r := &InstructionResponse{}
	err = json.Unmarshal(swapResp.Body(), r)
	if err != nil {
		return nil, err
	}

	return r, nil
}
