package main

import (
	"fmt"
	"github.com/AlexLex13/Infinity/internal/config"
)

func main() {
	cfg := config.MustLoad()
	fmt.Print(cfg)
}
