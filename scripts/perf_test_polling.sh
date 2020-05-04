#!/bin/bash

config=~/.cf/config.json
dora_path=~/workspace/cf-acceptance-tests/assets/dora
csv=/tmp/csv_test.csv
temp=/tmp/temp.txt
row=""

function reset () {
    rm $1
    touch $1
    chmod +w $1
}

function setup () {
    TIMEFORMAT=%R

    reset $csv
    reset $temp
}

function change_value () {
    value=$(( 10000000 * $1 ))
    sed -i '' -E "s/\"$2\": [0-9]+/\"$2\": $value/g" $config
}

function get_time_for_command() {
    echo $1
    (time $1)2>$temp
    time=$(cat $temp | grep -E [0-9]+.[0-9]+ | tail -1)
    echo $time,
}

setup

for i in $(seq 100 10 110); do :
    for j in $(seq 100 25 125); do :
        change_value $j JobPollingInterval
        #(time cf push dora -p $dora_path)2>$temp
        #time=$(cat $temp | grep -E [0-9]+.[0-9]+ | tail -1)
        #row+=$time,
        row+=$(get_time_for_command "cf push dora -p $dora_path")
    done
    echo $row >> $csv
    change_value $i JobPollingBackoffFactor
    row=""
done

