#!/usr/bin/env bash

export $(cat config/manager/secret/.env | xargs)

go run main.go