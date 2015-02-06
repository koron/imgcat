#!/bin/sh

x=330
y=102
w=228
h=192
layout=1
wrap=3
gap=6

dir=${1%/} ; shift

imgcat -x $x -y $y -width $w -height $h -layout $layout -wrap $wrap -gap $gap \
  -output "${dir}.png" ${dir}/*.png
