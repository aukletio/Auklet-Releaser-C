#!/bin/bash
x=30000000
time ../wrapper/wrapper -n ./main-instd $x
time ../wrapper/wrapper ./main-instd $x
time ./main-instd $x
