// Package handlers ...
package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/SerjRamone/gophermart/internal/models"
	"github.com/SerjRamone/gophermart/internal/server/handlers/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func Test_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	mockHasher := mocks.NewMockHasher(ctrl)

	secret := []byte("supersecret")

	validUserForm := models.UserForm{
		Login:    "user",
		Password: "valid",
	}

	invalidUserForm := models.UserForm{
		Login:    "user",
		Password: "invalid",
	}

	validUser := models.User{
		Login:        "user",
		PasswordHash: "valid",
	}

	mockHasher.EXPECT().CompareHashAndPass(validUser.PasswordHash, validUserForm.Password).Return(true)

	storageRecorder := mockStorage.EXPECT()
	storageRecorder.GetUser(gomock.Any(), validUserForm).Return(&validUser, nil)
	storageRecorder.GetUser(gomock.Any(), invalidUserForm).Return(nil, models.ErrUserNotExists)

	bHandler := NewBaseHandler(
		secret,
		3600,
		mockStorage,
		mockHasher,
	)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bHandler.Login(r.Context(), w, r)
	})
	srv := httptest.NewServer(handler)
	defer srv.Close()

	tests := []struct {
		name     string
		url      string
		userForm any
		method   string
		status   int
	}{
		{
			name:     "Test#1. Valid user",
			url:      "/api/user/login",
			userForm: models.UserForm{Login: "user", Password: "valid"},
			method:   http.MethodPost,
			status:   http.StatusOK,
		},
		{
			name:     "Test#2. Invalid user",
			url:      "/api/user/login",
			userForm: models.UserForm{Login: "user", Password: "invalid"},
			method:   http.MethodPost,
			status:   http.StatusUnauthorized,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := json.Marshal(tt.userForm)
			if err != nil {
				t.Error(err)
			}

			resp, _ := testRequest(
				t,
				srv,
				tt.method,
				tt.url,
				"",
				bytes.NewBuffer(b))

			if err := resp.Body.Close(); err != nil {
				t.Error(err)
			}

			require.Equal(t, tt.status, resp.StatusCode,
				fmt.Sprintf("Test: %s URL: %s, want: %d, have: %d", tt.name, tt.url, tt.status, resp.StatusCode))
		})
	}
}

func Test_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	mockHasher := mocks.NewMockHasher(ctrl)

	secret := []byte("supersecret")

	bHandler := NewBaseHandler(
		secret,
		3600,
		mockStorage,
		mockHasher,
	)

	userForm1 := models.UserForm{
		Login:    "user",
		Password: "valid",
	}

	userForm2 := models.UserForm{
		Login:    "user2",
		Password: "valid",
	}

	validUser := models.User{
		Login:        "user",
		PasswordHash: "valid",
	}

	mockHasher.EXPECT().GetHash(userForm1.Password).Return(validUser.PasswordHash, nil)
	mockHasher.EXPECT().GetHash(userForm2.Password).Return(userForm2.Password, nil)

	storageRecorder := mockStorage.EXPECT()
	createUser := storageRecorder.CreateUser(gomock.Any(), userForm1)
	createUser.Return(&validUser, nil)

	storageRecorder.CreateUser(gomock.Any(), userForm2).After(createUser).Return(nil, models.ErrUserAlreadyExists)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bHandler.Register(r.Context(), w, r)
	})
	srv := httptest.NewServer(handler)
	defer srv.Close()

	tests := []struct {
		name     string
		url      string
		userForm any
		method   string
		status   int
	}{
		{
			name:     "Test#1. Valid user",
			url:      "/api/user/register",
			userForm: models.UserForm{Login: "user", Password: "valid"},
			method:   http.MethodPost,
			status:   http.StatusOK,
		},
		{
			name:     "Test#2. Already exists user",
			url:      "/api/user/register",
			userForm: models.UserForm{Login: "user2", Password: "valid"},
			method:   http.MethodPost,
			status:   http.StatusConflict,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := json.Marshal(tt.userForm)
			if err != nil {
				t.Error(err)
			}

			resp, _ := testRequest(
				t,
				srv,
				tt.method,
				tt.url,
				"",
				bytes.NewBuffer(b))

			if err := resp.Body.Close(); err != nil {
				t.Error(err)
			}

			require.Equal(t, tt.status, resp.StatusCode,
				fmt.Sprintf("Test: %s URL: %s, want: %d, have: %d", tt.name, tt.url, tt.status, resp.StatusCode))
		})
	}
}

