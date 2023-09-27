#!/bin/bash
echo "started"
cd /torpedo-gin
go run /torpedo-gin/apiServer/pxone/apiserver.go
echo "ended"