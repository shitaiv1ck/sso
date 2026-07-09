package accgrpc

import (
	"fmt"
	"testing"

	ssov1 "github.com/shitaiv1ck/protos/gen/go/sso"
	errs "github.com/shitaiv1ck/sso/internal/core/errors"
	"github.com/shitaiv1ck/sso/internal/core/logger"
	mock_accgrpc "github.com/shitaiv1ck/sso/internal/features/account/transport/grpc/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestChangePassword(t *testing.T) {
	type mockBehavior func(s *mock_accgrpc.MockAccountService, req *ssov1.PasswordRequest)

	testTable := []struct {
		name           string
		inputReq       *ssov1.PasswordRequest
		mockBehavior   mockBehavior
		expectedStatus codes.Code
		expectedResp   *ssov1.Empty
	}{
		{
			name: "User changed password successfully",
			inputReq: &ssov1.PasswordRequest{
				UserId:      1,
				OldPassword: "12345678",
				NewPassword: "87654321",
			},
			mockBehavior: func(s *mock_accgrpc.MockAccountService, req *ssov1.PasswordRequest) {
				s.EXPECT().ChangePassword(gomock.Any(), int(req.GetUserId()), req.GetOldPassword(), req.GetNewPassword()).Return(nil)
			},
			expectedStatus: codes.OK,
			expectedResp:   &ssov1.Empty{},
		},
		{
			name: "Invalid old password",
			inputReq: &ssov1.PasswordRequest{
				UserId:      1,
				OldPassword: "12345678",
				NewPassword: "87654321",
			},
			mockBehavior: func(s *mock_accgrpc.MockAccountService, req *ssov1.PasswordRequest) {
				s.EXPECT().ChangePassword(gomock.Any(), int(req.GetUserId()), req.GetOldPassword(), req.GetNewPassword()).Return(errs.ErrInvalidArg)
			},
			expectedStatus: codes.InvalidArgument,
			expectedResp:   nil,
		},
		{
			name: "User not found",
			inputReq: &ssov1.PasswordRequest{
				UserId:      1,
				OldPassword: "12345678",
				NewPassword: "87654321",
			},
			mockBehavior: func(s *mock_accgrpc.MockAccountService, req *ssov1.PasswordRequest) {
				s.EXPECT().ChangePassword(gomock.Any(), int(req.GetUserId()), req.GetOldPassword(), req.GetNewPassword()).Return(errs.ErrNotFound)
			},
			expectedStatus: codes.NotFound,
			expectedResp:   nil,
		},
		{
			name: "Internal error",
			inputReq: &ssov1.PasswordRequest{
				UserId:      1,
				OldPassword: "12345678",
				NewPassword: "87654321",
			},
			mockBehavior: func(s *mock_accgrpc.MockAccountService, req *ssov1.PasswordRequest) {
				s.EXPECT().ChangePassword(gomock.Any(), int(req.GetUserId()), req.GetOldPassword(), req.GetNewPassword()).Return(fmt.Errorf("some internal error"))
			},
			expectedStatus: codes.Internal,
			expectedResp:   nil,
		},
		{
			name: "Invalid user ID format",
			inputReq: &ssov1.PasswordRequest{
				UserId:      -1,
				OldPassword: "12345678",
				NewPassword: "87654321",
			},
			mockBehavior:   func(s *mock_accgrpc.MockAccountService, req *ssov1.PasswordRequest) {},
			expectedStatus: codes.InvalidArgument,
			expectedResp:   nil,
		},
		{
			name: "Invalid old password format",
			inputReq: &ssov1.PasswordRequest{
				UserId:      1,
				OldPassword: "123456",
				NewPassword: "87654321",
			},
			mockBehavior:   func(s *mock_accgrpc.MockAccountService, req *ssov1.PasswordRequest) {},
			expectedStatus: codes.InvalidArgument,
			expectedResp:   nil,
		},
		{
			name: "Invalid new password format",
			inputReq: &ssov1.PasswordRequest{
				UserId:      1,
				OldPassword: "12345678",
				NewPassword: "876543",
			},
			mockBehavior:   func(s *mock_accgrpc.MockAccountService, req *ssov1.PasswordRequest) {},
			expectedStatus: codes.InvalidArgument,
			expectedResp:   nil,
		},
		{
			name: "Old password equal new password",
			inputReq: &ssov1.PasswordRequest{
				UserId:      1,
				OldPassword: "12345678",
				NewPassword: "12345678",
			},
			mockBehavior:   func(s *mock_accgrpc.MockAccountService, req *ssov1.PasswordRequest) {},
			expectedStatus: codes.InvalidArgument,
			expectedResp:   nil,
		},
		{
			name:           "Nil request",
			inputReq:       nil,
			mockBehavior:   func(s *mock_accgrpc.MockAccountService, req *ssov1.PasswordRequest) {},
			expectedStatus: codes.InvalidArgument,
			expectedResp:   nil,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			accountService := mock_accgrpc.NewMockAccountService(c)
			testCase.mockBehavior(accountService, testCase.inputReq)

			log := logger.NewTestLogger()
			ctx := logger.ContextWithLogger(t.Context(), log)

			accountGRPC := NewAccountGRPC(accountService)
			resp, err := accountGRPC.ChangePassword(ctx, testCase.inputReq)

			assert.Equal(t, testCase.expectedStatus, status.Code(err))
			assert.Equal(t, testCase.expectedResp, resp)
		})
	}
}

