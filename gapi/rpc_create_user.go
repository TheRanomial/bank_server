package gapi

import (
	"context"
	"time"

	db "github.com/TheRanomial/bank_server/db/sqlc"
	"github.com/TheRanomial/bank_server/pb"
	"github.com/TheRanomial/bank_server/util"
	"github.com/TheRanomial/bank_server/val"
	"github.com/TheRanomial/bank_server/worker"
	"github.com/hibiken/asynq"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) CreateUser(ctx context.Context,req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {

	violations:=validateCreateUserRequest(req)
	if violations!=nil{
		return nil,invalidArgumentError(violations)
	}

	createHashedPassword,err:=util.HashPassword(req.GetPassword())
	if err!=nil {
		return nil,status.Errorf(codes.Internal,"Failed to hash password")
	}

	arg:=db.CreateUserTxParams{
		CreateUserParams:db.CreateUserParams{
			Username:	req.GetUsername(),      
			HashedPassword:		createHashedPassword,   
			FullName:	req.GetFullName(),      
			Email:	req.GetEmail(),
		},
		AfterCreate: func(user db.User) error {
			taskPayload:=&worker.PayloadSendVerifyEmail{
				Username: user.Username,
			}
			opts:=[]asynq.Option{
				asynq.MaxRetry(10),
				asynq.ProcessIn(10*time.Second),
			}
			return s.taskDistributor.DistributeTaskSendVerifyEmail(ctx,taskPayload,opts...)
		},
	}

	txResult, err := s.store.CreateUserTx(ctx,arg)
	if err != nil {
		errCode := db.ErrorCode(err)
		if errCode == db.ForeignKeyViolation || errCode == db.UniqueViolation {
			return nil,status.Errorf(codes.AlreadyExists,"username or email already exists")
		}
		return nil,status.Errorf(codes.Internal,"Failed to create user")
	}

	res:=&pb.CreateUserResponse{
		User: ConvertUser(txResult.User),
	}
	return res, nil
}

func validateCreateUserRequest(req *pb.CreateUserRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := val.ValidateUsername(req.GetUsername()); err != nil {
		violations = append(violations, fieldViolation("username", err))
	}

	if err := val.ValidatePassword(req.GetPassword()); err != nil {
		violations = append(violations, fieldViolation("password", err))
	}

	if err := val.ValidateFullName(req.GetFullName()); err != nil {
		violations = append(violations, fieldViolation("full_name", err))
	}

	if err := val.ValidateEmail(req.GetEmail()); err != nil {
		violations = append(violations, fieldViolation("email", err))
	}

	return violations
}