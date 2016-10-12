#!/bin/bash

CJAR_DIR=.
CJAR_TMP_FILENAME=cookie_jar.tmp
EXPECTED_INSTANCES=
MAXIMUM_RETRIES=10
ACTUAL_INSTANCES=0
TRIES=0

usage() { echo "Usage: $0 -e <expected number of instances>  [-m <maximum number of tries - default 10>] [-d <cookie jar directory - default .>]" 1>&2; exit 1; }

while getopts ":e:m:d:" o; do
    case "${o}" in
        e)
            EXPECTED_INSTANCES=${OPTARG}
            ;;
        m)
            MAXIMUM_RETRIES=${OPTARG}
            ;;
        d)
            CJAR_DIR=${OPTARG}
            ;;
        *)
            usage
            ;;
    esac
done
shift $((OPTIND-1))

if [ -z "${EXPECTED_INSTANCES}" ]; then
    usage
fi

while [ $TRIES -lt $MAXIMUM_RETRIES -a $ACTUAL_INSTANCES -lt $EXPECTED_INSTANCES ]
do

    TRIES=$[$TRIES+1]
    curl -s -X POST dora.sunset.cf-app.com/session -c $CJAR_TMP_FILENAME > /dev/null || exit $?
    instance=`grep JSESSIONID cookie_jar.tmp | cut -f7`
    echo "instance >>$instance<<"
    CJAR=${CJAR_DIR}/cookie_jar_$instance.cjar


    if [ -f $CJAR ]; then
      echo "cookie jar for $instance already exists"
      rm $CJAR_TMP_FILENAME
    else
      mv $CJAR_TMP_FILENAME $CJAR
      ACTUAL_INSTANCES=$[$ACTUAL_INSTANCES+1]
    fi

done
echo "found $ACTUAL_INSTANCES in $TRIES tries"

