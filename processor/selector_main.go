package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/treaster/shire/examples/calc/interpreter"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	ctx := interpreter.Context{}
	for {
		fmt.Printf("> ")

		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("input error:", err.Error())
			continue
		}

		expr, err := interpreter.Parse(input)
		if err != nil {
			fmt.Println("parse error:", err.Error())
			continue
		}

		err = expr.Exec(ctx)
		if err != nil {
			fmt.Println("eval error:", err.Error())
			continue
		}
	}
}
