#!/bin/sh
set -e

RUN=$1
FLAGS=$2

set +ex

check_errcheck() {
  go install github.com/kisielk/errcheck
  errcheck -ignoretests -ignoregenerated ./...
}

check_fmt() {
  gofmt -l .
}

check_imports() {
  go install golang.org/x/tools/cmd/goimports
  goimports -l .
}

check_lint() {
  go install golang.org/x/lint/golint
  golint -set_exit_status ./...
}

check_sec() {
  go install github.com/securego/gosec/v2/cmd/gosec
  gosec -quiet ${FLAGS} ./...
}

check_shadow() {
  go install golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow
  go vet -vettool=`which shadow` ${FLAGS} ./...
}

check_staticcheck() {
  go install honnef.co/go/tools/config && go install honnef.co/go/tools/cmd/staticcheck@latest
  staticcheck ${FLAGS} ./...
}

check_vet() {
  go vet ${FLAGS} ./...
}

case ${RUN} in
	"errcheck" )
		check_errcheck
		;;
	"fmt" )
		check_fmt
		;;
	"imports" )
		check_imports
		;;
	"lint" )
		check_lint
		;;
	"sec" )
		check_sec
		;;
	"shadow" )
		check_shadow
		;;
	"staticcheck" )
		check_staticcheck
		;;
	"vet" )
		check_vet
		;;
	* )
    check_imports
    check_fmt
    check_lint
    # check_sec
    # check_shadow
    check_errcheck
    check_staticcheck
    check_vet
esac
