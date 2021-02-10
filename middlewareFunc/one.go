package middlewareFunc

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo"
)

func MiddleOne(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		fmt.Println("dari middleware One")
		return next(ctx)
	}
}

func MiddleTwo(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		fmt.Println("dari middleware Two")
		return next(ctx)
	}
}

//MiddleSomething adalah gabungan echo dengan 3rd party
func MiddleSomething(next http.Handler) http.Handler {
	//sebagai catatan paramater dan return harus sesuai dengan third party
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Ini adalah 3rd party middleware")
		next.ServeHTTP(w, r)
	})
}