func Test_AddOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	mockHasher := mocks.NewMockHasher(ctrl)

	secret := []byte("supersecret")

	bHandler := NewBaseHandler(
		secret,
		3600,
		mockStorage,
		mockHasher,
	)

	orderForm1 := models.OrderForm{
		UserID: "1",
		Number: "7305748056314637",
	}

	orderForm2 := models.OrderForm{
		UserID: "2",
		Number: "1090888814505555",
	}

	orderForm3 := models.OrderForm{
		UserID: "2",
		Number: "7305748056314637",
	}

	order1 := models.Order{
		ID:         "1",
		UserID:     "1",
		Status:     models.OrderStatusNew,
		Number:     "7305748056314637",
		Accrual:    0,
		UploadedAt: time.Now(),
	}

	userForm1 := models.UserForm{
		Login:    "user1",
		Password: "pass1",
	}

	userFormClaims1 := models.UserForm{
		Login: "user1",
	}

	userForm2 := models.UserForm{
		Login:    "user2",
		Password: "pass2",
	}

	userFormClaims2 := models.UserForm{
		Login: "user2",
	}

	user1 := models.User{
		ID:           "1",
		Login:        "user1",
		PasswordHash: "pass1",
	}

	user2 := models.User{
		ID:           "2",
		Login:        "user2",
		PasswordHash: "pass2",
	}

	mockHasher.EXPECT().CompareHashAndPass(gomock.Any(), gomock.Any()).AnyTimes().Return(true)

	storageRecorder := mockStorage.EXPECT()

	storageRecorder.GetUser(gomock.Any(), userForm1).AnyTimes().Return(&user1, nil)
	storageRecorder.GetUser(gomock.Any(), userForm2).AnyTimes().Return(&user2, nil)
	storageRecorder.GetUser(gomock.Any(), userFormClaims1).AnyTimes().Return(&user1, nil)
	storageRecorder.GetUser(gomock.Any(), userFormClaims2).AnyTimes().Return(&user2, nil)

	storageRecorder.CreateOrder(gomock.Any(), orderForm1).AnyTimes().Return(&order1, nil)
	storageRecorder.CreateOrder(gomock.Any(), orderForm2).AnyTimes().Return(nil, models.ErrOrderAlreadyExists)
	storageRecorder.CreateOrder(gomock.Any(), orderForm3).AnyTimes().Return(nil, errors.New("not found"))

	login := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bHandler.Login(r.Context(), w, r)
	})
	orders := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			bHandler.PostOrder(r.Context(), w, r)
		})
		bHandler.JWTMiddleware(handler).ServeHTTP(w, r)
	})

	mux := http.NewServeMux()
	mux.HandleFunc("/api/user/login", login)
	mux.HandleFunc("/api/user/orders", withMiddleware(orders, bHandler.JWTMiddleware))

	srv := httptest.NewServer(mux)
	defer srv.Close()

	var tests = []struct {
		body   any
		name   string
		url    string
		method string
		auth   *models.UserForm
		status int
	}{
		{
			name:   "Test#1. Success adding",
			url:    "/api/user/orders",
			status: http.StatusAccepted,
			method: http.MethodPost,
			auth:   &models.UserForm{Login: "user1", Password: "pass1"},
			body:   7305748056314637,
		},
		{
			name:   "Test#2. Unauthorized",
			url:    "/api/user/orders",
			status: http.StatusUnauthorized,
			method: http.MethodPost,
			auth:   nil,
			body:   7305748056314637,
		},
		{
			name:   "Test#3. Already exists order",
			url:    "/api/user/orders",
			status: http.StatusAccepted,
			method: http.MethodPost,
			auth:   &models.UserForm{Login: "user1", Password: "pass1"},
			body:   7305748056314637,
		},
		{
			name:   "Test#4. Not unique order (other user)",
			url:    "/api/user/orders",
			status: http.StatusBadRequest,
			method: http.MethodPost,
			auth:   &models.UserForm{Login: "user2", Password: "pass2"},
			body:   7305748056314637,
		},
	}

	for _, tt := range tests {
		b, err := json.Marshal(tt.body)
		if err != nil {
			t.Error(err)
		}

		resp, _ := testRequest(t,
			srv,
			tt.method,
			tt.url,
			getAuthToken(t, srv, tt.auth),
			bytes.NewBuffer(b))

		if err := resp.Body.Close(); err != nil {
			t.Error(err)
		}

		require.Equal(t, tt.status, resp.StatusCode, fmt.Sprintf("Test: %s URL: %s, want: %d, have: %d", tt.name, tt.url, tt.status, resp.StatusCode))
	}
}

