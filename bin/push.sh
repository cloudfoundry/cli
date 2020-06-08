#!/usr/bin/env bash
set -e

git pull -r

branch_name=$(git rev-parse --abbrev-ref HEAD)
if [[ "$branch_name" == "master" ]]; then
  printf "It looks like you are committing to master. If this is something that needs to be backported, please make the commit there and then use this script to merge it forward.\n Do you want to continue? [y/N]:"

  read input
  if [[ "$input" == "y" || "$input" == "Y" ]]; then
    git push
  fi
  exit 0
fi

git push
printf "Do you want to merge this commit forward to master? [y/N]: "

read input
if [[ "$input" == "y" || "$input" == "Y" ]]; then
  git rebase origin/master
  git checkout master
  git pull -r
  git merge $branch_name
  make clean && make build
  git push

  git checkout $branch_name
fi
