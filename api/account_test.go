package api

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"

	db "github.com/TheRanomial/bank_server/db/sqlc"
	"github.com/TheRanomial/bank_server/util"
	"github.com/stretchr/testify/require"
)



func randomAccount() db.Account{
	return db.Account{
		ID: util.RandomInt(1,1000),
		Owner: util.RandomOwner(),
		Balance: util.RandomBalance(),
		Currency: util.RandomCurrency(),
	}
}

func requireBodyMatchAccount(t *testing.T, body *bytes.Buffer, account db.Account) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotAccount db.Account
	err = json.Unmarshal(data, &gotAccount)
	require.NoError(t, err)
	require.Equal(t, account, gotAccount)
}
