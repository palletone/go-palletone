#!/bin/bash

rm ./node/gptn
rm ./node/*/gptn -rf
tar czvf ./logs/$1.tar.gz ./node