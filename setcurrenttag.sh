#!/usr/bin/env bash

for current_tag in $(git tag --sort=-creatordate)
do

if [ "$current_tag" != 0 ];then
    export AUTHY_CURRENT_TAG=$current_tag
    break
fi
done

