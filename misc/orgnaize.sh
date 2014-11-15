#!/bin/sh

x=109
y=98
w=686
h=368
layout=0
gap=6

dir=$1 ; shift

imgcat -x $x -y $y -width $w -height $h -layout $layout -gap $gap \
  -output "${dir}.png" ${dir}/*.png
