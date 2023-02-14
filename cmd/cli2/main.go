package main

import (
	"fmt"
	_ "fmt"
	"io"
	_ "log"
	"os"
	_ "os"
	_ "path/filepath"
	"time"
	_ "time"

	alg "github.com/AlexandrM09/DDOperation/pkg/balancingservices"
)

func main() {
	defer duration(track(os.Stdout, "App duration"))
	pool := &alg.PoolWell{}
	pool.Building("config.yaml", 25)
	fmt.Printf("cli run \n")
	pool.Run()
	fmt.Printf("programm exit \n")
}
func track(out io.Writer, msg string) (io.Writer, string, time.Time) {
	return out, msg, time.Now()
}

func duration(out io.Writer, msg string, start time.Time) {

	out.Write([]byte(fmt.Sprintf("%v: %.1fsec\n", msg, time.Since(start).Seconds())))
}
