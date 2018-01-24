#!/usr/bin/env bash

go build ./cmd/bacom || exit $?
go build ./cmd/test-server || exit $?

./test-server -version=0 &
sleep .1

FAILED=0

./bacom test -target-host=localhost:1235 -version=0.x > /dev/null
if [[ $? -ne 0 ]]; then
    echo "FAIL ./bacom test -target-host=localhost:1235 -version=0.x"
    echo "     should have 0 return code"
    FAILED=1
fi

curl http://localhost:1235/stop || exit $?
wait

./test-server -version=1 &
sleep .1

./bacom test -target-host=localhost:1235 -version="<=1.x" > /dev/null
if [[ $? -ne 0 ]]; then
    echo "FAIL ./bacom test -target-host=localhost:1235 -version=<=1.x"
    echo "     should have 0 return code"
    FAILED=1
fi

curl http://localhost:1235/stop || exit $?
wait

./test-server -version=2 &
sleep .1


./bacom test -target-host=localhost:1235 > /dev/null
if [[ $? -eq 0 ]]; then
    echo "FAIL ./bacom test -target-host=localhost:1235"
    echo "     should have non-zero return code"
    FAILED=1
fi

./bacom test -target-host=localhost:1235 -conf=bacom-tests/ignore-bar.json > /dev/null
if [[ $? -ne 0 ]]; then
    echo "FAIL ./bacom test -target-host=localhost:1235"
    echo "     should have 0 return code"
    FAILED=1
fi

curl http://localhost:1235/stop || exit $?
wait

if [[ $FAILED -ne 0 ]]; then
    exit 1
else
    echo "OK"
fi