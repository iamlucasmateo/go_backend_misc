package token

import (
	"testing"
	"time"

	"github.com/go_backend_misc/util"
	"github.com/stretchr/testify/require"
)

func CheckTokenMaker(t *testing.T, tokenMaker TokenMaker) {
	username := util.RandomOwner()
	duration := time.Minute

	issuedAt := time.Now()
	expiredAt := issuedAt.Add(duration)

	token, err := tokenMaker.CreateToken(username, duration)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	payload, err := tokenMaker.VerifyToken(token)
	require.NoError(t, err)
	require.NotEmpty(t, payload)

	require.NotZero(t, payload.ID)
	require.Equal(t, username, payload.Username)
	require.WithinDuration(t, issuedAt, payload.IssuedAt, time.Second)
	require.WithinDuration(t, expiredAt, payload.ExpiredAt, time.Second)
}

func CheckExpiredToken(t *testing.T, tokenMaker TokenMaker) {
	token, err := tokenMaker.CreateToken(util.RandomOwner(), -time.Minute)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	payload, err := tokenMaker.VerifyToken(token)
	require.Error(t, err)
	require.EqualError(t, err, ErrExpiredToken.Error())
	require.Nil(t, payload)
}
