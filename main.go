package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func MyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Код, выполняемый ДО обработки запроса
		fmt.Println("Before handler")
		// Вызов следующего handler в цепочке
		next.ServeHTTP(w, r)
		// Код, выполняемый ПОСЛЕ обработки запроса
		fmt.Println("After handler")
	})
}

func main() {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(MyMiddleware)

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World"))
	})
	http.ListenAndServe(":8080", router)
}
