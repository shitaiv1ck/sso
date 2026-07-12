package authgrpc

import (
	"fmt"
	"testing"
	"time"

	ssov1 "github.com/shitaiv1ck/protos/gen/go/sso"
	"github.com/shitaiv1ck/sso/internal/core/domain"
	errs "github.com/shitaiv1ck/sso/internal/core/errors"
	"github.com/shitaiv1ck/sso/internal/core/logger"
	mock_authgrpc "github.com/shitaiv1ck/sso/internal/features/auth/transport/grpc/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestRegister(t *testing.T) {
	type mockBehaivor func(s *mock_authgrpc.MockAuthService, req *ssov1.RegisterRequest)

	testTable := []struct {
		name           string
		inputReq       *ssov1.RegisterRequest
		mockBehavior   mockBehaivor
		expectedStatus codes.Code
		expectedResp   *ssov1.RegisterResponse
	}{
		{
			name: "User registered successfully",
			inputReq: &ssov1.RegisterRequest{
				Email:    "test@test.com",
				Password: "12345678",
			},
			mockBehavior: func(s *mock_authgrpc.MockAuthService, req *ssov1.RegisterRequest) {
				s.EXPECT().Register(gomock.Any(), req.GetEmail(), req.GetPassword()).Return(1, nil)
			},
			expectedStatus: codes.OK,
			expectedResp:   &ssov1.RegisterResponse{UserId: 1},
		},
		{
			name: "User is already registered",
			inputReq: &ssov1.RegisterRequest{
				Email:    "test@test.com",
				Password: "12345678",
			},
			mockBehavior: func(s *mock_authgrpc.MockAuthService, req *ssov1.RegisterRequest) {
				s.EXPECT().Register(gomock.Any(), req.GetEmail(), req.GetPassword()).Return(-1, errs.ErrAlreadyExists)
			},
			expectedStatus: codes.AlreadyExists,
			expectedResp:   nil,
		},
		{
			name: "Invalid email format",
			inputReq: &ssov1.RegisterRequest{
				Email:    "invalid-email",
				Password: "12345678",
			},
			mockBehavior:   func(s *mock_authgrpc.MockAuthService, req *ssov1.RegisterRequest) {},
			expectedStatus: codes.InvalidArgument,
			expectedResp:   nil,
		},
		{
			name: "Invalid password format",
			inputReq: &ssov1.RegisterRequest{
				Email:    "test@test.com",
				Password: "12345",
			},
			mockBehavior:   func(s *mock_authgrpc.MockAuthService, req *ssov1.RegisterRequest) {},
			expectedStatus: codes.InvalidArgument,
			expectedResp:   nil,
		},
		{
			name:           "Nil request",
			inputReq:       nil,
			mockBehavior:   func(s *mock_authgrpc.MockAuthService, req *ssov1.RegisterRequest) {},
			expectedStatus: codes.InvalidArgument,
			expectedResp:   nil,
		},
		{
			name: "Internal error",
			inputReq: &ssov1.RegisterRequest{
				Email:    "test@test.com",
				Password: "12345678",
			},
			mockBehavior: func(s *mock_authgrpc.MockAuthService, req *ssov1.RegisterRequest) {
				s.EXPECT().Register(gomock.Any(), req.GetEmail(), req.GetPassword()).Return(-1, fmt.Errorf("some internal error"))
			},
			expectedStatus: codes.Internal,
			expectedResp:   nil,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			authService := mock_authgrpc.NewMockAuthService(c)
			testCase.mockBehavior(authService, testCase.inputReq)

			log := logger.NewTestLogger()
			ctx := logger.ContextWithLogger(t.Context(), log)

			authGRPC := NewAuthGRPC(authService)
			resp, err := authGRPC.Register(ctx, testCase.inputReq)

			assert.Equal(t, testCase.expectedStatus, status.Code(err))
			assert.Equal(t, testCase.expectedResp, resp)
		})
	}
}

