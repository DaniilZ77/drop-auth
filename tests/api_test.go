package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/tests/testhelpers"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ApiTestSuite struct {
	suite.Suite
	pgContainer      *testhelpers.PostgresContainer
	redisContainer   *testhelpers.RedisContainer
	backendContainer *testhelpers.BackendContainer
	network          *testhelpers.Network
	ctx              context.Context
}

func (suite *ApiTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	network, err := testhelpers.CreateNetwork(suite.ctx)
	if err != nil {
		log.Fatal(err)
	}
	suite.network = network

	pgContainer, databaseHost, err := testhelpers.CreatePostgresContainer(suite.ctx, network.DockerNetwork)
	if err != nil {
		log.Fatal(err)
	}
	suite.pgContainer = pgContainer

	redisContainer, redisHost, err := testhelpers.CreateRedisContainer(suite.ctx, network.DockerNetwork)
	if err != nil {
		log.Fatal(err)
	}
	suite.redisContainer = redisContainer

	backendContainer, err := testhelpers.CreateBackendContainer(suite.ctx,
		network.DockerNetwork.Name,
		*databaseHost,
		*redisHost,
	)
	if err != nil {
		log.Fatal(err)
	}
	suite.backendContainer = backendContainer
}

func (suite *ApiTestSuite) TearDownSuite() {
	if err := suite.pgContainer.DB.Close(suite.ctx); err != nil {
		log.Fatalf("error closing postgres connection: %s", err)
	}
	if err := suite.pgContainer.Terminate(suite.ctx); err != nil {
		log.Fatalf("error terminating postgres container: %s", err)
	}
	if err := suite.redisContainer.Terminate(suite.ctx); err != nil {
		log.Fatalf("error terminating redis container: %s", err)
	}
	if err := suite.backendContainer.Terminate(suite.ctx); err != nil {
		log.Fatalf("error terminating backend container: %s", err)
	}
	if err := suite.network.Remove(suite.ctx); err != nil {
		log.Fatalf("error removing network: %s", err)
	}
}

func (suite *ApiTestSuite) SetupTest() {
	query, err := os.ReadFile("./testdata/init.sql")
	if err != nil {
		log.Fatal(err)
	}
	if _, err := suite.pgContainer.DB.Exec(suite.ctx, string(query)); err != nil {
		log.Fatal(err)
	}
}

func (suite *ApiTestSuite) TearDownTest() {
	if _, err := suite.pgContainer.DB.Exec(suite.ctx, "truncate table users_admins cascade"); err != nil {
		log.Fatal(err)
	}
	if _, err := suite.pgContainer.DB.Exec(suite.ctx, "truncate table users cascade"); err != nil {
		log.Fatal(err)
	}
}

