package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

func main() {
	num, _ := strconv.Atoi(os.Args[1])
	time.Sleep(time.Second)
	fmt.Print(num * 2)
}
