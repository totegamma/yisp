package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/pkg/errors"
	"github.com/totegamma/yisp"
	"github.com/totegamma/yisp/yaml"
)

func main() {
	data := `
!yisp
&mkpod
- lambda
- [!string name, !string image]
- apiVersion: v1
  kind: Pod
  metadata:
    name: *name
  spec:
    containers:
      - name: *name
        image: *image
---
!yisp
- *mkpod
- mypod1
- myimage1
---
message: this is a normal yaml document
fruits:
  - apple
  - banana
  - chocolate
`

	decoder := yaml.NewDecoder(strings.NewReader(data))
	if decoder == nil {
		panic("failed to create decoder")
	}
	for {
		var root yaml.Node
		err := decoder.Decode(&root)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			panic(err)
		}

		env := yisp.NewEnvironment()
		parsed, err := yisp.Parse(&root, env)
		if err != nil {
			panic(err)
		}

		evaluated, err := yisp.Eval(parsed, yisp.GetGlobals())
		if err != nil {
			panic(err)
		}

		result := yisp.Render(evaluated)

		yisp.YamlPrint(result)
		fmt.Println("---")
	}
}
