#!/bin/sh

sudo adduser www --disabled-password
sudo su www -c "/usr/local/go/bin/go get github.com/gobwas/ws"
sudo su www -c "/usr/local/go/bin/go get github.com/gorilla/websocket"
sudo su www -c "/usr/local/go/bin/go get github.com/gorilla/websocket"
sudo su www -c "/usr/local/go/bin/go get github.com/json-iterator/go"
sudo su www -c "/usr/local/go/bin/go get github.com/valyala/fasttemplate"
sudo su www -c "mkdir -p /home/www/go/src/github.com/cchan/"
sudo su www -c "ln -s $PWD /home/www/go/src/github.com/cchan/syncdoc"
sudo cp syncdoc.clive.io.service /etc/systemd/system/
sudo service syncdoc.clive.io start
sudo systemctl enable syncdoc.clive.io.service
