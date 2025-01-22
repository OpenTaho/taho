#!/bin/bash

main() {
  rm -f "$HOME/go/bin/taho"
  go install .

  if [ -d tests ]; then
    cd tests || exit 1
  fi

  test 0 3ca25d984982495dbdcaecacdeabfc59ce5342d2
  test 1 91268ce3a9f4edb082291f5283ecce8fdf6c71d0
  test 2 44373f9b0b1095d02c744f1a9b9e7e7f8995bdb4
  test 3 6422ff9a0533e512f26a33241ec8366266b307bb
  test 4 8525389041f324edd6850012b7c5f640f2a04906
  test 5 88c1a8f8f8f6a38562f0af0f97690c524e8dd533
  test 6 5623700e77b64f571fc189e02124f06a3d874b91
  test 7 76d6aa2652bfadf456872703bc5c7603f2ca5f0c
  test 8 d7e88f5dbc7a89081fa93db51dbc8cbd2a8a30a8
}

test() {
  expected="$2"
  (
    rm -rf "$1-result"
    cp -r "$1" "$1-result"
    cd "$1-result" || exit 1

    taho > /dev/null

    result1="$(
      (
        sha1sum ./*.tf
        sha1sum ./*.md
      ) |
      sha1sum |
      sed 's/ .*//'
    )"

    if [[ "$expected" != "$result1" ]]; then
      >&2 echo "Test $1 failed; got $result1"
      exit 1
    fi

    result2="$(
      (
        sha1sum ./*.tf
        sha1sum ./*.md
      ) |
      sha1sum |
      sed 's/ .*//'
    )"

    if [[ "$result1" != "$result2" ]]; then
      >&2 echo "Test $1 failed because result1 != result2"
      exit 1
    fi
  ) || exit 1

  rm -rf test
}

main