package user

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"

	"user/internal/svc"
	"user/internal/types"
	"user/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginRequest) (resp *types.LoginResponse, err error) {
	// 1. 查询用户
	user, err := l.svcCtx.UsersModel.FindOneByUsername(l.ctx, req.Username)
	switch {
	case err == nil:
		// 用户存在，继续
	case errors.Is(err, model.ErrNotFound):
		l.Logger.Infof("登录失败: 用户名 %s 不存在", req.Username)
		return nil, errors.New("账号或密码错误")
	default:
		l.Logger.Errorf("查询用户失败: %v", err)
		return nil, errors.New("内部错误")
	}

	// 2. 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		l.Logger.Infof("登录失败: 用户 %s 密码错误", req.Username)
		return nil, errors.New("账号或密码错误")
	}

	// 3. 生成 JWT
	now := time.Now()
	accessExpire := now.Add(time.Duration(l.svcCtx.Config.Auth.AccessExpire) * time.Second)
	refreshAfter := now.Add(time.Duration(l.svcCtx.Config.Auth.AccessExpire/2) * time.Second)

	claims := jwt.MapClaims{
		"userId":   user.Id,
		"username": user.Username,
		"exp":      accessExpire.Unix(),
		"iat":      now.Unix(),
		"iss":      "user",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(l.svcCtx.Config.Auth.AccessSecret))
	if err != nil {
		l.Logger.Errorf("JWT 签名失败: %v", err)
		return nil, errors.New("内部错误")
	}

	return &types.LoginResponse{
		AccessToken:  signedToken,
		AccessExpire: accessExpire.Unix(),
		RefreshAfter: refreshAfter.Unix(),
	}, nil
}
