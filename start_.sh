#!/bin/sh

# start the spitty executable
./lpgoblog 2> /home/lambros/public/lambrospetrou.com/log/stderr_lpgoblog.log 1> /home/lambros/public/lambrospetrou.com/log/stdout_lpgoblog.log &

# log its PID for easy temrination
echo $! > RUNNING_PID

