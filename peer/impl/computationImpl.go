package impl

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"time"
)

// input the cost of the computation, the computation's id and the list of nodes that participated in the computation

func (n *node) Compute(executable []byte, data []byte) ([]byte, error) {
	code := string(executable)
	timestamp := time.Now().Unix()
	filename := "executables/" + strconv.FormatInt(timestamp, 10) + ".go"
	file, err := os.Create(filename)
	defer file.Close()
	// len variable captures the length
	// of the string written to the file.
	num, err := file.WriteString(code)
	if err != nil {
		return nil, err
	}

	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	fmt.Println(path)

	app := "go"
	arg0 := "run"
	output, err := exec.Command(app, arg0, filename).Output()
	if err != nil {
		fmt.Println(":(")
		return nil, err
	}
	fmt.Print("output ", string(output))

	fmt.Println("finished ", num)
	return nil, nil
}

func (n *node) SampleFunc() {
	fmt.Println("lol")
}