func TestChangeEmail(t *testing.T) {
	type mockBehavior func(s *mock_accgrpc.MockAccountService, req *ssov1.EmailRequest)

	testTable := []struct {
		name           string
		inputReq       *ssov1.EmailRequest
		mockBehavior   mockBehavior
		expectedStatus codes.Code
		expectedResp   *ssov1.Empty
	}{
		{
			name: "User changed email successfully",
			inputReq: &ssov1.EmailRequest{
				UserId:   1,
				Password: "12345678",
				NewEmail: "test@test.com",
			},
			mockBehavior: func(s *mock_accgrpc.MockAccountService, req *ssov1.EmailRequest) {
				s.EXPECT().ChangeEmail(gomock.Any(), int(req.GetUserId()), req.GetPassword(), req.GetNewEmail()).Return(nil)
			},
			expectedStatus: codes.OK,
			expectedResp:   &ssov1.Empty{},
		},
		{
			name: "Invalid password",
			inputReq: &ssov1.EmailRequest{
				UserId:   1,
				Password: "12345678",
				NewEmail: "test@test.com",
			},
			mockBehavior: func(s *mock_accgrpc.MockAccountService, req *ssov1.EmailRequest) {
				s.EXPECT().ChangeEmail(gomock.Any(), int(req.GetUserId()), req.GetPassword(), req.GetNewEmail()).Return(errs.ErrInvalidArg)
			},
			expectedStatus: codes.InvalidArgument,
			expectedResp:   nil,
		},
		{
			name: "Email not unique",
			inputReq: &ssov1.EmailRequest{
				UserId:   1,
				Password: "12345678",
				NewEmail: "test@test.com",
			},
			mockBehavior: func(s *mock_accgrpc.MockAccountService, req *ssov1.EmailRequest) {
				s.EXPECT().ChangeEmail(gomock.Any(), int(req.GetUserId()), req.GetPassword(), req.GetNewEmail()).Return(errs.ErrAlreadyExist)
			},
			expectedStatus: codes.AlreadyExists,
			expectedResp:   nil,
		},
		{
			name: "Internal error",
			inputReq: &ssov1.EmailRequest{
				UserId:   1,
				Password: "12345678",
				NewEmail: "test@test.com",
			},
			mockBehavior: func(s *mock_accgrpc.MockAccountService, req *ssov1.EmailRequest) {
				s.EXPECT().ChangeEmail(gomock.Any(), int(req.GetUserId()), req.GetPassword(), req.GetNewEmail()).Return(fmt.Errorf("some internal error"))
			},
			expectedStatus: codes.Internal,
			expectedResp:   nil,
		},
		{
			name: "User not found",
			inputReq: &ssov1.EmailRequest{
				UserId:   1,
				Password: "12345678",
				NewEmail: "test@test.com",
			},
			mockBehavior: func(s *mock_accgrpc.MockAccountService, req *ssov1.EmailRequest) {
				s.EXPECT().ChangeEmail(gomock.Any(), int(req.GetUserId()), req.GetPassword(), req.GetNewEmail()).Return(errs.ErrNotFound)
			},
			expectedStatus: codes.NotFound,
			expectedResp:   nil,
		},
		{
			name: "Invalid user ID format",
			inputReq: &ssov1.EmailRequest{
				UserId:   -1,
				Password: "12345678",
				NewEmail: "test@test.com",
			},
			mockBehavior:   func(s *mock_accgrpc.MockAccountService, req *ssov1.EmailRequest) {},
			expectedStatus: codes.InvalidArgument,
			expectedResp:   nil,
		},
		{
			name: "Invalid password format",
			inputReq: &ssov1.EmailRequest{
				UserId:   1,
				Password: "123456",
				NewEmail: "test@test.com",
			},
			mockBehavior:   func(s *mock_accgrpc.MockAccountService, req *ssov1.EmailRequest) {},
			expectedStatus: codes.InvalidArgument,
			expectedResp:   nil,
		},
		{
			name: "Invalid new email format",
			inputReq: &ssov1.EmailRequest{
				UserId:   1,
				Password: "12345678",
				NewEmail: "invalid-format",
			},
			mockBehavior:   func(s *mock_accgrpc.MockAccountService, req *ssov1.EmailRequest) {},
			expectedStatus: codes.InvalidArgument,
			expectedResp:   nil,
		},
		{
			name:           "Nil request",
			inputReq:       nil,
			mockBehavior:   func(s *mock_accgrpc.MockAccountService, req *ssov1.EmailRequest) {},
			expectedStatus: codes.InvalidArgument,
			expectedResp:   nil,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			accountService := mock_accgrpc.NewMockAccountService(c)
			testCase.mockBehavior(accountService, testCase.inputReq)

			log := logger.NewTestLogger()
			ctx := logger.ContextWithLogger(t.Context(), log)

			accountGRPC := NewAccountGRPC(accountService)
			resp, err := accountGRPC.ChangeEmail(ctx, testCase.inputReq)

			assert.Equal(t, testCase.expectedStatus, status.Code(err))
			assert.Equal(t, testCase.expectedResp, resp)
		})
	}
}
