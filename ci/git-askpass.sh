#!/bin/sh
# Hands the GitHub credential to git without putting it in a URL or argv.
case "$1" in
  Username*) echo "$GIT_USER" ;;
  *) echo "$GIT_TOKEN" ;;
esac
