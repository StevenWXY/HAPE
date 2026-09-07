package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/StevenWXY/HAPE/internal/domain"
	"github.com/StevenWXY/HAPE/internal/store"
)

// BNBExecution uses a read-only RPC connection. Signing keys stay in the
// separate executor process and never enter the API server or browser.
type BNBExecution struct {
	RPCURL        string
	ChainID       string
	Token         string
	Treasury      string
	Decimals      int
	Confirmations int
	Client        *http.Client
}

func (b *BNBExecution) call(method string, params any, output any) error {
	body, err := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": 1, "method": method, "params": params})
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, b.RPCURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := b.Client
	if client == nil {
		client = http.DefaultClient
	}
	response, err := client.Do(req)
	if err != nil {
		return apiError(http.StatusBadGateway, "bnb_rpc_unavailable", "The BNB RPC is unavailable")
	}
	defer response.Body.Close()
	var envelope struct {
		Result json.RawMessage `json:"result"`
		Error  json.RawMessage `json:"error"`
	}
	if response.StatusCode != 200 || json.NewDecoder(io.LimitReader(response.Body, 2<<20)).Decode(&envelope) != nil || (len(envelope.Error) > 0 && string(envelope.Error) != "null") || len(envelope.Result) == 0 || string(envelope.Result) == "null" {
		return apiError(http.StatusBadGateway, "bnb_receipt_pending", "The RPC has not returned a confirmed result")
	}
	if err := json.Unmarshal(envelope.Result, output); err != nil {
		return apiError(http.StatusBadGateway, "bnb_invalid_response", "The RPC returned an invalid response")
	}
	return nil
}

func hexNumber(value string) (*big.Int, bool) {
	if !strings.HasPrefix(value, "0x") {
		return nil, false
	}
	return new(big.Int).SetString(value[2:], 16)
}

func (b *BNBExecution) validateChain() error {
	if b.ChainID != "0x38" || !validWalletAddress(b.Token) || !validWalletAddress(b.Treasury) || b.Decimals < 0 || b.Decimals > 18 || b.Confirmations < 3 {
		return apiError(http.StatusServiceUnavailable, "bnb_execution_not_configured", "Real CLIP settlement requires a mainnet contract, treasury and at least three confirmations")
	}
	var chain string
	if err := b.call("eth_chainId", []any{}, &chain); err != nil {
		return err
	}
	if NormalizeBNBChain(chain) != b.ChainID {
		return apiError(http.StatusConflict, "bnb_chain_mismatch", "The RPC chain does not match the configured CLIP network")
	}
	return nil
}

func (s *Service) ConfigureBNBExecution(b *BNBExecution) error {
	if err := b.validateChain(); err != nil {
		return err
	}
	var decimals string
	if err := b.call("eth_call", []any{map[string]string{"to": b.Token, "data": "0x313ce567"}, "latest"}, &decimals); err != nil {
		return err
	}
	d, ok := hexNumber(decimals)
	if !ok || !d.IsInt64() || d.Int64() != int64(b.Decimals) {
		return apiError(http.StatusConflict, "bnb_token_decimals_mismatch", "The token decimals do not match configuration")
	}
	s.airdropVerifier = b.Verify
	s.bnbExecution = b
	for _, item := range s.Snapshot().Airdrops {
		if item.Status == "submitted" && validTransactionHash(item.TxHash) {
			_, _ = s.UpdateAirdrop(item.ID, AirdropResultInput{Status: "confirmed", TxHash: item.TxHash, ExecutorRef: item.ExecutorRef})
		}
	}
	return s.SyncBNBTreasury(b)
}

