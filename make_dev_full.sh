#!/usr/bin/env bash
 
VOLUME=FULL_DISK
DEVICE=$(hdiutil attach -nomount ram://64)
diskutil erasevolume MS-DOS ${VOLUME} ${DEVICE}
touch "/Volumes/${VOLUME}/full"
cat /dev/zero > "/Volumes/${VOLUME}/space"
