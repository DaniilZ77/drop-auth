package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/db/generated"
	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/domain/model"
	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/lib/logger/slogdiscard"
	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/service/mocks"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type dependencies struct {
	userService          *UserService
	userProvider         *mocks.UserProvider
	userModifier         *mocks.UserModifier
	refreshTokenModifier *mocks.RefreshTokenModifier
	refreshTokenProvider *mocks.RefreshTokenProvider
}

func createService(t *testing.T) dependencies {
	t.Helper()

	userProvider := mocks.NewUserProvider(t)
	userModifier := mocks.NewUserModifier(t)
	refreshTokenModifier := mocks.NewRefreshTokenModifier(t)
	refreshTokenProvider := mocks.NewRefreshTokenProvider(t)
	authConfig := model.AuthConfig{
		Secret:          "secret",
		AccessTokenTTL:  20,
		RefreshTokenTTL: 4200,
	}

	return dependencies{
		userService:          New(userModifier, userProvider, refreshTokenProvider, refreshTokenModifier, authConfig, slogdiscard.NewDiscardLogger()),
		userProvider:         userProvider,
		userModifier:         userModifier,
		refreshTokenModifier: refreshTokenModifier,
		refreshTokenProvider: refreshTokenProvider,
	}
}

type tokens struct {
	id    string
	admin *string
	exp   time.Time
}

func decodeToken(t *testing.T, secret string, tokenString string) tokens {
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		return []byte(secret), nil
	})
	require.NoError(t, err)

	var res tokens
	for key, value := range claims {
		switch key {
		case "id":
			id, ok := value.(string)
			require.True(t, ok)
			res.id = id
		case "admin":
			admin, ok := value.(string)
			if ok {
				res.admin = &admin
			}
		case "exp":
			exp, ok := value.(float64)
			require.True(t, ok)
			res.exp = time.Unix(int64(exp), 0)
		}
	}

	return res
}

func TestLogin_SuccessUserExists(t *testing.T) {
	t.Parallel()

	s := createService(t)
	ctx := context.Background()

	user := generated.SaveUserParams{
		Username:  "qwerty",
		Pseudonym: "qwerty",
		FirstName: "Aleskandr",
		LastName:  "Igorev",
	}
	id := uuid.New()

	s.userProvider.On("GetUserAdminByUsername", mock.Anything, user.Username).
		Return(&generated.GetUserAdminByUsernameRow{
			ID: id,
			Scale: generated.NullAdminScale{
				AdminScale: generated.AdminScaleMinor,
				Valid:      true,
			},
		}, nil).Once()

	var rt string
	s.refreshTokenModifier.On("SetRefreshToken", mock.Anything, id.String(), mock.MatchedBy(func(refreshToken string) bool {
		rt = refreshToken
		return uuid.Validate(refreshToken) == nil
	}), mock.Anything).Return(nil).Once()

	exp := time.Now().Add(time.Minute * time.Duration(s.userService.authConfig.AccessTokenTTL))
	accessToken, refreshToken, err := s.userService.Login(ctx, user)
	require.NoError(t, err)
	require.NotNil(t, refreshToken)
	assert.Equal(t, rt, *refreshToken)

	decodedAccessToken := decodeToken(t, s.userService.authConfig.Secret, *accessToken)
	assert.Equal(t, id.String(), decodedAccessToken.id)
	require.NotNil(t, decodedAccessToken.admin)
	assert.Equal(t, "minor", *decodedAccessToken.admin)

	const delta = 10 // 10 seconds
	assert.InDelta(t, exp.Unix(), decodedAccessToken.exp.Unix(), delta)
}

