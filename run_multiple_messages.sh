#! /usr/bin/bash

commands=(
    'echo "{\"Type\":0,\"Data\":{\"Ip\":\"127.0.0.2\",\"Port\":8080,\"ConnsCap\":1}}" > /dev/tcp/127.0.0.1/8080'
    'echo "{\"Type\":0,\"Data\":{\"Ip\":\"127.0.0.2\",\"Port\":8080,\"ConnsCap\":1}}" > /dev/tcp/127.0.0.1/8080'
    'echo "{\"Type\":0,\"Data\":{\"Ip\":\"127.0.0.2\",\"Port\":8080,\"ConnsCap\":1}}" > /dev/tcp/127.0.0.1/8080'
    'echo "{\"Type\":0,\"Data\":{\"Ip\":\"127.0.0.2\",\"Port\":8080,\"ConnsCap\":1}}" > /dev/tcp/127.0.0.1/8080'
    'echo "{\"Type\":0,\"Data\":{\"Ip\":\"127.0.0.2\",\"Port\":8080,\"ConnsCap\":1}}" > /dev/tcp/127.0.0.1/8080'
    'echo "{\"Type\":0,\"Data\":{\"Ip\":\"127.0.0.2\",\"Port\":8080,\"ConnsCap\":1}}" > /dev/tcp/127.0.0.1/8080'
    'echo "{\"Type\":0,\"Data\":{\"Ip\":\"127.0.0.2\",\"Port\":8080,\"ConnsCap\":1}}" > /dev/tcp/127.0.0.1/8080'
    'echo "{\"Type\":0,\"Data\":{\"Ip\":\"127.0.0.2\",\"Port\":8080,\"ConnsCap\":1}}" > /dev/tcp/127.0.0.1/8080'
    'echo "{\"Type\":0,\"Data\":{\"Ip\":\"127.0.0.2\",\"Port\":8080,\"ConnsCap\":1}}" > /dev/tcp/127.0.0.1/8080'
    'echo "{\"Type\":0,\"Data\":{\"Ip\":\"127.0.0.2\",\"Port\":8080,\"ConnsCap\":1}}" > /dev/tcp/127.0.0.1/8080'
    'echo "{\"Type\":0,\"Data\":{\"Ip\":\"127.0.0.2\",\"Port\":8080,\"ConnsCap\":1}}" > /dev/tcp/127.0.0.1/8080'
    'echo "{\"Type\":0,\"Data\":{\"Ip\":\"127.0.0.2\",\"Port\":8080,\"ConnsCap\":1}}" > /dev/tcp/127.0.0.1/8080'

)
parallel --jobs 12 ::: "${commands[@]}"

echo "All messages sent"