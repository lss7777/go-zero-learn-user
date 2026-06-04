// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package svc

import (
	"user/internal/config"
	"user/internal/middleware"
	"user/model"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/rest"
)

type ServiceContext struct {
	Config     config.Config
	UsersModel model.UsersModel
	Timing     rest.Middleware
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.MySQL.DataSource)
	return &ServiceContext{
		Config: c,
		UsersModel: model.NewUsersModel(conn, cache.CacheConf{
			{
				RedisConf: c.Redis,
				Weight:    100,
			},
		}),
		Timing: middleware.NewTimingMiddleware().Handle,
	}
}