func TestLogin_SuccessUserNotExists(t *testing.T) {
	t.Parallel()

	s := createService(t)
	ctx := context.Background()

	user := generated.SaveUserParams{
		Username:  "qwerty",
		Pseudonym: "qwerty",
		FirstName: "Aleskandr",
		LastName:  "Igorev",
	}
	id := uuid.New()

	s.userProvider.On("GetUserAdminByUsername", mock.Anything, user.Username).
		Return(nil, model.ErrUserNotFound).Once()

	s.userModifier.On("SaveUser", mock.Anything, user).
		Return(&id, nil).Once()

	s.refreshTokenModifier.On("SetRefreshToken", mock.Anything, id.String(), mock.MatchedBy(func(refreshToken string) bool {
		return uuid.Validate(refreshToken) == nil
	}), mock.Anything).Return(nil).Once()

	accessToken, _, err := s.userService.Login(ctx, user)
	require.NoError(t, err)

	decodedAccessToken := decodeToken(t, s.userService.authConfig.Secret, *accessToken)
	assert.Equal(t, id.String(), decodedAccessToken.id)
	assert.Nil(t, decodedAccessToken.admin)
}

func TestLogin_FailEmptyPseudonym(t *testing.T) {
	t.Parallel()

	s := createService(t)
	ctx := context.Background()

	user := generated.SaveUserParams{
		Username:  "qwerty",
		Pseudonym: "",
		FirstName: "Aleskandr",
		LastName:  "Igorev",
	}

	s.userProvider.On("GetUserAdminByUsername", mock.Anything, user.Username).
		Return(nil, model.ErrUserNotFound).Once()

	_, _, err := s.userService.Login(ctx, user)
	require.ErrorIs(t, err, model.ErrEmptyPseudonym)
}