func TestLogin(t *testing.T) {
	type mockBehavior func(s *mock_authgrpc.MockAuthService, req *ssov1.LoginRequest)

	testTable := []struct {
		name           string
		inputReq       *ssov1.LoginRequest
		mockBehavior   mockBehavior
		expectedStatus codes.Code
		expectedResp   *ssov1.LoginResponse
	}{
		{
			name: "User logged in successfully",
			inputReq: &ssov1.LoginRequest{
				Email:    "test@test.com",
				Password: "12345678",
				AppId:    1,
			},
			mockBehavior: func(s *mock_authgrpc.MockAuthService, req *ssov1.LoginRequest) {
				s.EXPECT().Login(gomock.Any(), req.GetEmail(), req.GetPassword(), int(req.GetAppId())).Return(
					domain.SessionShort{
						AccessToken:  "test.test.test",
						RefreshToken: "test",
						TTL:          30 * time.Second,
					},
					nil,
				)
			},
			expectedStatus: codes.OK,
			expectedResp: &ssov1.LoginResponse{
				AccessToken:  "test.test.test",
				RefreshToken: "test",
				SessionTtl: &ssov1.Duration{
					Seconds: int64(30),
				},
			},
		},
		{
			name: "Invalid credentials",
			inputReq: &ssov1.LoginRequest{
				Email:    "test@test.com",
				Password: "12345678",
				AppId:    1,
			},
			mockBehavior: func(s *mock_authgrpc.MockAuthService, req *ssov1.LoginRequest) {
				s.EXPECT().Login(gomock.Any(), req.GetEmail(), req.GetPassword(), int(req.GetAppId())).Return(
					domain.SessionShort{},
					errs.ErrInvalidCredentials,
				)
			},
			expectedStatus: codes.Unauthenticated,
			expectedResp:   nil,
		},
		{
			name: "App JWT key not configured",
			inputReq: &ssov1.LoginRequest{
				Email:    "test@test.com",
				Password: "12345678",
				AppId:    1,
			},
			mockBehavior: func(s *mock_authgrpc.MockAuthService, req *ssov1.LoginRequest) {
				s.EXPECT().Login(gomock.Any(), req.GetEmail(), req.GetPassword(), int(req.GetAppId())).Return(
					domain.SessionShort{},
					errs.ErrKeyNotConfigured,
				)
			},
			expectedStatus: codes.Internal,
			expectedResp:   nil,
		},
		{
			name: "Invalid email format",
			inputReq: &ssov1.LoginRequest{
				Email:    "invalid-format",
				Password: "12345678",
				AppId:    1,
			},
			mockBehavior:   func(s *mock_authgrpc.MockAuthService, req *ssov1.LoginRequest) {},
			expectedStatus: codes.InvalidArgument,
			expectedResp:   nil,
		},
		{
			name: "Invalid password format",
			inputReq: &ssov1.LoginRequest{
				Email:    "test@test.com",
				Password: "123456",
				AppId:    1,
			},
			mockBehavior:   func(s *mock_authgrpc.MockAuthService, req *ssov1.LoginRequest) {},
			expectedStatus: codes.InvalidArgument,
			expectedResp:   nil,
		},
		{
			name: "Invalid app ID format",
			inputReq: &ssov1.LoginRequest{
				Email:    "test@test.com",
				Password: "12345678",
				AppId:    -1,
			},
			mockBehavior:   func(s *mock_authgrpc.MockAuthService, req *ssov1.LoginRequest) {},
			expectedStatus: codes.InvalidArgument,
			expectedResp:   nil,
		},
		{
			name:           "Nil request",
			inputReq:       nil,
			mockBehavior:   func(s *mock_authgrpc.MockAuthService, req *ssov1.LoginRequest) {},
			expectedStatus: codes.InvalidArgument,
			expectedResp:   nil,
		},
		{
			name: "Internal error",
			inputReq: &ssov1.LoginRequest{
				Email:    "test@test.com",
				Password: "12345678",
				AppId:    1,
			},
			mockBehavior: func(s *mock_authgrpc.MockAuthService, req *ssov1.LoginRequest) {
				s.EXPECT().Login(gomock.Any(), req.GetEmail(), req.GetPassword(), int(req.GetAppId())).Return(
					domain.SessionShort{},
					fmt.Errorf("some internal error"),
				)
			},
			expectedStatus: codes.Internal,
			expectedResp:   nil,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			authService := mock_authgrpc.NewMockAuthService(c)
			testCase.mockBehavior(authService, testCase.inputReq)

			log := logger.NewTestLogger()
			ctx := logger.ContextWithLogger(t.Context(), log)

			authGRPC := NewAuthGRPC(authService)
			resp, err := authGRPC.Login(ctx, testCase.inputReq)

			assert.Equal(t, testCase.expectedStatus, status.Code(err))
			assert.Equal(t, testCase.expectedResp, resp)
		})
	}
}

