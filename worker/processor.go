package worker

import (
	"context"

	db "github.com/TheRanomial/bank_server/db/sqlc"
	"github.com/TheRanomial/bank_server/mail"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

type TaskProcessor interface {
	Start() error
	ProcessTaskSendVerifyEmail(ctx context.Context,task *asynq.Task) error
}

type RedisTaskProcessor struct {
	server *asynq.Server
	store db.Store
	mailer mail.EmailSender
}

func NewRedisTaskProcessor(redisOpt asynq.RedisClientOpt,store db.Store,mailer mail.EmailSender) TaskProcessor{
	s:=asynq.NewServer(
		redisOpt,
		asynq.Config{
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				log.Error().Err(err).Str("type",task.Type()).Msg("process task failed")
			}),
			Logger: NewLogger(),
		},
	)
	return &RedisTaskProcessor{
		server: s,
		store: store,
		mailer: mailer,
	}
}

func (p *RedisTaskProcessor) Start() error {
	mux:=asynq.NewServeMux()

	mux.HandleFunc(TaskSendVerifyEmail,p.ProcessTaskSendVerifyEmail)

	return p.server.Start(mux)
}