func (s *Service) SyncBNBTreasury(b *BNBExecution) error {
	if err := b.validateChain(); err != nil {
		return err
	}
	var raw string
	if err := b.call("eth_call", []any{map[string]string{"to": b.Token, "data": "0x70a08231000000000000000000000000" + strings.TrimPrefix(strings.ToLower(b.Treasury), "0x")}, "latest"}, &raw); err != nil {
		return err
	}
	amount, ok := hexNumber(raw)
	if !ok {
		return apiError(http.StatusBadGateway, "bnb_invalid_balance", "Invalid CLIP balance")
	}
	amount.Quo(amount, new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(b.Decimals)), nil))
	if !amount.IsInt64() || amount.Int64() > 1000000000 {
		return apiError(http.StatusBadGateway, "bnb_invalid_balance", "CLIP balance exceeds supply bounds")
	}
	return s.store.Update(func(state *store.State) error {
		if state.CLIPTreasury.ContractAddress != "" && state.CLIPTreasury.ContractAddress != "not-published" && (!strings.EqualFold(state.CLIPTreasury.ContractAddress, b.Token) || !strings.EqualFold(state.CLIPTreasury.PlatformWallet, b.Treasury)) {
			return apiError(http.StatusConflict, "bnb_treasury_changed", "A funded ledger cannot change its token or treasury")
		}
		reserved := 0
		for _, migration := range state.ExternalMigrations {
			reserved += migration.TreasuryReserved
		}
		available := int(amount.Int64()) - state.CLIPTreasury.LedgerOutstanding - reserved
		if available < 0 {
			available = 0
		}
		state.CLIPTreasury.TreasuryBalance = available
		state.CLIPTreasury.ContractAddress, state.CLIPTreasury.PlatformWallet, state.CLIPTreasury.Network = b.Token, b.Treasury, b.ChainID
		state.CLIPTreasury.MintStatus = "onchain-verified"
		return nil
	})
}

func (b *BNBExecution) Verify(item domain.AirdropRecord, input AirdropResultInput) error {
	if err := b.validateChain(); err != nil {
		return err
	}
	if item.ChainID != b.ChainID {
		return apiError(http.StatusConflict, "bnb_chain_mismatch", "The airdrop uses another chain")
	}
	if input.Status == "submitted" {
		return nil
	}
	var receipt struct {
		TransactionHash string `json:"transactionHash"`
		Status          string `json:"status"`
		BlockNumber     string `json:"blockNumber"`
		BlockHash       string `json:"blockHash"`
		Logs            []struct {
			Address string   `json:"address"`
			Topics  []string `json:"topics"`
			Data    string   `json:"data"`
			Removed bool     `json:"removed"`
		} `json:"logs"`
	}
	if err := b.call("eth_getTransactionReceipt", []any{input.TxHash}, &receipt); err != nil {
		return err
	}
	if !strings.EqualFold(receipt.TransactionHash, input.TxHash) {
		return apiError(http.StatusConflict, "bnb_receipt_mismatch", "The transaction receipt does not match")
	}
	var head string
	if err := b.call("eth_blockNumber", []any{}, &head); err != nil {
		return err
	}
	block, ok := hexNumber(receipt.BlockNumber)
	latest, valid := hexNumber(head)
	if !ok || !valid || new(big.Int).Sub(latest, block).Cmp(big.NewInt(int64(b.Confirmations-1))) < 0 {
		return apiError(http.StatusConflict, "bnb_receipt_pending", "The transaction needs more confirmations")
	}
	var canonical struct {
		Hash string `json:"hash"`
	}
	if err := b.call("eth_getBlockByNumber", []any{receipt.BlockNumber, false}, &canonical); err != nil {
		return err
	}
	if receipt.BlockHash == "" || !strings.EqualFold(receipt.BlockHash, canonical.Hash) {
		return apiError(http.StatusConflict, "bnb_receipt_pending", "The transaction is not in the canonical chain")
	}
	if input.Status == "failed" {
		if receipt.Status != "0x0" {
			return apiError(http.StatusConflict, "bnb_transaction_not_failed", "A successful or pending transfer cannot be refunded")
		}
		return nil
	}
	if receipt.Status != "0x1" {
		return apiError(http.StatusConflict, "bnb_transaction_failed", "The CLIP transfer reverted")
	}
	amount := new(big.Int).Mul(big.NewInt(int64(item.Amount)), new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(b.Decimals)), nil))
	from := "0x000000000000000000000000" + strings.TrimPrefix(strings.ToLower(b.Treasury), "0x")
	to := "0x000000000000000000000000" + strings.TrimPrefix(strings.ToLower(item.WalletAddress), "0x")
	for _, log := range receipt.Logs {
		if log.Removed || !strings.EqualFold(log.Address, b.Token) || len(log.Topics) != 3 || !strings.EqualFold(log.Topics[0], "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef") || !strings.EqualFold(log.Topics[1], from) || !strings.EqualFold(log.Topics[2], to) {
			continue
		}
		actual, ok := hexNumber(log.Data)
		if ok && actual.Cmp(amount) == 0 {
			return nil
		}
	}
	return apiError(http.StatusConflict, "bnb_transfer_mismatch", fmt.Sprintf("No matching transfer of %d CLIP to the recipient", item.Amount))
}
