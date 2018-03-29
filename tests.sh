#!/usr/bin/env bash

go build ./cmd/bacom || exit $?
go build ./cmd/test-server || exit $?
go build ./cmd/preprocess-example || exit $?

./test-server -version=0 &
sleep .1

FAILED=0
preprocess='./preprocess-example -remove-header=X-Version -set-header="Authorization: Bearer foo"'

./bacom test -target-host=localhost:1235 -version=0.x -target-preprocess="$preprocess" > /dev/null
if [[ $? -ne 0 ]]; then
    echo "FAIL ./bacom test -target-host=localhost:1235 -version=0.x"
    echo "     should have 0 return code"
    FAILED=1
fi

curl http://localhost:1235/stop || exit $?
wait

./test-server -version=1 &
sleep .1

./bacom test -target-host=localhost:1235 -version="<=1.x" -target-preprocess="$preprocess" > /dev/null
if [[ $? -ne 0 ]]; then
    echo "FAIL ./bacom test -target-host=localhost:1235 -version=<=1.x"
    echo "     should have 0 return code"
    FAILED=1
fi

curl http://localhost:1235/stop || exit $?
wait

./test-server -version=2 &
sleep .1


./bacom test -target-host=localhost:1235 -target-preprocess="$preprocess" > /dev/null
if [[ $? -eq 0 ]]; then
    echo "FAIL ./bacom test -target-host=localhost:1235"
    echo "     should have non-zero return code"
    FAILED=1
fi

./bacom test -target-host=localhost:1235 -conf=bacom-tests/ignore-bar.json -target-preprocess="$preprocess" > /dev/null
if [[ $? -ne 0 ]]; then
    echo "FAIL ./bacom test -target-host=localhost:1235 -conf=bacom-tests/ignore-bar.json"
    echo "     should have 0 return code"
    FAILED=1
fi

./bacom test -target-host=localhost:1235 -conf=bacom-tests/ignore-bar.yaml -target-preprocess="$preprocess" > /dev/null
if [[ $? -ne 0 ]]; then
    echo "FAIL ./bacom test -target-host=localhost:1235 -conf=bacom-tests/ignore-bar.yaml"
    echo "     should have 0 return code"
    FAILED=1
fi

./bacom test -target-host=localhost:1235 -conf=bacom-tests/ignore-bar.toml -target-preprocess="$preprocess" > /dev/null
if [[ $? -ne 0 ]]; then
    echo "FAIL ./bacom test -target-host=localhost:1235 -conf=bacom-tests/ignore-bar.toml"
    echo "     should have 0 return code"
    FAILED=1
fi

curl http://localhost:1235/stop || exit $?
wait

##
# Testing JSON streams
##

./test-server -stream -version=0 &
sleep .1

FAILED=0
preprocess='./preprocess-example -remove-header=X-Version -set-header="Authorization: Bearer foo"'

./bacom test -dir=./bacom-tests-stream -target-host=localhost:1235 -version=0.x -target-preprocess="$preprocess" > /dev/null
if [[ $? -ne 0 ]]; then
    echo "FAIL ./bacom test -dir=./bacom-tests-stream -target-host=localhost:1235 -version=0.x"
    echo "     should have 0 return code"
    FAILED=1
fi

curl http://localhost:1235/stop || exit $?
wait

./test-server -stream -version=1 &
sleep .1

./bacom test -dir=./bacom-tests-stream -target-host=localhost:1235 -version="<=1.x" -target-preprocess="$preprocess" > /dev/null
if [[ $? -ne 0 ]]; then
    echo "FAIL ./bacom test -dir=./bacom-tests-stream -target-host=localhost:1235 -version=<=1.x"
    echo "     should have 0 return code"
    FAILED=1
fi

curl http://localhost:1235/stop || exit $?
wait

./test-server -stream -version=2 &
sleep .1


./bacom test -dir=./bacom-tests-stream -target-host=localhost:1235 -target-preprocess="$preprocess" > /dev/null
if [[ $? -eq 0 ]]; then
    echo "FAIL ./bacom test -dir=./bacom-tests-stream -target-host=localhost:1235"
    echo "     should have non-zero return code"
    FAILED=1
fi

./bacom test -dir=./bacom-tests-stream -target-host=localhost:1235 -conf=bacom-tests-stream/ignore-bar.json -target-preprocess="$preprocess" > /dev/null
if [[ $? -ne 0 ]]; then
    echo "FAIL ./bacom test -dir=./bacom-tests-stream -target-host=localhost:1235 -conf=bacom-tests-stream/ignore-bar.json"
    echo "     should have 0 return code"
    FAILED=1
fi

./bacom test -dir=./bacom-tests-stream -target-host=localhost:1235 -conf=bacom-tests-stream/ignore-bar.yaml -target-preprocess="$preprocess" > /dev/null
if [[ $? -ne 0 ]]; then
    echo "FAIL ./bacom test -dir=./bacom-tests-stream -target-host=localhost:1235 -conf=bacom-tests-stream/ignore-bar.yaml"
    echo "     should have 0 return code"
    FAILED=1
fi

./bacom test -dir=./bacom-tests-stream -target-host=localhost:1235 -conf=bacom-tests-stream/ignore-bar.toml -target-preprocess="$preprocess" > /dev/null
if [[ $? -ne 0 ]]; then
    echo "FAIL ./bacom test -dir=./bacom-tests-stream -target-host=localhost:1235 -conf=bacom-tests-stream/ignore-bar.toml"
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

if [[ $FAILED -ne 0 ]]; then
    exit 1
else
    echo "OK"
fi
