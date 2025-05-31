//go:generate ./generate.sh
package k8stypes

import (
	"embed"
	"path/filepath"
)

//go:embed schemas/*
var schemas embed.FS

func GetSchema(Group, Version, Kind string) ([]byte, error) {

	path := filepath.Join("schemas", Group, Version, Kind+".json")
	bytes, err := schemas.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}
