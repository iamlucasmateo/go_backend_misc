package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go_backend_misc/token"
	"github.com/stretchr/testify/require"
)

func addAuthorization(t *testing.T, request *http.Request, tokenMaker token.TokenMaker, username string) {
	token, err := tokenMaker.CreateToken(username, time.Minute)
	require.NoError(t, err)
	request.Header.Set("Authorization", "bearer "+token)
}

func TestAuthMiddleware(t *testing.T) {
	type authTestCase struct {
		name          string
		setupAuth     func(request *http.Request, tokenMaken token.TokenMaker)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}
	okTestCase := authTestCase{
		name: "OK",
		setupAuth: func(request *http.Request, tokenMaker token.TokenMaker) {
			addAuthorization(t, request, tokenMaker, "test")
		},
		checkResponse: func(recorder *httptest.ResponseRecorder) {
			require.Equal(t, http.StatusOK, recorder.Code)
		},
	}
	noHeaderTestCase := authTestCase{
		name:      "No authorization header",
		setupAuth: func(request *http.Request, tokenMaker token.TokenMaker) {},
		checkResponse: func(recorder *httptest.ResponseRecorder) {
			require.Equal(t, http.StatusUnauthorized, recorder.Code)

			var content map[string]string
			if err := json.Unmarshal(recorder.Body.Bytes(), &content); err != nil {
				t.Errorf("error decoding response body: %v", err)
			}
			expected := "Authorization header is not provided"
			require.Equal(t, expected, content["error"])

		},
	}
	invalidResponseTestCase := authTestCase{
		name: "Invalid auth header",
		setupAuth: func(request *http.Request, tokenMaker token.TokenMaker) {
			request.Header.Set("Authorization", "three auth fields")
		},
		checkResponse: func(recorder *httptest.ResponseRecorder) {
			require.Equal(t, http.StatusUnauthorized, recorder.Code)

			var content map[string]string
			if err := json.Unmarshal(recorder.Body.Bytes(), &content); err != nil {
				t.Errorf("error decoding response body: %v", err)
			}
			expected := "invalid authorization header"
			require.Equal(t, expected, content["error"])
		},
	}

	unsupportedAuthTestCase := authTestCase{
		name: "Unsupported Auth type",
		setupAuth: func(request *http.Request, tokenMaker token.TokenMaker) {
			request.Header.Set("Authorization", "not_bearer Token")
		},
		checkResponse: func(recorder *httptest.ResponseRecorder) {
			require.Equal(t, http.StatusUnauthorized, recorder.Code)

			var content map[string]string
			json.Unmarshal(recorder.Body.Bytes(), &content)
			expected := "unsupported authorization type not_bearer"
			require.Equal(t, expected, content["error"])
		},
	}

	verifyTokenErrorTestCase := authTestCase{
		name: "Verify Token error",
		setupAuth: func(request *http.Request, tokenMaker token.TokenMaker) {
			_, err := tokenMaker.CreateToken("test", time.Minute)
			require.NoError(t, err)
			wrongToken := "some_wrong_token"
			request.Header.Set("Authorization", "bearer "+wrongToken)
		},
		checkResponse: func(recorder *httptest.ResponseRecorder) {
			require.Equal(t, http.StatusUnauthorized, recorder.Code)

			var content map[string]string
			json.Unmarshal(recorder.Body.Bytes(), &content)
			expected := "invalid token"
			require.Equal(t, expected, content["error"])
		},
	}

	testCases := []authTestCase{
		okTestCase, noHeaderTestCase, invalidResponseTestCase,
		unsupportedAuthTestCase, verifyTokenErrorTestCase,
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := newTestServer(t, nil)
			server.router.GET(
				"/auth",
				authMiddleware(server.tokenMaker),
				func(ctx *gin.Context) { ctx.JSON(http.StatusOK, gin.H{}) },
			)
			recorder := httptest.NewRecorder()
			request, err := http.NewRequest(http.MethodGet, "/auth", nil)
			require.NoError(t, err)

			tc.setupAuth(request, server.tokenMaker)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}
