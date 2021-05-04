package commands

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"math/big"
	"testing"

	"github.com/holiman/uint256"
	"github.com/ledgerwatch/turbo-geth/cmd/rpcdaemon/filters"
	"github.com/ledgerwatch/turbo-geth/common"
	"github.com/ledgerwatch/turbo-geth/common/u256"
	"github.com/ledgerwatch/turbo-geth/core/types"
	"github.com/ledgerwatch/turbo-geth/crypto"
	"github.com/ledgerwatch/turbo-geth/gointerfaces/txpool"
	"github.com/stretchr/testify/require"
)

func TestSendRawTransaction(t *testing.T) {
	t.Skip("Fix in the next PR")
	db, err := createTestKV()
	require.NoError(t, err)
	defer db.Close()
	conn := createTestGrpcConn()
	defer conn.Close()
	txPool := txpool.NewTxpoolClient(conn)
	ff := filters.New(context.Background(), nil, txPool)
	api := NewEthAPI(db, nil, txPool, 5000000, ff, nil)

	// Call GetTransactionReceipt for un-protected transaction
	var testKey, _ = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
	txn := transaction(40, 1000000, testKey)
	buf := bytes.NewBuffer(nil)
	err = txn.MarshalBinary(buf)
	require.NoError(t, err)

	txsCh := make(chan []types.Transaction, 1)
	defer close(txsCh)
	id := api.filters.SubscribePendingTxs(txsCh)
	defer api.filters.UnsubscribePendingTxs(id)

	_, err = api.SendRawTransaction(context.Background(), buf.Bytes())
	require.NoError(t, err)
	select {
	case <-txsCh:
	default:
		t.Fatalf("expected notification")

	}
}

func transaction(nonce uint64, gaslimit uint64, key *ecdsa.PrivateKey) types.Transaction {
	return pricedTransaction(nonce, gaslimit, u256.Num1, key)
}

func pricedTransaction(nonce uint64, gaslimit uint64, gasprice *uint256.Int, key *ecdsa.PrivateKey) types.Transaction {
	tx, _ := types.SignTx(types.NewTransaction(nonce, common.Address{}, uint256.NewInt().SetUint64(100), gaslimit, gasprice, nil), *types.LatestSignerForChainID(big.NewInt(1337)), key)
	return tx
}