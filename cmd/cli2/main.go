package main

import (
	"fmt"
	"io"
	_ "log"
	"os"

	//  "path/filepath"
	"time"

	alg "github.com/AlexandrM09/DDOperation/pkg/balancingservices"
)

func main() {
	defer duration(track(os.Stdout, "App duration"))
	pool := &alg.PoolWell{}
	_ = pool.Building("config.yaml", 10)
	fmt.Printf("cli run \n")
	_ = pool.Run()
	fmt.Printf("programm exit \n")
}
func track(out io.Writer, msg string) (io.Writer, string, time.Time) {
	return out, msg, time.Now()
}

func duration(out io.Writer, msg string, start time.Time) {

	_, _ = out.Write([]byte(fmt.Sprintf("%v: %.1fsec\n", msg, time.Since(start).Seconds())))
}