func Test_GetOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	mockHasher := mocks.NewMockHasher(ctrl)

	secret := []byte("supersecret")

	bHandler := NewBaseHandler(
		secret,
		3600,
		mockStorage,
		mockHasher,
	)

	userForm1 := models.UserForm{
		Login:    "user1",
		Password: "pass1",
	}

	userFormClaims1 := models.UserForm{
		Login: "user1",
	}

	user1 := models.User{
		ID:           "1",
		Login:        "user1",
		PasswordHash: "pass1",
	}

	orders := []*models.Order{
		{
			ID:         "1",
			UserID:     "1",
			Number:     "7305748056314637",
			Status:     models.OrderStatusNew,
			Accrual:    0,
			UploadedAt: time.Now(),
		},
	}

	mockHasher.EXPECT().CompareHashAndPass(gomock.Any(), gomock.Any()).AnyTimes().Return(true)

	storageRecorder := mockStorage.EXPECT()

	storageRecorder.GetUser(gomock.Any(), userForm1).AnyTimes().Return(&user1, nil)
	storageRecorder.GetUser(gomock.Any(), userFormClaims1).AnyTimes().Return(&user1, nil)

	storageRecorder.GetUserOrders(gomock.Any(), &user1).AnyTimes().Return(orders, nil)

	login := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bHandler.Login(r.Context(), w, r)
	})
	ordersHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			bHandler.GetOrder(r.Context(), w, r)
		})
		bHandler.JWTMiddleware(handler).ServeHTTP(w, r)
	})

	mux := http.NewServeMux()
	mux.HandleFunc("/api/user/login", login)
	mux.HandleFunc("/api/user/orders", withMiddleware(ordersHandler, bHandler.JWTMiddleware))

	srv := httptest.NewServer(mux)
	defer srv.Close()

	var tests = []struct {
		name   string
		url    string
		method string
		auth   *models.UserForm
		status int
	}{
		{
			name:   "Test#1. Unauthorized",
			url:    "/api/user/orders",
			status: http.StatusUnauthorized,
			method: http.MethodGet,
			auth:   nil,
		},
		{
			name:   "Test#2. Valid",
			url:    "/api/user/orders",
			status: http.StatusOK,
			method: http.MethodGet,
			auth:   &models.UserForm{Login: "user1", Password: "pass1"},
		},
	}

	for _, tt := range tests {
		var b []byte
		resp, _ := testRequest(t,
			srv,
			tt.method,
			tt.url,
			getAuthToken(t, srv, tt.auth),
			bytes.NewBuffer(b))

		if err := resp.Body.Close(); err != nil {
			t.Error(err)
		}

		require.Equal(t, tt.status, resp.StatusCode, fmt.Sprintf("Test: %s URL: %s, want: %d, have: %d", tt.name, tt.url, tt.status, resp.StatusCode))
	}
}

