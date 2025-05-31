#! /bin/bash

rm -rf schemas

cd scripts
go run generate.go
mv schemas ../

