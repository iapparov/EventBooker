// @title           EventBooker API
// @version         1.0
// @description     API для сервиса бронирования
// @BasePath        /

package main

import (
	"log"

	"eventbooker/internal/app"
	"eventbooker/internal/config"

	wbzlog "github.com/wb-go/wbf/zlog"
)

func main() {
	wbzlog.Init()

	cfg := config.MustLoad()

	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("failed to initialize application: %v", err)
	}

	application.Run()
}
