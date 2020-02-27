package main

import (
	"log"
	"net/http"
	"github.com/weihongguo/gglmm"

	_ "github.com/jinzhu/gorm/dialects/mysql"
)

// Example --
type Example struct {
	gglmm.Model
	Example string `json:"example"`
}

func main() {
	gglmm.RegisterGormDB("mysql", "example:123456@(127.0.0.1:3306)/example?charset=utf8mb4&parseTime=true&loc=UTC", 10, 5, 600)
	defer gglmm.CloseGormDB()

	gglmm.RegisterRedisCacher("tcp", "127.0.0.1:6379", 10, 5, 3)
	defer gglmm.CloseRedisCacher()

	gglmm.RegisterBasePath("/api/example")

	// 登录态中间件请参考gglmm-account

	gglmm.RegisterHTTPHandler(gglmm.NewRESTHTTPService(Example{}), "/example").
		Middleware(gglmm.Middleware{
			Name: "example",
			Func: middlewareFunc,
		}).
		RESTAction(gglmm.RESTAll)

	gglmm.ListenAndServe(":10000")
}

func middlewareFunc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("before")
		next.ServeHTTP(w, r)
		log.Printf("after")
		return
	})
}