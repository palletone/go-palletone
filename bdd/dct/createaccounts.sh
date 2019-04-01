#!/bin/bash

num=1

while [ $num -le $1 ] ;
do
	./createaccount.sh
	let ++num;
done
