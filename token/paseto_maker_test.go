package token

import (
	"testing"

	"github.com/go_backend_misc/util"
	"github.com/stretchr/testify/require"
)

func TestPasetoMaker(t *testing.T) {
	maker, err := NewPasetoMaker(util.RandomString(32))
	require.NoError(t, err)
	CheckTokenMaker(t, maker)

}

func TestExpiredPasetoToken(t *testing.T) {
	maker, err := NewPasetoMaker(util.RandomString(32))
	require.NoError(t, err)
	CheckExpiredToken(t, maker)
}

func TestInvalidPasetoToken(t *testing.T) {
	maker, err := NewPasetoMaker(util.RandomString(32))
	require.NoError(t, err)

	payload, err := maker.VerifyToken("some_wrong_token")
	require.Error(t, err)
	require.EqualError(t, err, ErrInvalidToken.Error())
	require.Nil(t, payload)
}
