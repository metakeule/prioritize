#!/bin/bash
go build -o ./prioritize;
zgok build -e prioritize -z static -o prioritize_all;
chmod +x prioritize_all;