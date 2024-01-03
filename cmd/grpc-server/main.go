package main

import (
	"fmt"
	"gRPC-server/iternal/config"
)

func main() {
	// todo: инициализировать обьект конфига
	cfg := config.MustLoad()
	fmt.Println(cfg)
	// todo: инициализировать логгер
	// todo: инициализировать приложение (app)
	// todo: заупск
}
