package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/David89Lee/pyrolysis_ty_server/middleware"
	"github.com/David89Lee/pyrolysis_ty_server/model"
	"github.com/David89Lee/pyrolysis_ty_server/router/response"
)

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResult struct {
	Token    string `json:"token"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
}

// @Summary 登录
// @Tags 用户登录
// @Accept json
// @Produce  json
// @Param body body LoginRequest true "用户名密码"
// @Success 200 {object} response.Response{data=LoginResult}
// @Router /api/v1/auth/login [post]
func Login(c *gin.Context) {
	appG := response.Gin{C: c}

	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appG.Response(http.StatusOK, response.InvalidParams, err.Error())
		return
	}

	user, err := model.FindUserByUsername(req.Username)
	if err != nil {
		zap.L().Error("查询用户失败", zap.Error(err))
		appG.Response(http.StatusOK, response.ERROR, nil)
		return
	}
	if user == nil || !user.CheckPassword(req.Password) {
		appG.Response(http.StatusOK, response.ErrorLoginFail, nil)
		return
	}

	token, err := middleware.GenerateToken(user.Username)
	if err != nil {
		zap.L().Error("生成token失败", zap.Error(err))
		appG.Response(http.StatusOK, response.ErrorGenerateToken, nil)
		return
	}
	appG.Response(http.StatusOK, response.SUCCESS, LoginResult{
		Token:    token,
		Username: user.Username,
		Nickname: user.Nickname,
	})
}
