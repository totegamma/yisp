package core

import "testing"

func TestLookupYispNodeByPath(t *testing.T) {
	root := NewYispMap()
	root.Set("app", testMapNode(map[string]*YispNode{
		"metadata": testMapNode(map[string]*YispNode{
			"name": testStringNode("yisp"),
		}),
	}))

	got, ok := LookupYispNodeByPath(root, "app.metadata.name")
	if !ok {
		t.Fatal("expected app.metadata.name to resolve")
	}
	if got.Kind != KindString || got.Value != "yisp" {
		t.Fatalf("unexpected value: %#v", got)
	}
}

func TestLookupYispNodeByPathOptionalMissing(t *testing.T) {
	root := NewYispMap()
	root.Set("app", testMapNode(map[string]*YispNode{
		"metadata": testMapNode(map[string]*YispNode{
			"name": testStringNode("yisp"),
		}),
	}))

	got, ok := LookupYispNodeByPath(root, "app.metadata.labels?")
	if !ok {
		t.Fatal("expected optional missing segment to resolve")
	}
	if got.Kind != KindNull || got.Value != nil {
		t.Fatalf("unexpected value: %#v", got)
	}
}

func TestEnvGetDotPath(t *testing.T) {
	env := NewEnv()
	env.Set("props", testMapNode(map[string]*YispNode{
		"name": testStringNode("yisp"),
	}))

	got, ok := env.Get("props?.name")
	if !ok {
		t.Fatal("expected props?.name to resolve")
	}
	if got.Kind != KindString || got.Value != "yisp" {
		t.Fatalf("unexpected value: %#v", got)
	}
}

func testMapNode(items map[string]*YispNode) *YispNode {
	m := NewYispMap()
	for key, item := range items {
		m.Set(key, item)
	}
	return &YispNode{
		Kind:  KindMap,
		Value: m,
	}
}

func testStringNode(value string) *YispNode {
	return &YispNode{
		Kind:  KindString,
		Value: value,
	}
}