func Test_Balance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	mockHasher := mocks.NewMockHasher(ctrl)

	secret := []byte("supersecret")

	bHandler := NewBaseHandler(
		secret,
		3600,
		mockStorage,
		mockHasher,
	)

	userForm1 := models.UserForm{
		Login:    "user1",
		Password: "pass1",
	}

	userFormClaims1 := models.UserForm{
		Login: "user1",
	}

	user1 := models.User{
		ID:           "1",
		Login:        "user1",
		PasswordHash: "pass1",
	}

	userBalance1 := models.UserBalance{
		Current:   0.100,
		Withdrawn: 0.500,
	}

	mockHasher.EXPECT().CompareHashAndPass(gomock.Any(), gomock.Any()).AnyTimes().Return(true)

	storageRecorder := mockStorage.EXPECT()

	storageRecorder.GetUser(gomock.Any(), userForm1).AnyTimes().Return(&user1, nil)
	storageRecorder.GetUser(gomock.Any(), userFormClaims1).AnyTimes().Return(&user1, nil)

	storageRecorder.GetUserBalance(gomock.Any(), user1.ID).AnyTimes().Return(&userBalance1, nil)

	login := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bHandler.Login(r.Context(), w, r)
	})
	ordersHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			bHandler.Balance(r.Context(), w, r)
		})
		bHandler.JWTMiddleware(handler).ServeHTTP(w, r)
	})

	mux := http.NewServeMux()
	mux.HandleFunc("/api/user/login", login)
	mux.HandleFunc("/api/user/balance", withMiddleware(ordersHandler, bHandler.JWTMiddleware))

	srv := httptest.NewServer(mux)
	defer srv.Close()

	var tests = []struct {
		name          string
		url           string
		method        string
		auth          *models.UserForm
		status        int
		wantCurrent   float64
		wantWithdrawn float64
	}{
		{
			name:   "Test#1. Unauthorized",
			url:    "/api/user/balance",
			status: http.StatusUnauthorized,
			method: http.MethodGet,
			auth:   nil,
		},
		{
			name:          "Test#2. Valid",
			url:           "/api/user/balance",
			status:        http.StatusOK,
			method:        http.MethodGet,
			auth:          &models.UserForm{Login: "user1", Password: "pass1"},
			wantCurrent:   0.100,
			wantWithdrawn: 0.500,
		},
	}

	for _, tt := range tests {
		var b []byte
		resp, rBytes := testRequest(t,
			srv,
			tt.method,
			tt.url,
			getAuthToken(t, srv, tt.auth),
			bytes.NewBuffer(b))

		if err := resp.Body.Close(); err != nil {
			t.Error(err)
		}

		require.Equal(t, tt.status, resp.StatusCode, fmt.Sprintf("Test: %s URL: %s, want: %d, have: %d", tt.name, tt.url, tt.status, resp.StatusCode))

		if resp.StatusCode == http.StatusOK {
			var balance models.UserBalance

			if err := json.Unmarshal(rBytes, &balance); err != nil {
				t.Error(err)
			}

			require.Equal(t, tt.wantCurrent, balance.Current,
				fmt.Sprintf("Current balance: %s URL: %s, want: %g, have: %g",
					tt.name, tt.url, tt.wantCurrent, balance.Current))

			require.Equal(t, tt.wantWithdrawn, balance.Withdrawn,
				fmt.Sprintf("Withdrawn balance: %s URL: %s, want: %g, have: %g",
					tt.name, tt.url, tt.wantWithdrawn, balance.Withdrawn))
		}
	}
}