func (suite *ApiTestSuite) TestLogin_Success() {
	t := suite.T()

	if testing.Short() {
		t.Skip()
	}

	resp, err := suite.backendContainer.PostRequest(
		"/v1/auth/login",
		`{"pseudonym": "qwerty"}`,
		testhelpers.WithTmaToken(map[string]string{
			"username":   "aleks123",
			"first_name": "Alexander",
			"last_name":  "Ilin",
		}),
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	type user struct {
		Username  string
		FirstName string
		LastName  string
		Pseudonym string
	}

	row, err := suite.pgContainer.DB.Query(suite.ctx, `select username, first_name, last_name, pseudonym from users where username = 'aleks123' limit 1`)
	require.NoError(t, err)
	u, err := pgx.CollectOneRow(row, pgx.RowToStructByName[user])
	require.NoError(t, err)
	assert.Equal(t, user{
		Username:  "aleks123",
		FirstName: "Alexander",
		LastName:  "Ilin",
		Pseudonym: "qwerty",
	}, u)
}

func (suite *ApiTestSuite) TestLogin_FailTokenNotProvided() {
	t := suite.T()

	if testing.Short() {
		t.Skip()
	}

	resp, err := suite.backendContainer.PostRequest("/v1/auth/login", `{}`)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func (suite *ApiTestSuite) TestLogin_FailEmptyPseudonym() {
	t := suite.T()

	if testing.Short() {
		t.Skip()
	}

	resp, err := suite.backendContainer.PostRequest(
		"/v1/auth/login",
		`{}`,
		testhelpers.WithTmaToken(map[string]string{
			"username":   "aleks123",
			"first_name": "Alexander",
			"last_name":  "Ilin",
		}),
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func (suite *ApiTestSuite) TestLogin_SuccessTwice() {
	t := suite.T()

	if testing.Short() {
		t.Skip()
	}

	params := map[string]string{
		"username":   "aleks123",
		"first_name": "Alexander",
		"last_name":  "Ilin",
	}

	resp, err := suite.backendContainer.PostRequest("/v1/auth/login", `{"pseudonym": "qwerty"}`, testhelpers.WithTmaToken(params))
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	resp, err = suite.backendContainer.PostRequest("/v1/auth/login", `{"pseudonym": "qwerty"}`, testhelpers.WithTmaToken(params))
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func (suite *ApiTestSuite) TestRefreshToken_Success() {
	t := suite.T()

	if testing.Short() {
		t.Skip()
	}

	type tokens struct {
		RefreshToken string `json:"refreshToken"`
	}

	resp, err := suite.backendContainer.PostRequest("/v1/auth/login", `{"pseudonym": "qwerty"}`, testhelpers.WithTmaToken(map[string]string{
		"username":   "aleks123",
		"first_name": "Alexander",
		"last_name":  "Ilin",
	}))
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	refreshToken := &tokens{}
	err = json.NewDecoder(resp.Body).Decode(&refreshToken)
	require.NoError(t, err)

	resp, err = suite.backendContainer.PostRequest(
		"/v1/auth/token/refresh",
		fmt.Sprintf(`{"refreshToken":%q}`, refreshToken.RefreshToken))
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func (suite *ApiTestSuite) TestRefreshToken_FailSecondRequestWithSameToken() {
	t := suite.T()

	if testing.Short() {
		t.Skip()
	}

	type tokens struct {
		RefreshToken string `json:"refreshToken"`
	}

	resp, err := suite.backendContainer.PostRequest("/v1/auth/login", `{"pseudonym": "qwerty"}`, testhelpers.WithTmaToken(map[string]string{
		"username":   "aleks123",
		"first_name": "Alexander",
		"last_name":  "Ilin",
	}))
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	refreshToken := &tokens{}
	err = json.NewDecoder(resp.Body).Decode(&refreshToken)
	require.NoError(t, err)

	resp, err = suite.backendContainer.PostRequest(
		"/v1/auth/token/refresh",
		fmt.Sprintf(`{"refreshToken":%q}`, refreshToken.RefreshToken))
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	resp, err = suite.backendContainer.PostRequest(
		"/v1/auth/token/refresh",
		fmt.Sprintf(`{"refreshToken":%q}`, refreshToken.RefreshToken))
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func (suite *ApiTestSuite) TestGetUsers_Success() {
	t := suite.T()

	if testing.Short() {
		t.Skip()
	}

	type pagination struct {
		Pages          string `json:"pages"`
		CurPage        string `json:"curPage"`
		Records        string `json:"records"`
		RecordsPerPage string `json:"recordsPerPage"`
	}

	type user struct {
		UserID    string `json:"userId"`
		Username  string `json:"username"`
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
		Pseudonym string `json:"pseudonym"`
		CreatedAt string `json:"createdAt"`
	}

	type users struct {
		Pagination pagination `json:"pagination"`
		Users      []user     `json:"users"`
	}

	var userID string

	t.Run("filter by username test", func(t *testing.T) {
		vals := url.Values{}
		vals.Add("username", "svannozzii2")
		vals.Add("limit", "1")
		resp, err := suite.backendContainer.GetRequest("/v1/users", vals)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		users := &users{}
		err = json.NewDecoder(resp.Body).Decode(&users)
		require.NoError(t, err)

		require.Len(t, users.Users, 1)
		assert.Equal(t, "svannozzii2", users.Users[0].Username)
		userID = users.Users[0].UserID
	})

	t.Run("filter by id test", func(t *testing.T) {
		vals := url.Values{}
		vals.Add("userId", userID)
		vals.Add("limit", "1")
		resp, err := suite.backendContainer.GetRequest("/v1/users", vals)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		users := &users{}
		err = json.NewDecoder(resp.Body).Decode(&users)
		require.NoError(t, err)
		require.Len(t, users.Users, 1)
		assert.Equal(t, userID, users.Users[0].UserID)
	})

	t.Run("orderBy, limit and offset test", func(t *testing.T) {
		vals := url.Values{}
		vals.Add("limit", "50")
		vals.Add("offset", "25")
		vals.Add("orderBy.order", "asc")
		vals.Add("orderBy.field", "created_at")
		resp, err := suite.backendContainer.GetRequest("/v1/users", vals)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		users := &users{}
		err = json.NewDecoder(resp.Body).Decode(&users)
		require.NoError(t, err)

		require.Len(t, users.Users, 50)

		var prevTime int64
		for _, user := range users.Users {
			createdAt, err := time.Parse(time.RFC3339, user.CreatedAt)
			require.NoError(t, err)
			assert.GreaterOrEqual(t, createdAt.Unix(), prevTime)
			prevTime = createdAt.Unix()
		}
	})

	t.Run("filter by first_name and last_name test", func(t *testing.T) {
		vals := url.Values{}
		vals.Add("firstName", "Dede")
		vals.Add("lastName", "Depport")
		vals.Add("limit", "10")
		resp, err := suite.backendContainer.GetRequest("/v1/users", vals)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		users := &users{}
		err = json.NewDecoder(resp.Body).Decode(&users)
		require.NoError(t, err)

		assert.GreaterOrEqual(t, len(users.Users), 1)
		for _, user := range users.Users {
			assert.Equal(t, "Dede", user.FirstName)
			assert.Equal(t, "Depport", user.LastName)
		}
	})

	t.Run("filter by pseudonym test", func(t *testing.T) {
		vals := url.Values{}
		vals.Add("pseudonym", "aspeed1l")
		vals.Add("limit", "10")
		resp, err := suite.backendContainer.GetRequest("/v1/users", vals)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		users := &users{}
		err = json.NewDecoder(resp.Body).Decode(&users)
		require.NoError(t, err)

		assert.GreaterOrEqual(t, len(users.Users), 1)
		for _, user := range users.Users {
			assert.Equal(t, "aspeed1l", user.Pseudonym)
		}
	})
}

func (suite *ApiTestSuite) TestGetUsers_FailValidation() {
	t := suite.T()

	if testing.Short() {
		t.Skip()
	}

	tests := []struct {
		name    string
		addVals func(url.Values)
	}{
		{
			name: "invalid orderBy test",
			addVals: func(vals url.Values) {
				vals.Add("limit", "1")
				vals.Add("orderBy.order", "asc")
			},
		},
		{
			name: "too long firstName test",
			addVals: func(vals url.Values) {
				vals.Add("limit", "1")
				vals.Add("firstName", strings.Repeat("qwerty", 200))
			},
		},
		{
			name: "empty firstName test",
			addVals: func(vals url.Values) {
				vals.Add("limit", "1")
				vals.Add("firstName", "")
			},
		},
		{
			name: "invalid orderBy field",
			addVals: func(vals url.Values) {
				vals.Add("limit", "1")
				vals.Add("orderBy.order", "asc")
				vals.Add("orderBy.field", "invalidField")
			},
		},
		{
			name: "too long lastName test",
			addVals: func(vals url.Values) {
				vals.Add("limit", "1")
				vals.Add("lastName", strings.Repeat("qwerty", 200))
			},
		},
		{
			name: "empty lastName test",
			addVals: func(vals url.Values) {
				vals.Add("limit", "1")
				vals.Add("lastName", "")
			},
		},
		{
			name: "too long pseudonym test",
			addVals: func(vals url.Values) {
				vals.Add("limit", "1")
				vals.Add("pseudonym", strings.Repeat("qwerty", 200))
			},
		},
		{
			name: "empty pseudonym test",
			addVals: func(vals url.Values) {
				vals.Add("pseudonym", "")
			},
		},
		{
			name: "too big limit",
			addVals: func(vals url.Values) {
				vals.Add("limit", "1000000")
			},
		},
		{
			name: "empty limit",
			addVals: func(vals url.Values) {
				vals.Add("limit", "")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vals := url.Values{}
			tt.addVals(vals)
			resp, err := suite.backendContainer.GetRequest("/v1/users", vals)
			require.NoError(t, err)
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		})
	}
}

func (suite *ApiTestSuite) TestUpdateUser_Success() {
	t := suite.T()

	if testing.Short() {
		t.Skip()
	}

	type tokens struct {
		AccessToken string `json:"accessToken"`
	}

	resp, err := suite.backendContainer.PostRequest("/v1/auth/login", `{"pseudonym": "qwerty"}`, testhelpers.WithTmaToken(map[string]string{
		"username":   "aleks123",
		"first_name": "Alexander",
		"last_name":  "Ilin",
	}))
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	accessToken := &tokens{}
	err = json.NewDecoder(resp.Body).Decode(&accessToken)
	require.NoError(t, err)

	resp, err = suite.backendContainer.PatchRequest("/v1/user",
		`{"lastName": "Ovechkin","firstName":"Dmitry","pseudonym":"qwerty123"}`,
		testhelpers.WithBearerToken(accessToken.AccessToken))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	type user struct {
		Username  string
		FirstName string
		LastName  string
		Pseudonym string
	}

	row, err := suite.pgContainer.DB.Query(suite.ctx, `select username, first_name, last_name, pseudonym from users where username = 'aleks123' limit 1`)
	require.NoError(t, err)
	u, err := pgx.CollectOneRow(row, pgx.RowToStructByName[user])
	require.NoError(t, err)
	assert.Equal(t, user{
		Username:  "aleks123",
		FirstName: "Dmitry",
		LastName:  "Ovechkin",
		Pseudonym: "qwerty123",
	}, u)
}

func TestApiTestSuite(t *testing.T) {
	suite.Run(t, new(ApiTestSuite))
}
