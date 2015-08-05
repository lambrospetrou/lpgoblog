#!/bin/bash

# check for updates in the packages and dependencies and download them
#go get -u
#go get
# build the new Spito
#go build -o goblog-t

# move the executable to the public directory while restarting supervisor
#mv goblog-t /home/lambros/public/lambrospetrou.com/public/blog/goblog
#cp -rf templates /home/lambros/public/lambrospetrou.com/public/blog/templates
#cp -rf static /home/lambros/public/lambrospetrou.com/public/blog/static

BLOG_DIR=/home/lambros/public/lambrospetrou.com/public/blog

mv goblog $BLOG_DIR/goblog
rm -rf $BLOG_DIR/templates
cp -rf templates $BLOG_DIR/templates
rm -rf $BLOG_DIR/static
cp -rf static $BLOG_DIR/static

sudo supervisorctl restart lpgoblog
