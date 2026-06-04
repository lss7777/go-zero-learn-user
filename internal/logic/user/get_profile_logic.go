package user

import (
	"context"
	"encoding/json"
	"errors"

	"user/internal/svc"
	"user/internal/types"
	"user/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetProfileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetProfileLogic {
	return &GetProfileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetProfileLogic) GetProfile() (resp *types.GetProfileResponse, err error) {
	// 1. 从 JWT claims 中提取 userId
	userIdValue := l.ctx.Value("userId")
	if userIdValue == nil {
		return nil, errors.New("未授权访问")
	}

	// go-zero 使用 jwt.WithJSONNumber() 解析 token，数值字段类型为 json.Number
	userIdStr, ok := userIdValue.(json.Number)
	if !ok {
		l.Logger.Errorf("userId claim 类型异常: %T", userIdValue)
		return nil, errors.New("无效的令牌")
	}

	userId, err := userIdStr.Int64()
	if err != nil {
		l.Logger.Errorf("userId 解析失败: %v", err)
		return nil, errors.New("内部错误")
	}

	// 2. 查询用户
	user, err := l.svcCtx.UsersModel.FindOne(l.ctx, userId)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			l.Logger.Infof("用户不存在: %d", userId)
			return nil, errors.New("用户不存在")
		}
		l.Logger.Errorf("查询用户失败: %v", err)
		return nil, errors.New("内部错误")
	}

	// 3. 返回（不含 Password）
	return &types.GetProfileResponse{
		Id:       user.Id,
		Username: user.Username,
	}, nil
}
