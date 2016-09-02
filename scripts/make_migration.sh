#!/bin/bash

name=$1
id=`date +%s`

touch db/migrations/${id}_${name}.go
touch db/migrations/${id}_${name}_test.go
