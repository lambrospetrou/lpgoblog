#!/bin/bash

# check for updates in the packages and dependencies and download them
go get -u
# build the new Spito
go build -o goblog-t

# move the executable to the public directory while restarting supervisor
mv goblog-t /home/lambros/public/lambrospetrou.com/public/blog/goblog
sudo supervisorctl restart lpgoblog
