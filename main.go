package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"database/sql"
	"math/rand"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"

	"github.com/joho/godotenv"
)

// リクエストパラメータをマッピングする構造体を定義する
type UserRequest struct {
	UserId string `json:"userid" binding:"required"`
	Name   string `json:"name" binding:"required"`
	Age    int    `json:"age" binding:"required,min=0"`
}

type User struct {
	ID     int    `json:"id"`
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}

type CreateUserRequest struct {
	UserID   string `json:"user_id" binding:"required"`
	Role     string `json:"role" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UpdateUserRequest struct {
	Role string `json:"role" binding:"required"`
}

func main() {
	// .env読みこみ
	if err := godotenv.Load(); err != nil {
		os.Exit(1)
	}
	// DB関係
	db_user := os.Getenv("DB_USER")
	db_pass := os.Getenv("DB_PASS")
	db_POST := os.Getenv("DB_POST")

	// Ginのルーターを作って便利なミドルウェアを自動でつけてくれる（DefaultはLogger():アクセスログをつけるミドルウェアとRecovery():panicでもサーバを落とさない）
	router := gin.Default()
	// データベースの接続情報
	dsn := db_user + ":" + db_pass + "@tcp(" + db_POST + ")/lesson_go_db"

	// データベースへの接続（接続テスト）
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 実際に接続可能か確認する
	err = db.Ping()
	if err != nil {
		log.Fatal("データベースに接続できませんでした：", err)
	}
	fmt.Println("MariaDBへの接続に成功")

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

	router.GET("/search", func(ctx *gin.Context) {
		rows, err := db.Query("SELECT id, user_id, role FROM t_user")
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var users []User
		for rows.Next() {
			var user User
			if err := rows.Scan(&user.ID, &user.UserID, &user.Role); err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			users = append(users, user)
		}

		if err := rows.Err(); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"users": users,
		})
	})

	// 新規レコード登録
	router.POST("/users", func(ctx *gin.Context) {
		var req CreateUserRequest
		if err := ctx.ShouldBindBodyWithJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		// パスワードをハッシュ化する
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "パスワードのハッシュ化に失敗しました"})
			return
		}

		// Go-langはデフォルトでプリペアドステートメント有効。（SQLはJava Servletのように命令すれば良い。）
		result, err := db.Exec("INSERT INTO t_user (user_id, role, password) VALUES (?, ?, ?)", req.UserID, req.Role, string(hashedPassword))
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		id, err := result.LastInsertId()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusCreated, gin.H{
			"message": "ユーザを作成しました",
			"id":      id,
			"user_id": req.UserID,
			"role":    req.Role,
		})
	})

	// レコード更新
	router.PUT("/users/:id", func(ctx *gin.Context) {
		idParam := ctx.Param("id")
		id, err := strconv.Atoi(idParam)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "idは数値で指定してください"})
			return
		}

		var req UpdateUserRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		result, err := db.Exec("UPDATE t_user SET role = ? WHERE id = ?", req.Role, id)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if rowsAffected == 0 {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "対象のユーザが見つかりません"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"message": "ユーザを更新しました",
			"id":      id,
			"role":    req.Role,
		})
	})

	// レコード削除
	router.DELETE("/users/:id", func(ctx *gin.Context) {
		idParam := ctx.Param("id")
		id, err := strconv.Atoi(idParam)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "idは数値で指定してください"})
			return
		}

		result, err := db.Exec("DELETE FROM t_user WHERE id = ?", id)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if rowsAffected == 0 {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "対象のユーザが見つかりません"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"message": "ユーザを削除しました",
			"id":      id,
		})
	})

	router.Run(":8080")
}
