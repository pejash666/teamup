#!/bin/bash

serv_name="teamup"
go mod tidy
go build ./
chmod +x ./$serv_name
# shellcheck disable=SC2083
# 注意这里不能用go run 而要直接运行exe！
./$serv_name
