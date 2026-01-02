package main

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type AuthRequest struct {
	Key string `json:"key"`
}

func TestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Код, выполняемый ДО обработки запроса
		fmt.Println("Before handler")
		// Вызов следующего handler в цепочке
		next.ServeHTTP(w, r)
		// Код, выполняемый ПОСЛЕ обработки запроса
		fmt.Println("After handler")
	})
}

func LoggerMiddleware(log *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		log.Info("logger middleware enabled") //выведится 1 раз при старте

		fn := func(w http.ResponseWriter, r *http.Request) {
			entry := log.With(
				slog.String("method", r.Method),
				slog.String("request_id", middleware.GetReqID(r.Context())),
				slog.String("real_ip", r.RemoteAddr), //без router.Use(middleware.RealIP) ip возможно будет не правильный
			)

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor) //для отслеживания статуса и размера
			start := time.Now()

			next.ServeHTTP(ww, r) //переход к обработке запроса

			entry.Info("request completed",
				slog.Int("status", ww.Status()),
				slog.Int("bytes", ww.BytesWritten()),
				slog.String("duration", time.Since(start).String()),
			)
		}

		return http.HandlerFunc(fn)
	}
}
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const validKey = "12h1fn43u1@!bu"
		// 1. Читаем тело
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			render.Status(r, http.StatusUnauthorized)
			fmt.Println(err)
			return
		}
		// 2. Восстанавливаем тело
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		// 3. Декодируем JSON
		var authReq AuthRequest
		if err := render.DecodeJSON(bytes.NewBuffer(bodyBytes), &authReq); err != nil {
			render.Status(r, http.StatusUnauthorized)
			fmt.Println(err)
			return
		}
		// 4. Проверяем ключ
		if authReq.Key != validKey {
			render.Status(r, http.StatusUnauthorized)
			render.JSON(w, r, "Invalid API key")
			return
		}
		next.ServeHTTP(w, r)
	})
}
func Handler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello World"))
	time.Sleep(time.Second)
}

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	router := chi.NewRouter()

	router.Use(TestMiddleware)
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(LoggerMiddleware(log))
	router.Use(AuthMiddleware)

	router.Get("/", Handler)
	http.ListenAndServe(":8080", router)
}
