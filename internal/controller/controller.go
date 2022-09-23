package controller

import (
	"GrabVotes/internal/logic"
	"GrabVotes/internal/pkg/jwt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

const (
	UserIdKey   = "userID"
	TicketIdKey = "ticketID"
)

func JWTAuth(c *gin.Context) {

	token := c.PostForm("token")
	if token == "" {
		c.JSON(http.StatusOK, "需要登录")
		c.Abort()
		return
	}
	mc, tokenValid, err := jwt.ParseToken(token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, "服务繁忙")
		c.Abort()
		return
	}
	if !tokenValid {
		c.JSON(http.StatusOK, "无效token，请登录后重试")
		c.Abort()
		return
	}
	// 将当前请求的userID信息保存到请求的上下文c上
	c.Set(UserIdKey, mc.UserID)
	c.Next() // 后续的处理函数可以用过c.Get("userID")来获取当前请求的用户信息

}

func GrabAction(c *gin.Context) {
	userID := c.GetString(UserIdKey)
	//  还要获取要抢哪张票的信息
	ticketID := c.PostForm(TicketIdKey)
	successSign, err := logic.GrabAction(userID, ticketID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, "服务繁忙")
		log.Println("抢票异常：", err.Error())
		c.Abort()
		return
	}
	if successSign {
		c.JSON(http.StatusOK, "抢票成功")
		return
	}
	c.JSON(http.StatusOK, "抢票失败")
}

// DefrayAction 用户支付
func DefrayAction(c *gin.Context) {

}