func Test_BalanceWithdrawn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	mockHasher := mocks.NewMockHasher(ctrl)

	secret := []byte("supersecret")

	bHandler := NewBaseHandler(
		secret,
		3600,
		mockStorage,
		mockHasher,
	)

	userForm1 := models.UserForm{
		Login:    "user1",
		Password: "pass1",
	}

	userFormClaims1 := models.UserForm{
		Login: "user1",
	}

	user1 := models.User{
		ID:           "1",
		Login:        "user1",
		PasswordHash: "pass1",
	}

	mockHasher.EXPECT().CompareHashAndPass(gomock.Any(), gomock.Any()).AnyTimes().Return(true)

	storageRecorder := mockStorage.EXPECT()

	storageRecorder.GetUser(gomock.Any(), userForm1).AnyTimes().Return(&user1, nil)
	storageRecorder.GetUser(gomock.Any(), userFormClaims1).AnyTimes().Return(&user1, nil)

	storageRecorder.CreateWithdrawal(gomock.Any(), user1.ID, "7305748056314637", 100.100).AnyTimes().Return(nil)
	storageRecorder.CreateWithdrawal(gomock.Any(), user1.ID, "1090888814505555", 200.200).AnyTimes().Return(models.ErrNotEnoughPoints)

	login := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bHandler.Login(r.Context(), w, r)
	})
	withdraw := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			bHandler.Withdraw(r.Context(), w, r)
		})
		bHandler.JWTMiddleware(handler).ServeHTTP(w, r)
	})

	mux := http.NewServeMux()
	mux.HandleFunc("/api/user/login", login)
	mux.HandleFunc("/api/user/balance/withdraw", withMiddleware(withdraw, bHandler.JWTMiddleware))

	srv := httptest.NewServer(mux)
	defer srv.Close()

	var tests = []struct {
		name   string
		url    string
		method string
		auth   *models.UserForm
		status int
		order  string
		sum    float64
	}{
		{
			name:   "Test#1. Unauthorized",
			url:    "/api/user/balance/withdraw",
			status: http.StatusUnauthorized,
			method: http.MethodPost,
			auth:   nil,
		},
		{
			name:   "Test#2. Valid withdraw",
			url:    "/api/user/balance/withdraw",
			status: http.StatusOK,
			method: http.MethodPost,
			auth:   &models.UserForm{Login: "user1", Password: "pass1"},
			order:  "7305748056314637",
			sum:    100.100,
		},
		{
			name:   "Test#3. Not enough points",
			url:    "/api/user/balance/withdraw",
			status: http.StatusPaymentRequired,
			method: http.MethodPost,
			auth:   &models.UserForm{Login: "user1", Password: "pass1"},
			order:  "1090888814505555",
			sum:    200.200,
		},
	}

	for _, tt := range tests {
		withdraw := struct {
			Order string  `json:"order"`
			Sum   float64 `json:"sum"`
		}{
			Order: tt.order,
			Sum:   tt.sum,
		}
		b, err := json.Marshal(withdraw)
		if err != nil {
			t.Error(err)
		}
		resp, _ := testRequest(t,
			srv,
			tt.method,
			tt.url,
			getAuthToken(t, srv, tt.auth),
			bytes.NewBuffer(b))

		if err := resp.Body.Close(); err != nil {
			t.Error(err)
		}

		require.Equal(t, tt.status, resp.StatusCode, fmt.Sprintf("Test: %s URL: %s, want: %d, have: %d", tt.name, tt.url, tt.status, resp.StatusCode))
	}
}

