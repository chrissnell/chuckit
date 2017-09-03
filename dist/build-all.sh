#!/bin/bash

TAG=`git describe --tag --abbrev=0`

platforms=( darwin linux freebsd windows openbsd )

PROG=chuckit
PROG_WITH_TAG=${PROG}-${TAG}
BINDIR=./dist/binaries/

echo "--> Building ${prog}"
for plat in "${platforms[@]}"; do
  echo "----> Building for ${plat}/amd64"
  if [ "$plat" = "windows" ]; then
    GOOS=$plat GOARCH=amd64 go build -o ${PROG_WITH_TAG}-win64.exe
    echo "Compressing..."
    zip -9 ${PROG_WITH_TAG}-win64.zip ${PROG_WITH_TAG}-win64.exe
    mv ${PROG_WITH_TAG}-win64.zip $BINDIR
    rm ${PROG_WITH_TAG}-win64.exe
  else
    OUT="${PROG_WITH_TAG}-${plat}-amd64"
    GOOS=$plat GOARCH=amd64 go build -o $OUT
    echo "Compressing..."
    gzip -f $OUT
    mv ${OUT}.gz $BINDIR
  fi
done

# Build Linux/ARM
echo "----> Building for linux/arm"
OUT="${PROG_WITH_TAG}-linux-arm"
GOOS=linux GOARCH=arm go build -o $OUT
echo "Compressing..."
gzip -f $OUT
mv ${OUT}.gz $BINDIR
cd ..
