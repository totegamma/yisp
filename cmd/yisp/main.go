package main

import (
	"fmt"
	"github.com/totegamma/yisp"
)

func main() {

	env := yisp.NewEnv()

	evaluated := yisp.EvaluateYisp("./testdata/include.yaml", env)
	result, err := yisp.Render(evaluated)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println(result)

}
