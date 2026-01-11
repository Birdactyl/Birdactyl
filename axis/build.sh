#!/bin/bash
cd "$(dirname "$0")"
go build -o axis .
echo "Built: axis"