func TestLogin_Fail(t *testing.T) {
	t.Parallel()

	s := createService(t)
	ctx := context.Background()

	getUserByUsernameErr := errors.New("failed to get user by username")
	saveUserErr := errors.New("failed to save user")
	setRefreshTokenErr := errors.New("failed to set refresh token")

	tests := []struct {
		name string
		err  error
		beh  func()
	}{
		{
			name: "save user error",
			err:  saveUserErr,
			beh: func() {
				s.userProvider.On("GetUserAdminByUsername", mock.Anything, mock.Anything).
					Return(nil, model.ErrUserNotFound).Once()

				s.userModifier.On("SaveUser", mock.Anything, mock.Anything).
					Return(nil, saveUserErr).Once()
			},
		},
		{
			name: "set refresh token error",
			err:  setRefreshTokenErr,
			beh: func() {
				s.userProvider.On("GetUserAdminByUsername", mock.Anything, mock.Anything).
					Return(&generated.GetUserAdminByUsernameRow{}, nil).Once()

				s.refreshTokenModifier.On("SetRefreshToken", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(setRefreshTokenErr).Once()
			},
		},
		{
			name: "get user by username error",
			err:  getUserByUsernameErr,
			beh: func() {
				s.userProvider.On("GetUserAdminByUsername", mock.Anything, mock.Anything).
					Return(nil, getUserByUsernameErr).Once()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.beh()

			_, _, err := s.userService.Login(ctx, generated.SaveUserParams{Pseudonym: "qwerty"})
			assert.ErrorIs(t, err, tt.err)
		})
	}
}

func TestRefreshToken_Success(t *testing.T) {
	t.Parallel()

	s := createService(t)
	ctx := context.Background()

	token := uuid.NewString()
	userID := uuid.New()
	userIDString := userID.String()

	s.refreshTokenProvider.On("GetRefreshToken", mock.Anything, token).
		Return(&userIDString, nil).Once()

	s.userProvider.On("GetUserAdminByID", mock.Anything, userID).Return(&generated.GetUserAdminByIDRow{
		ID: userID,
		Scale: generated.NullAdminScale{
			AdminScale: generated.AdminScaleMajor,
			Valid:      true,
		},
	}, nil).Once()

	var rt string
	s.refreshTokenModifier.On("ReplaceRefreshToken", mock.Anything, token, mock.MatchedBy(func(newRefreshToken string) bool {
		rt = newRefreshToken
		return uuid.Validate(newRefreshToken) == nil
	}), userIDString, mock.Anything).
		Return(nil).Once()

	exp := time.Now().Add(time.Minute * time.Duration(s.userService.authConfig.AccessTokenTTL))
	accessToken, refreshToken, err := s.userService.RefreshToken(ctx, token)
	require.NoError(t, err)
	require.NotNil(t, accessToken)
	require.NotNil(t, refreshToken)
	assert.Equal(t, rt, *refreshToken)

	decodedAccessToken := decodeToken(t, s.userService.authConfig.Secret, *accessToken)
	assert.Equal(t, userIDString, decodedAccessToken.id)
	require.NotNil(t, decodedAccessToken.admin)
	assert.Equal(t, "major", *decodedAccessToken.admin)

	const delta = 10 // 10 seconds
	assert.InDelta(t, exp.Unix(), decodedAccessToken.exp.Unix(), delta)
}

func TestRefreshToken_Fail(t *testing.T) {
	t.Parallel()

	s := createService(t)
	ctx := context.Background()

	getRefreshTokenErr := errors.New("failed to get refresh token")
	getUserByIDErr := errors.New("failed to get user by id")
	replaceRefreshTokenErr := errors.New("failed to replace refresh token")

	userID := uuid.NewString()

	tests := []struct {
		name string
		err  error
		beh  func()
	}{
		{
			name: "get refresh token error",
			err:  getRefreshTokenErr,
			beh: func() {
				s.refreshTokenProvider.On("GetRefreshToken", mock.Anything, mock.Anything).
					Return(nil, getRefreshTokenErr).Once()
			},
		},
		{
			name: "get user by id error",
			err:  getUserByIDErr,
			beh: func() {
				s.refreshTokenProvider.On("GetRefreshToken", mock.Anything, mock.Anything).
					Return(&userID, nil).Once()

				s.userProvider.On("GetUserAdminByID", mock.Anything, mock.Anything).
					Return(nil, getUserByIDErr).Once()
			},
		},
		{
			name: "replace refresh token error",
			err:  replaceRefreshTokenErr,
			beh: func() {
				s.refreshTokenProvider.On("GetRefreshToken", mock.Anything, mock.Anything).
					Return(&userID, nil).Once()

				s.userProvider.On("GetUserAdminByID", mock.Anything, mock.Anything).
					Return(&generated.GetUserAdminByIDRow{}, nil).Once()

				s.refreshTokenModifier.On("ReplaceRefreshToken", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(replaceRefreshTokenErr).Once()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.beh()

			_, _, err := s.userService.RefreshToken(ctx, "")
			assert.ErrorIs(t, err, tt.err)
		})
	}
}

func TestAddAdmin_Success(t *testing.T) {
	t.Parallel()

	s := createService(t)
	ctx := context.Background()

	username := "qwerty"
	userID := uuid.New()

	s.userProvider.On("GetUserAdminByUsername", mock.Anything, username).
		Return(&generated.GetUserAdminByUsernameRow{ID: userID}, nil).Once()

	s.userModifier.On("SaveAdmin", mock.Anything, generated.SaveAdminParams{
		UserID: userID,
		Scale:  generated.AdminScaleMinor,
	}).Return(nil).Once()

	createdAt := time.Now()
	admin, err := s.userService.AddAdmin(ctx, username, generated.AdminScaleMajor)
	require.NoError(t, err)
	assert.NotNil(t, admin)

	assert.Equal(t, userID, admin.ID)
	assert.Equal(t, generated.AdminScaleMinor, admin.Scale)
	assert.Equal(t, username, admin.Username)

	const delta = 10 // 10 seconds
	assert.InDelta(t, createdAt.Unix(), admin.CreatedAt.Unix(), delta)
}

func TestAddAdmin_FailAdminNotMajor(t *testing.T) {
	t.Parallel()

	s := createService(t)
	ctx := context.Background()

	_, err := s.userService.AddAdmin(ctx, "", generated.AdminScaleMinor)
	assert.ErrorIs(t, err, model.ErrAdminNotMajor)
}

func TestAddAdmin_FailAdminAlreadyExists(t *testing.T) {
	t.Parallel()

	s := createService(t)
	ctx := context.Background()

	username := "qwerty"

	s.userProvider.On("GetUserAdminByUsername", mock.Anything, username).
		Return(&generated.GetUserAdminByUsernameRow{Scale: generated.NullAdminScale{Valid: true}}, nil).Once()

	_, err := s.userService.AddAdmin(ctx, username, generated.AdminScaleMajor)
	assert.ErrorIs(t, err, model.ErrAdminAlreadyExists)
}

func TestAddAdmin_Fail(t *testing.T) {
	t.Parallel()

	s := createService(t)
	ctx := context.Background()

	getUserByUsernameErr := errors.New("failed to get user by username")
	saveAdminErr := errors.New("failed to save admin")

	tests := []struct {
		name string
		err  error
		beh  func()
	}{
		{
			name: "get user by username error",
			err:  getUserByUsernameErr,
			beh: func() {
				s.userProvider.On("GetUserAdminByUsername", mock.Anything, mock.Anything).
					Return(nil, getUserByUsernameErr).Once()
			},
		},
		{
			name: "save admin error",
			err:  saveAdminErr,
			beh: func() {
				s.userProvider.On("GetUserAdminByUsername", mock.Anything, mock.Anything).
					Return(&generated.GetUserAdminByUsernameRow{}, nil).Once()

				s.userModifier.On("SaveAdmin", mock.Anything, mock.Anything).
					Return(saveAdminErr).Once()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.beh()

			_, err := s.userService.AddAdmin(ctx, "", generated.AdminScaleMajor)
			assert.ErrorIs(t, err, tt.err)
		})
	}
}

func TestDeleteAdmin_Success(t *testing.T) {
	t.Parallel()

	s := createService(t)
	ctx := context.Background()

	id := uuid.New()

	s.userProvider.On("GetUserAdminByID", mock.Anything, id).
		Return(&generated.GetUserAdminByIDRow{
			Scale: generated.NullAdminScale{
				Valid:      true,
				AdminScale: generated.AdminScaleMinor,
			}}, nil).Once()

	s.userModifier.On("DeleteAdmin", mock.Anything, id).Return(nil).Once()

	err := s.userService.DeleteAdmin(ctx, id, generated.AdminScaleMajor)
	require.NoError(t, err)
}

func TestDeleteAdmin_FailAdminNotMajor(t *testing.T) {
	t.Parallel()

	s := createService(t)
	ctx := context.Background()

	err := s.userService.DeleteAdmin(ctx, uuid.New(), generated.AdminScaleMinor)
	assert.ErrorIs(t, err, model.ErrAdminNotMajor)
}

func TestDeleteAdmin_FailCannotDeleteMajorAdmin(t *testing.T) {
	t.Parallel()

	s := createService(t)
	ctx := context.Background()

	id := uuid.New()

	s.userProvider.On("GetUserAdminByID", mock.Anything, id).
		Return(&generated.GetUserAdminByIDRow{
			Scale: generated.NullAdminScale{
				Valid:      true,
				AdminScale: generated.AdminScaleMajor,
			}}, nil).Once()

	err := s.userService.DeleteAdmin(ctx, id, generated.AdminScaleMajor)
	assert.ErrorIs(t, err, model.ErrCannotDeleteMajorAdmin)
}

func TestDeleteAdmin_Fail(t *testing.T) {
	t.Parallel()

	s := createService(t)
	ctx := context.Background()

	getUserByIDErr := errors.New("failed to get user by id")
	deleteAdminErr := errors.New("failed to delete admin")

	tests := []struct {
		name string
		err  error
		beh  func()
	}{
		{
			name: "get user by id error",
			err:  getUserByIDErr,
			beh: func() {
				s.userProvider.On("GetUserAdminByID", mock.Anything, mock.Anything).
					Return(nil, getUserByIDErr).Once()
			},
		},
		{
			name: "delete admin error",
			err:  deleteAdminErr,
			beh: func() {
				s.userProvider.On("GetUserAdminByID", mock.Anything, mock.Anything).
					Return(&generated.GetUserAdminByIDRow{}, nil).Once()

				s.userModifier.On("DeleteAdmin", mock.Anything, mock.Anything).
					Return(deleteAdminErr).Once()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.beh()

			err := s.userService.DeleteAdmin(ctx, uuid.New(), generated.AdminScaleMajor)
			assert.ErrorIs(t, err, tt.err)
		})
	}
}
