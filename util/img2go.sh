#!/bin/sh
#
# Script to convert an '.img' file to go source code
#
# Usage:
#    img2go.sh filename [ package-name ]
#
# Reads filename.img and creates filename_img.go. If the package-name is not specified, the current
# working directory name is used instead.
if [ $# -ne 1 -a $# -ne 2 ]; then
  echo "Usage: $0 filename [ package-name ]"
  exit 1
fi
pkg=`basename ${PWD}`
if [ $# -eq 2 ]; then
  pkg=$2
fi
echo "Package will be $pkg"
out="${1}_img.go"
in="${1}.img"
cat <<EOF > ${out}
// This file is generated from $in by img2go.sh
package $pkg

var $1_img = []uint32{
EOF
for w in `cat ${in}`; do
  echo "\t0x$w," >> ${out}
done
echo "}" >> ${out}
