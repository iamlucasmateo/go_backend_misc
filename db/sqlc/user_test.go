package db

import (
	"context"
	"testing"

	"github.com/go_backend_misc/util"
	"github.com/stretchr/testify/require"
)

func createRandomUser(userSuffix string) (User, CreateUserParams, error) {
	username := util.RandomString(7) + userSuffix
	arg := CreateUserParams{
		Username:       username,
		HashedPassword: util.RandomString(10),
		FullName:       util.RandomString(10) + " " + util.RandomString(10),
		Email:          util.RandomEmail(username),
	}
	user, err := testQueries.CreateUser(context.Background(), arg)
	return user, arg, err
}

func TestCreateUser(t *testing.T) {
	user, arg, err := createRandomUser("_test_create_user")
	if err != nil {
		t.Errorf("error creating user: %v", err)
	}
	require.NoError(t, err)
	require.NotEmpty(t, user)

	require.Equal(t, arg.Username, user.Username)
	require.Equal(t, arg.Email, user.Email)
	require.Equal(t, arg.HashedPassword, user.HashedPassword)
	require.Equal(t, arg.FullName, user.FullName)
	require.NotZero(t, user.CreatedAt)
	require.True(t, user.PasswordChangedAt.IsZero())

}

func TestGetUser(t *testing.T) {
	createdUser, _, _ := createRandomUser("_test_get_user")
	retrievedUser, err := testQueries.GetUserByUsername(context.Background(), createdUser.Username)
	require.NoError(t, err)
	require.Equal(t, createdUser.Username, retrievedUser.Username)
	require.Equal(t, createdUser.Email, retrievedUser.Email)
	require.Equal(t, createdUser.HashedPassword, retrievedUser.HashedPassword)
	require.Equal(t, createdUser.FullName, retrievedUser.FullName)
	require.Equal(t, createdUser.CreatedAt, retrievedUser.CreatedAt)
	require.Equal(t, createdUser.PasswordChangedAt, retrievedUser.PasswordChangedAt)
}
