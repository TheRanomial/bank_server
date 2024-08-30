package gapi

import (
	"context"
	"database/sql"

	db "github.com/TheRanomial/bank_server/db/sqlc"
	"github.com/TheRanomial/bank_server/pb"
	"github.com/TheRanomial/bank_server/util"
	"github.com/TheRanomial/bank_server/val"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Server) LoginUser(ctx context.Context,req *pb.LoginUserRequest) (*pb.LoginUserResponse,error){

	violations:=validateLoginUserRequest(req)
	if violations!=nil{
		return nil,invalidArgumentError(violations)
	}

	user,err:=s.store.GetUser(ctx,req.Username)
	if err!=nil{
		if err==sql.ErrNoRows{
			return nil,status.Errorf(codes.Internal,"The given user doesn't exist")
		}
		return nil,status.Errorf(codes.Internal,"Failed to fetch user")
	}

	err=util.CheckPassword(req.Password,user.HashedPassword)
	if err!=nil{
		return nil,status.Errorf(codes.InvalidArgument,"Given password is incorrect")
	}

	accessToken, accessPayload, err := s.tokenMaker.CreateToken(
		user.Username,
		s.config.AccessTokenDuration,
	)

	if err != nil {
		return nil,status.Errorf(codes.Internal,"Failed to create token")
	}

	refreshToken,refreshPayload,err:=s.tokenMaker.CreateToken(
		user.Username,
		s.config.RefreshTokenDuration,
	)

	if err!=nil{
		return nil,status.Errorf(codes.Internal,"Failed to create token")
	}

	mtdt:=s.extractMetadata(ctx)
	session,err:=s.store.CreateSession(ctx,db.CreateSessionParams{
		ID:           refreshPayload.ID,
		Username:     user.Username,
		RefreshToken: refreshToken,
		UserAgent:    mtdt.UserAgent,
		ClientIp:     mtdt.ClientIP,
		IsBlocked:    false,
		ExpiresAt:    refreshPayload.ExpiredAt,
	})

	if err!=nil{
		return nil,status.Errorf(codes.Unimplemented,"Failed to create session")
	}

	rsp := &pb.LoginUserResponse{
		User:                  ConvertUser(user),
		SessionId:             session.ID.String(),
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: timestamppb.New(refreshPayload.ExpiredAt),
		AccessTokenExpiresAt:  timestamppb.New(accessPayload.ExpiredAt),
	}
	
	return rsp,nil
}

func validateLoginUserRequest(req *pb.LoginUserRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := val.ValidateUsername(req.GetUsername()); err != nil {
		violations = append(violations, fieldViolation("username", err))
	}

	if err := val.ValidatePassword(req.GetPassword()); err != nil {
		violations = append(violations, fieldViolation("password", err))
	}
	return violations
}