#!/bin/bash

# script to generate an icon file from a single .png image
# feed a 1024x1024 .png file to this script like so:
# ./make_icons.sh image_file.png
#
# this script only works on a darwin (os x) machine
# source:
# https://stackoverflow.com/a/20703594


expectHeight="pixelWidth: 1024"

size=$(sips -g pixelWidth qri_desktop_icon_1024x1024.png | grep pixelWidth)

if [ expect != size ];
then echo "creating icons";
else echo ".png is the wrong size:
${size}" && exit;
fi

mkdir icon.iconset
sips -z 16 16     $1 --out icon.iconset/icon_16x16.png
sips -z 32 32     $1 --out icon.iconset/icon_16x16@2x.png
sips -z 32 32     $1 --out icon.iconset/icon_32x32.png
sips -z 64 64     $1 --out icon.iconset/icon_32x32@2x.png
sips -z 128 128   $1 --out icon.iconset/icon_128x128.png
sips -z 256 256   $1 --out icon.iconset/icon_128x128@2x.png
sips -z 256 256   $1 --out icon.iconset/icon_256x256.png
sips -z 512 512   $1 --out icon.iconset/icon_256x256@2x.png
sips -z 512 512   $1 --out icon.iconset/icon_512x512.png
cp $1 icon.iconset/icon_512x512@2x.png
iconutil -c icns icon.iconset
rm -R icon.iconset

mkdir icons
sips -z 16 16     $1 --out icons/16x16.png
sips -z 24 24     $1 --out icons/24x24.png
sips -z 32 32     $1 --out icons/32x32.png
sips -z 48 48     $1 --out icons/48x48.png
sips -z 64 64     $1 --out icons/64x64.png
sips -z 96 96     $1 --out icons/96x96.png
sips -z 128 128   $1 --out icons/128x128.png
sips -z 256 256   $1 --out icons/256x256.png
sips -z 512 512   $1 --out icons/512x512.png
cp $1 icons/1024x1024.png