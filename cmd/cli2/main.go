package main

import (
	_ "fmt"
	_ "log"
	_ "os"
	_ "path/filepath"
	_ "time"

	alg "github.com/AlexandrM09/DDOperation/pkg/balancingservices"
)

func main() {
	pool := &alg.PoolWell{}
	pool.Building("config.yaml", 60)
}
