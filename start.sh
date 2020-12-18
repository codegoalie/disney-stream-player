#!/bin/bash

set -e

export PKG_CONFIG_PATH=/usr/lib/x86_64-linux-gnu/pkgconfig/:/usr/share/pkgconfig/
go run main.go
