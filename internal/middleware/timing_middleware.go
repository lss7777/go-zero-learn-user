package middleware

import (
	"net/http"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

type TimingMiddleware struct {
}

func NewTimingMiddleware() *TimingMiddleware {
	return &TimingMiddleware{}
}

func (m *TimingMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next(w, r)
		logx.WithContext(r.Context()).Infof("请求耗时: %s %s %v", r.Method, r.URL.Path, time.Since(start))
	}
}