func TestRefresh(t *testing.T) {
	type mockBehavior func(s *mock_authgrpc.MockAuthService, req *ssov1.RefreshRequest)

	testTable := []struct {
		name           string
		inputReq       *ssov1.RefreshRequest
		mockBehavior   mockBehavior
		expectedStatus codes.Code
		expectedResp   *ssov1.RefreshResponse
	}{
		{
			name: "Session refreshed successfully",
			inputReq: &ssov1.RefreshRequest{
				RefreshToken: "80ARI/q1v7B4btJFr8RT+FvGRC3pL/ht2h8QAdf7dyY=",
				AppId:        1,
			},
			mockBehavior: func(s *mock_authgrpc.MockAuthService, req *ssov1.RefreshRequest) {
				s.EXPECT().Refresh(gomock.Any(), req.GetRefreshToken(), int(req.GetAppId())).Return(
					domain.SessionShort{
						AccessToken:  "test.test.test",
						RefreshToken: "90ARI/q1v7B4btJFr8RT+FvGRC3pL/ht2h8QAdf7dyY=",
						TTL:          30 * time.Second,
					},
					nil,
				)
			},
			expectedStatus: codes.OK,
			expectedResp: &ssov1.RefreshResponse{
				AccessToken:  "test.test.test",
				RefreshToken: "90ARI/q1v7B4btJFr8RT+FvGRC3pL/ht2h8QAdf7dyY=",
				SessionTtl: &ssov1.Duration{
					Seconds: int64(30),
				},
			},
		},
		{
			name: "Failed to refresh session",
			inputReq: &ssov1.RefreshRequest{
				RefreshToken: "80ARI/q1v7B4btJFr8RT+FvGRC3pL/ht2h8QAdf7dyY=",
				AppId:        1,
			},
			mockBehavior: func(s *mock_authgrpc.MockAuthService, req *ssov1.RefreshRequest) {
				s.EXPECT().Refresh(gomock.Any(), req.GetRefreshToken(), int(req.GetAppId())).Return(
					domain.SessionShort{},
					errs.ErrRefSession,
				)
			},
			expectedStatus: codes.Unauthenticated,
			expectedResp:   nil,
		},
		{
			name: "App not found",
			inputReq: &ssov1.RefreshRequest{
				RefreshToken: "80ARI/q1v7B4btJFr8RT+FvGRC3pL/ht2h8QAdf7dyY=",
				AppId:        1,
			},
			mockBehavior: func(s *mock_authgrpc.MockAuthService, req *ssov1.RefreshRequest) {
				s.EXPECT().Refresh(gomock.Any(), req.GetRefreshToken(), int(req.GetAppId())).Return(
					domain.SessionShort{},
					errs.ErrNotFound,
				)
			},
			expectedStatus: codes.NotFound,
			expectedResp:   nil,
		},
		{
			name: "Invalid refresh token format",
			inputReq: &ssov1.RefreshRequest{
				RefreshToken: "8ARI/q1v7B4=",
				AppId:        1,
			},
			mockBehavior:   func(s *mock_authgrpc.MockAuthService, req *ssov1.RefreshRequest) {},
			expectedStatus: codes.InvalidArgument,
			expectedResp:   nil,
		},
		{
			name: "Invalid app ID format",
			inputReq: &ssov1.RefreshRequest{
				RefreshToken: "80ARI/q1v7B4btJFr8RT+FvGRC3pL/ht2h8QAdf7dyY=",
				AppId:        -1,
			},
			mockBehavior:   func(s *mock_authgrpc.MockAuthService, req *ssov1.RefreshRequest) {},
			expectedStatus: codes.InvalidArgument,
			expectedResp:   nil,
		},
		{
			name: "Internal error",
			inputReq: &ssov1.RefreshRequest{
				RefreshToken: "80ARI/q1v7B4btJFr8RT+FvGRC3pL/ht2h8QAdf7dyY=",
				AppId:        1,
			},
			mockBehavior: func(s *mock_authgrpc.MockAuthService, req *ssov1.RefreshRequest) {
				s.EXPECT().Refresh(gomock.Any(), req.GetRefreshToken(), int(req.GetAppId())).Return(
					domain.SessionShort{},
					fmt.Errorf("some internal error"),
				)
			},
			expectedStatus: codes.Internal,
			expectedResp:   nil,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			authService := mock_authgrpc.NewMockAuthService(c)
			testCase.mockBehavior(authService, testCase.inputReq)

			log := logger.NewTestLogger()
			ctx := logger.ContextWithLogger(t.Context(), log)

			authGRPC := NewAuthGRPC(authService)
			resp, err := authGRPC.Refresh(ctx, testCase.inputReq)

			assert.Equal(t, testCase.expectedStatus, status.Code(err))
			assert.Equal(t, testCase.expectedResp, resp)
		})
	}
}

