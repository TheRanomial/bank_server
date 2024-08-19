package api

import (
	"database/sql"
	"net/http"
	"time"

	db "github.com/TheRanomial/bank_server/db/sqlc"
	"github.com/TheRanomial/bank_server/util"
	"github.com/gin-gonic/gin"
)

type CreateUserRequest struct {
	Username       string `json:"username" binding:"required,alphanum"`
	Password 	   string `json:"password" binding:"required,min=6"`
	FullName       string `json:"full_name" binding:"required"`
	Email          string `json:"email" binding:"required,email"`
}

type userResponse struct {
	Username       	  string `json:"username"`
	FullName          string `json:"full_name"`
	Email             string `json:"email"`
	PasswordChangedAt time.Time  `json:"password_changed_at"`
	CreatedAt         time.Time	`json:"created_at"`
}

func newUserResponse(user db.User) userResponse {
	return userResponse{
		Username:          user.Username,
		FullName:          user.FullName,
		Email:             user.Email,
		PasswordChangedAt: user.PasswordChangedAt,
		CreatedAt:         user.CreatedAt,
	}
}

func (s *Server) CreateUser(ctx *gin.Context) {
	req:=CreateUserRequest{}

	if err:=ctx.ShouldBindJSON(&req); err!=nil {
		ctx.JSON(http.StatusBadRequest,errorResponse(err))
		return
	}
	
	createHashedPassword,err:=util.HashPassword(req.Password)
	if err!=nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	arg:=db.CreateUserParams{
		Username:	req.Username,      
		HashedPassword:		createHashedPassword,   
		FullName:	req.FullName,      
		Email:	req.Email,
	}

	user, err := s.store.CreateUser(ctx, arg)
	if err != nil {
		errCode := db.ErrorCode(err)
		if errCode == db.ForeignKeyViolation || errCode == db.UniqueViolation {
			ctx.JSON(http.StatusForbidden, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	resp:=newUserResponse(user)
	
	ctx.JSON(http.StatusOK,resp)
}

type loginUserRequest struct {
	Username       string `json:"username" binding:"required,alphanum"`
	Password 	   string `json:"password" binding:"required,min=6"`
}

type loginUserResponse struct {
	AccessToken   string    	`json:"access_token"`
	User 		  userResponse  `json:"user"`
}

func (s *Server) LoginUser(ctx *gin.Context){
	req:=loginUserRequest{}

	if err:=ctx.ShouldBindJSON(&req); err!=nil {
		ctx.JSON(http.StatusBadRequest,errorResponse(err))
		return
	}

	user,err:=s.store.GetUser(ctx,req.Username)
	if err!=nil{
		if err==sql.ErrNoRows{
			ctx.JSON(http.StatusNotFound,errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	err=util.CheckPassword(req.Password,user.HashedPassword)
	if err!=nil{
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	accessToken,_,err:=s.tokenMaker.CreateToken(
		user.Username,
		s.config.AccessTokenDuration,
	)

	if err!=nil{
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	resp:=loginUserResponse{
		AccessToken: accessToken,
		User: newUserResponse(user),
	}
	
	ctx.JSON(http.StatusOK,resp)
}