func Test_Withdrawals(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	mockHasher := mocks.NewMockHasher(ctrl)

	secret := []byte("supersecret")

	bHandler := NewBaseHandler(
		secret,
		3600,
		mockStorage,
		mockHasher,
	)

	userForm1 := models.UserForm{
		Login:    "user1",
		Password: "pass1",
	}

	userFormClaims1 := models.UserForm{
		Login: "user1",
	}

	user1 := models.User{
		ID:           "1",
		Login:        "user1",
		PasswordHash: "pass1",
	}

	withdrawals := []*models.Withdrawal{
		{
			OrderNumber: "8885901057661813",
			Total:       0.100,
			CreatedAt:   time.Now(),
		},
		{
			OrderNumber: "1154576128108785",
			Total:       100.100,
			CreatedAt:   time.Now(),
		},
		{
			OrderNumber: "7956829830887973",
			Total:       111.222,
			CreatedAt:   time.Now(),
		},
	}

	mockHasher.EXPECT().CompareHashAndPass(gomock.Any(), gomock.Any()).AnyTimes().Return(true)

	storageRecorder := mockStorage.EXPECT()

	storageRecorder.GetUser(gomock.Any(), userForm1).AnyTimes().Return(&user1, nil)
	storageRecorder.GetUser(gomock.Any(), userFormClaims1).AnyTimes().Return(&user1, nil)

	storageRecorder.GetWithdrawals(gomock.Any(), user1.ID).AnyTimes().Return(withdrawals, nil)

	login := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bHandler.Login(r.Context(), w, r)
	})
	ordersHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			bHandler.Withdrawals(r.Context(), w, r)
		})
		bHandler.JWTMiddleware(handler).ServeHTTP(w, r)
	})

	mux := http.NewServeMux()
	mux.HandleFunc("/api/user/login", login)
	mux.HandleFunc("/api/user/withdrawals", withMiddleware(ordersHandler, bHandler.JWTMiddleware))

	srv := httptest.NewServer(mux)
	defer srv.Close()

	var tests = []struct {
		name   string
		url    string
		method string
		auth   *models.UserForm
		status int
	}{
		{
			name:   "Test#1. Unauthorized",
			url:    "/api/user/withdrawals",
			status: http.StatusUnauthorized,
			method: http.MethodGet,
			auth:   nil,
		},
		{
			name:   "Test#2. Valid",
			url:    "/api/user/withdrawals",
			status: http.StatusOK,
			method: http.MethodGet,
			auth:   &models.UserForm{Login: "user1", Password: "pass1"},
		},
	}

	for _, tt := range tests {
		var b []byte
		resp, _ := testRequest(t,
			srv,
			tt.method,
			tt.url,
			getAuthToken(t, srv, tt.auth),
			bytes.NewBuffer(b))

		if err := resp.Body.Close(); err != nil {
			t.Error(err)
		}

		require.Equal(t, tt.status, resp.StatusCode, fmt.Sprintf("Test: %s URL: %s, want: %d, have: %d", tt.name, tt.url, tt.status, resp.StatusCode))
	}
}

func getAuthToken(t *testing.T, ts *httptest.Server, uf *models.UserForm) string {
	t.Helper()

	if uf == nil {
		return ""
	}

	b, err := json.Marshal(uf)
	if err != nil {
		t.Error(err)
	}

	resp, _ := testRequest(t, ts, http.MethodPost, "/api/user/login", "", bytes.NewBuffer(b))
	if err := resp.Body.Close(); err != nil {
		t.Error(err)
	}

	return resp.Header.Get("Authorization")
}

func testRequest(t *testing.T, ts *httptest.Server,
	method string,
	path string,
	jwt string,
	body io.Reader,
) (*http.Response, []byte) {
	t.Helper()

	r, err := url.JoinPath(ts.URL, path)
	if err != nil {
		t.Errorf("URL %s test request  error : %v", err, path)
	}

	req, err := http.NewRequest(method, r, body)
	if jwt != "" {
		req.Header.Set("Authorization", jwt)
	}
	require.NoError(t, err)

	// fmt.Println(path)
	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Error(err)
		}
	}()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, respBody
}

func withMiddleware(handler http.HandlerFunc, middleware func(http.Handler) http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		middleware(handler).ServeHTTP(w, r)
	}
}
