package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/StevenWXY/HAPE/internal/domain"
)

func TestBNBReceiptRequiresMatchingConfirmedTransfer(t *testing.T) {
	wallet := "0x1111111111111111111111111111111111111111"
	treasury := "0x2222222222222222222222222222222222222222"
	token := "0x3333333333333333333333333333333333333333"
	hash := "0x" + strings.Repeat("a", 64)
	for _, scenario := range []string{"valid", "wrong-chain", "wrong-wallet", "wrong-amount", "wrong-contract", "pending", "reverted", "reorg"} {
		t.Run(scenario, func(t *testing.T) {
			b := &BNBExecution{RPCURL: "https://rpc.example", ChainID: "0x38", Token: token, Treasury: treasury, Decimals: 18, Confirmations: 3}
			b.Client = &http.Client{Transport: integrationRoundTripper(func(request *http.Request) (*http.Response, error) {
				var call struct {
					Method string `json:"method"`
				}
				_ = json.NewDecoder(request.Body).Decode(&call)
				var result any
				switch call.Method {
				case "eth_chainId":
					result = "0x38"
					if scenario == "wrong-chain" {
						result = "0x61"
					}
				case "eth_blockNumber":
					result = "0x12"
					if scenario == "pending" {
						result = "0x11"
					}
				case "eth_getBlockByNumber":
					result = map[string]any{"hash": "0xblock"}
					if scenario == "reorg" {
						result = map[string]any{"hash": "0xother"}
					}
				case "eth_getTransactionReceipt":
					recipient, contract, amount, status := wallet, token, int64(1000000000000000000), "0x1"
					if scenario == "wrong-wallet" {
						recipient = treasury
					}
					if scenario == "wrong-contract" {
						contract = treasury
					}
					if scenario == "wrong-amount" {
						amount = 2
					}
					if scenario == "reverted" {
						status = "0x0"
					}
					result = map[string]any{"transactionHash": hash, "blockNumber": "0x10", "blockHash": "0xblock", "status": status, "logs": []any{map[string]any{"address": contract, "topics": []string{"0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef", "0x000000000000000000000000" + treasury[2:], "0x000000000000000000000000" + recipient[2:]}, "data": fmt.Sprintf("0x%064x", amount)}}}
				default:
					t.Fatalf("unexpected RPC %s", call.Method)
				}
				body, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": 1, "result": result})
				return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(string(body)))}, nil
			})}
			err := b.Verify(domain.AirdropRecord{ChainID: "0x38", WalletAddress: wallet, Amount: 1}, AirdropResultInput{Status: "confirmed", TxHash: hash})
			if (err == nil) != (scenario == "valid") {
				t.Fatalf("receipt check err=%v", err)
			}
		})
	}
}
