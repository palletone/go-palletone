#!/bin/bash
branch=`git symbolic-ref --short -q HEAD`
echo $branch
