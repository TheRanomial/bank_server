package db

import (
	"context"
	"testing"

	"github.com/TheRanomial/bank_server/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func CreateRandomAccount(t *testing.T) *Account{
	user := CreateRandomUser(t)

	arg:=CreateAccountParams{
		Owner: 		user.Username,
		Balance: 	util.RandomBalance(),
		Currency:	util.RandomCurrency(),
	}

	account,err:=testQueries.CreateAccount(context.Background(),arg)
	assert.Nil(t,err)
	require.NotEmpty(t,account)
	require.Equal(t,arg.Owner,account.Owner)
	require.Equal(t,arg.Balance,account.Balance)
	require.Equal(t,arg.Currency,account.Currency)
	require.NotZero(t,account.ID)
	require.NotZero(t,account.CreatedAt)

	return &account
}

func TestCreateAccount(t *testing.T) {
	CreateRandomAccount(t)
}

func TestGetAccount(t *testing.T){
	acc1:=CreateRandomAccount(t)
	acc2,err:=testQueries.GetAccount(context.Background(),acc1.ID)
	require.NoError(t,err)

	require.Equal(t,acc1.Owner,acc2.Owner)
	require.Equal(t,acc1.Balance,acc2.Balance)
}

func TestUpdateAccount(t *testing.T){
	acc1:=CreateRandomAccount(t)

	arg:=UpdateAccountParams{
		ID: acc1.ID,
		Balance: util.RandomBalance(),
	}

	err:=testQueries.UpdateAccount(context.Background(),arg)
	require.NoError(t,err)
}

func TestDeleteAccount(t *testing.T){
	acc:=CreateRandomAccount(t)
	err:=testQueries.DeleteAccount(context.Background(),acc.ID)
	require.NoError(t,err)

	acc2,err:=testQueries.GetAccount(context.Background(),acc.ID)
	require.Error(t,err)
	require.Empty(t,acc2)
}