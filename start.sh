#!/bin/sh
sudo su www -c "PORT=3001 nohup /usr/local/go/bin/go run /home/www/go/src/github.com/cchan/syncdoc/server/server.go &"
