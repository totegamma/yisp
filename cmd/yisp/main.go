package main

import (
	"fmt"
	"github.com/totegamma/yisp"
)

func main() {

	result, err := yisp.EvaluateYisp("./testdata/test_template.yaml")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println(result)

}
