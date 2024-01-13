package tests

import (
	"gRPC-server/tests/suite"
	grpcv1 "github.com/Yankeegohome/protos/gen/go/gRPC-S"
	"github.com/brianvoe/gofakeit"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

const (
	emptyAppID     = 0
	appID          = 1
	appSecret      = "tests-secret"
	passDefaultLen = 10
)

func TestRegisterLogin_Login_HappyPath(t *testing.T) {
	ctx, st := suite.New(t)
	email := gofakeit.Email()
	pass := randomFakePassword()

	respReg, err := st.AuthClient.Register(ctx, &grpcv1.RegisterRequest{
		Email:    email,
		Password: pass,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, respReg.GetUserId())

	respLogin, err := st.AuthClient.Login(ctx, &grpcv1.LoginRequest{
		Email:    email,
		Password: pass,
		AppId:    appID,
	})
	require.NoError(t, err)

	token := respLogin.GetToken()
	require.NotEmpty(t, token)
	loginTime := time.Now()
	//tokenParsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
	//	return []byte(appSecret), nil
	//})
	tokenParsed, err := jwt.Parse(token, nil)
	require.NoError(t, err)
	claims, ok := tokenParsed.Claims.(jwt.MapClaims)
	assert.True(t, ok)

	assert.Equal(t, respReg.GetUserId(), int64(claims["uid"].(float64)))
	assert.Equal(t, email, claims["email"].(string))
	assert.Equal(t, appID, int(claims["app_id"].(float64)))

	const deltaSeconds = 1

	assert.InDelta(t, loginTime.Add(st.Cfg.TokenTTL), claims["exp"].(float64), deltaSeconds)
	// go run ./cmd/migrator/main.go --storage-path=./storage/grpc.db --migrations-path=./tests/migrations --migrations-table=migrations_test
}

func randomFakePassword() string {
	return gofakeit.Password(true, true, true, true, false, passDefaultLen)
}
