package routers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"uasd-wallet-manager/controllers"
)

func Adder(r *gin.Engine) {
	r.Use(Cors()) //解决跨域
	lottery := r.Group("/wallet")
	{

		lottery.POST("/getWallet", controllers.WalletController{}.GetWallet)

	}
}

func Cors() gin.HandlerFunc {
	return func(context *gin.Context) {
		method := context.Request.Method
		context.Header("Access-Control-Allow-Origin", "*")
		context.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization, Token")
		context.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		context.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		context.Header("Access-Control-Allow-Credentials", "true")
		if method == "OPTIONS" {
			context.AbortWithStatus(http.StatusNoContent)
		}
		context.Next()
	}
}
