package user

import (
	"context"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"

	"user/internal/svc"
	"user/internal/types"
	"user/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegisterLogic) Register(req *types.RegisterRequest) (resp *types.RegisterResponse, err error) {
	// 1. 校验两次密码是否一致
	if req.Password != req.ConfirmPassword {
		return nil, errors.New("两次密码不一致")
	}

	// 2. 检查用户名是否已存在
	_, err = l.svcCtx.UsersModel.FindOneByUsername(l.ctx, req.Username)
	switch {
	case err == nil:
		l.Logger.Infof("用户名已存在")
		return nil, errors.New("用户名已存在")
	case errors.Is(err, model.ErrNotFound):
		// 用户名可用，继续
	default:
		l.Logger.Errorf("查询用户名失败: %v", err)
		return nil, errors.New("内部错误")
	}

	// 3. 密码哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		l.Logger.Errorf("密码哈希失败: %v", err)
		return nil, errors.New("内部错误")
	}

	// 4. 生成用户ID（基于时间戳，后续替换为雪花算法）
	userId := time.Now().UnixMilli()

	// 5. 写入数据库
	_, err = l.svcCtx.UsersModel.Insert(l.ctx, &model.Users{
		Id:       userId,
		Username: req.Username,
		Password: string(hashedPassword),
	})
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			// MySQL 1062 Duplicate entry → 并发下被他人抢先
			return nil, errors.New("用户名已存在")
		}
		l.Logger.Errorf("插入用户失败: %v", err)
		return nil, errors.New("内部错误")
	}

	return &types.RegisterResponse{
		Id:       userId,
		Username: req.Username,
	}, nil
}