func TestLogout(t *testing.T) {
	type mockBehavior func(s *mock_authgrpc.MockAuthService, req *ssov1.LogoutRequest)

	testTable := []struct {
		name           string
		inputReq       *ssov1.LogoutRequest
		mockBehavior   mockBehavior
		expectedStatus codes.Code
		expectedResp   *ssov1.Empty
	}{
		{
			name: "User logged out successfully",
			inputReq: &ssov1.LogoutRequest{
				AccessToken:  "test.test.test",
				RefreshToken: "80ARI/q1v7B4btJFr8RT+FvGRC3pL/ht2h8QAdf7dyY=",
			},
			mockBehavior: func(s *mock_authgrpc.MockAuthService, req *ssov1.LogoutRequest) {
				s.EXPECT().Logout(gomock.Any(), req.GetRefreshToken(), req.GetAccessToken()).Return(nil)
			},
			expectedStatus: codes.OK,
			expectedResp:   &ssov1.Empty{},
		},
		{
			name: "Session not found",
			inputReq: &ssov1.LogoutRequest{
				AccessToken:  "test.test.test",
				RefreshToken: "80ARI/q1v7B4btJFr8RT+FvGRC3pL/ht2h8QAdf7dyY=",
			},
			mockBehavior: func(s *mock_authgrpc.MockAuthService, req *ssov1.LogoutRequest) {
				s.EXPECT().Logout(gomock.Any(), req.GetRefreshToken(), req.GetAccessToken()).Return(errs.ErrNotFound)
			},
			expectedStatus: codes.NotFound,
			expectedResp:   nil,
		},
		{
			name: "Internal error",
			inputReq: &ssov1.LogoutRequest{
				AccessToken:  "test.test.test",
				RefreshToken: "80ARI/q1v7B4btJFr8RT+FvGRC3pL/ht2h8QAdf7dyY=",
			},
			mockBehavior: func(s *mock_authgrpc.MockAuthService, req *ssov1.LogoutRequest) {
				s.EXPECT().Logout(gomock.Any(), req.GetRefreshToken(), req.GetAccessToken()).Return(fmt.Errorf("some internal error"))
			},
			expectedStatus: codes.Internal,
			expectedResp:   nil,
		},
		{
			name: "Invalid refresh token format",
			inputReq: &ssov1.LogoutRequest{
				AccessToken:  "test.test.test",
				RefreshToken: "80ARI/q1v7QAdf7dyY=",
			},
			mockBehavior:   func(s *mock_authgrpc.MockAuthService, req *ssov1.LogoutRequest) {},
			expectedStatus: codes.InvalidArgument,
			expectedResp:   nil,
		},
		{
			name:           "Nil request",
			inputReq:       nil,
			mockBehavior:   func(s *mock_authgrpc.MockAuthService, req *ssov1.LogoutRequest) {},
			expectedStatus: codes.InvalidArgument,
			expectedResp:   nil,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			authService := mock_authgrpc.NewMockAuthService(c)
			testCase.mockBehavior(authService, testCase.inputReq)

			log := logger.NewTestLogger()
			ctx := logger.ContextWithLogger(t.Context(), log)

			authGRPC := NewAuthGRPC(authService)
			resp, err := authGRPC.Logout(ctx, testCase.inputReq)

			assert.Equal(t, testCase.expectedStatus, status.Code(err))
			assert.Equal(t, testCase.expectedResp, resp)
		})
	}
}
