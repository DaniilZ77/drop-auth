package auth

import (
	"context"
	"net"
	"strconv"
	"testing"

	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/core"
	mocks "github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/core/mocks"
	authv1 "github.com/MAXXXIMUS-tropical-milkshake/beatflow-protos/gen/go/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

const (
	email     = "alex@gmail.com"
	password  = "Qwerty123456"
	firstName = "Alex"
	lastName  = "Ovechkin"
	pseudonym = "Ovi"
	ip        = "127.0.0.1"
	port      = "50051"
)

func createTestContextWithPeer(ip, port string) context.Context {
	addr := &net.TCPAddr{
		IP: net.ParseIP(ip),
		Port: func() int {
			if port == "" {
				return 0
			}
			p, _ := strconv.Atoi(port)
			return p
		}(),
	}

	p := &peer.Peer{
		Addr: addr,
	}

	return peer.NewContext(context.Background(), p)
}

func TestSignup_Success(t *testing.T) {
	t.Parallel()

	// init server and client
	authService := mocks.NewMockAuthService(t)
	client := &server{authService: authService}

	// vars
	username := "alex123"
	firstName := "Alex"
	lastName := "Ovechkin"
	pseudonym := "Ovi"
	code := "123456"
	user := &authv1.SignupRequest{
		Username: username,
		Email: &authv1.SignupEmail{
			Email: email,
			Code:  code,
		},
		FirstName: firstName,
		LastName:  lastName,
		Pseudonym: pseudonym,
		Password:  password,
	}
	coreUser := core.User{
		Username:     username,
		Email:        &[]string{email}[0],
		FirstName:    firstName,
		LastName:     lastName,
		Pseudonym:    pseudonym,
		PasswordHash: password,
	}
	retUser := &core.User{
		ID:       1,
		Username: username,
		Email:    &[]string{email}[0],
	}

	// mock behaviour
	authService.EXPECT().Signup(mock.Anything, code, mock.Anything, coreUser, ip+":"+port).Return(retUser, nil)

	res, err := client.Signup(createTestContextWithPeer(ip, port), user)
	require.NoError(t, err)
	assert.Equal(t, int64(retUser.ID), res.UserId)
	assert.Equal(t, *retUser.Email, res.Email)
	assert.Equal(t, retUser.Username, res.Username)
}

func TestSignup_ValidationErrors(t *testing.T) {
	t.Parallel()

	// init server and client
	authService := mocks.NewMockAuthService(t)
	client := &server{authService: authService}

	// vars
	username := "alex123"
	code := "123456"
	wantErr := status.Error(codes.InvalidArgument, core.ErrValidationFailed.Error())

	tests := []struct {
		name string
		user *authv1.SignupRequest
	}{
		{
			name: "empty username",
			user: &authv1.SignupRequest{
				Username: "",
				Email: &authv1.SignupEmail{
					Email: email,
					Code:  code,
				},
				FirstName: firstName,
				LastName:  lastName,
				Pseudonym: pseudonym,
				Password:  password,
			},
		},
		{
			name: "less than 8 chars password",
			user: &authv1.SignupRequest{
				Username: username,
				Email: &authv1.SignupEmail{
					Email: email,
					Code:  code,
				},
				FirstName: firstName,
				LastName:  lastName,
				Pseudonym: pseudonym,
				Password:  "123456",
			},
		},
		{
			name: "invalid email",
			user: &authv1.SignupRequest{
				Username: username,
				Email: &authv1.SignupEmail{
					Email: "alex@",
					Code:  code,
				},
				FirstName: firstName,
				LastName:  lastName,
				Pseudonym: pseudonym,
				Password:  password,
			},
		},
		{
			name: "empty first_name",
			user: &authv1.SignupRequest{
				Username: username,
				Email: &authv1.SignupEmail{
					Email: email,
					Code:  code,
				},
				FirstName: "",
				LastName:  lastName,
				Pseudonym: pseudonym,
				Password:  password,
			},
		},
		{
			name: "empty last_name",
			user: &authv1.SignupRequest{
				Username: username,
				Email: &authv1.SignupEmail{
					Email: email,
					Code:  code,
				},
				FirstName: firstName,
				LastName:  "",
				Pseudonym: pseudonym,
				Password:  password,
			},
		},
		{
			name: "empty pseudonym",
			user: &authv1.SignupRequest{
				Username: username,
				Email: &authv1.SignupEmail{
					Email: email,
					Code:  code,
				},
				FirstName: firstName,
				LastName:  lastName,
				Pseudonym: "",
				Password:  password,
			},
		},
		{
			name: "invalid telephone",
			user: &authv1.SignupRequest{
				Username: username,
				Telephone: &authv1.SignupTelephone{
					Telephone: "-79321900016",
					Code:      code,
				},
				FirstName: firstName,
				LastName:  lastName,
				Pseudonym: pseudonym,
				Password:  password,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.Signup(createTestContextWithPeer(ip, port), tt.user)
			require.Equal(t, err.Error(), wantErr.Error())
		})
	}
}

func TestSignup_SignupError(t *testing.T) {
	t.Parallel()

	// init server and client
	authService := mocks.NewMockAuthService(t)
	client := &server{authService: authService}

	// vars
	code := "123456"
	user := &authv1.SignupRequest{
		Username: "alex123",
		Email: &authv1.SignupEmail{
			Email: email,
			Code:  code,
		},
		FirstName: firstName,
		LastName:  lastName,
		Pseudonym: pseudonym,
		Password:  password,
	}

	tests := []struct {
		name      string
		behaviour func()
	}{
		{
			name: "email already exists",
			behaviour: func() {
				authService.EXPECT().Signup(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, core.ErrEmailAlreadyExists).Once()
			},
		},
		{
			name: "username already exists",
			behaviour: func() {
				authService.EXPECT().Signup(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, core.ErrUsernameAlreadyExists).Once()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.behaviour()

			_, err := client.Signup(createTestContextWithPeer(ip, port), user)
			st, ok := status.FromError(err)
			require.True(t, ok)
			require.Equal(t, st.Code(), codes.AlreadyExists)
		})
	}
}

func TestLogin_Success(t *testing.T) {
	t.Parallel()

	// init server and client
	authService := mocks.NewMockAuthService(t)
	client := &server{authService: authService}

	// vars
	user := &authv1.LoginRequest{
		Email:    email,
		Password: password,
	}
	coreUser := core.User{
		Email:        &[]string{email}[0],
		PasswordHash: password,
	}
	accessToken, refreshToken := "access_token", "refresh_token"

	// mock behaviour
	authService.EXPECT().Login(mock.Anything, coreUser).Return(&accessToken, &refreshToken, nil)

	res, err := client.Login(context.Background(), user)
	require.NoError(t, err)
	assert.Equal(t, accessToken, res.AccessToken)
	assert.Equal(t, refreshToken, res.RefreshToken)
}

func TestLogin_LoginError(t *testing.T) {
	t.Parallel()

	// init server and client
	authService := mocks.NewMockAuthService(t)
	client := &server{authService: authService}

	// vars
	ctx := context.Background()
	user := &authv1.LoginRequest{
		Email:    email,
		Password: password,
	}

	tests := []struct {
		name      string
		behaviour func()
	}{
		{
			name: "invalid credentials",
			behaviour: func() {
				authService.EXPECT().Login(mock.Anything, mock.Anything).Return(nil, nil, core.ErrInvalidCredentials).Once()
			},
		},
		{
			name: "user not found",
			behaviour: func() {
				authService.EXPECT().Login(mock.Anything, mock.Anything).Return(nil, nil, core.ErrUserNotFound).Once()
			},
		},
		{
			name: "user already deleted",
			behaviour: func() {
				authService.EXPECT().Login(mock.Anything, mock.Anything).Return(nil, nil, core.ErrAlreadyDeleted).Once()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.behaviour()

			_, err := client.Login(ctx, user)
			st, ok := status.FromError(err)
			require.True(t, ok)
			require.Equal(t, st.Code(), codes.Unauthenticated)
		})
	}
}
