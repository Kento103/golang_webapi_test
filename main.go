package main

import (
	"net/http"

	"math/rand"

	"github.com/gin-gonic/gin"
)

// リクエストパラメータをマッピングする構造体を定義する
type UserRequest struct {
	UserId string `json:"userid" binding:"required"`
	Name   string `json:"name" binding:"required"`
	Age    int    `json:"age" binding:"required,min=0"`
}

func main() {
	// Ginのルーターを作って便利なミドルウェアを自動でつけてくれる（DefaultはLogger():アクセスログをつけるミドルウェアとRecovery():panicでもサーバを落とさない）
	router := gin.Default()

	router.GET("/", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "Helloworld",
		})
	})

	router.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	router.GET("/echo", func(ctx *gin.Context) {
		// クエリパラメーターを取得する（存在しない場合は空文字となる）
		message := ctx.Query("message")
		// クエリパラメータを取得する（存在しない場合は右のデフォルトを使用する
		page := ctx.DefaultQuery("page", "1")

		ctx.JSON(http.StatusOK, gin.H{
			"message": message,
			"page":    page,
		})
	})

	router.POST("/create", func(ctx *gin.Context) {
		var request UserRequest
		uid := rand.Intn(1000)

		// JSONリクエストボディを構造体でバインドする
		if err := ctx.ShouldBindJSON(&request); err != nil {
			// バリデーションエラーのときは400を返す
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		// 処理結果をJSONで返す
		ctx.JSON(http.StatusOK, gin.H{
			"message":         "ユーザを作成しました",
			"uid":             uid,
			"received_userid": request.UserId,
			"received_name":   request.Name,
			"recived_age":     request.Age,
		})
	})

	router.Run(":8080")
}
