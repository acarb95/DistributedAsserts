#!/bin/bash

HOSTS=5
SLEEPTIME=20

DINV=$GOPATH/src/bitbucket.org/bestchai/dinv
testDir=$GOPATH/src/github.com/acarb95/DistributedAsserts/tests/microbenchmark_tests

function shutdown {
    kill `ps | pgrep ricart | awk '{print $1}'` > /dev/null
}

function runTest {
    pids=()
    for (( i=0; i<$HOSTS; i++))
    do
        # echo $i
        go run node.go -id=$i -hosts=$HOSTS -time=$SLEEPTIME &
    done

    sleep 15
    mkdir results
    mv *.log-Log.txt results
    mv *Encoded.txt results
    # shutdown
}

function runLogMerger {
    directory=./results
    echo $directory
    cd $directory
    #merging consistant cuts
    dinv -v -l *Encoded.txt *log-Log.txt
    mkdir dinv-output
    mv *.trace dinv-output
    # TODO make dtrace files
    for trace_file in ./dinv-output/*; do
        mv "$trace_file" "./dinv-output/$(basename "$trace_file" .trace).dtrace"
    done
    #regular daikon output
    dinv -l -plan=NONE *Encoded.txt *log-Log.txt 
    mkdir daikon-output
    mv *.trace daikon-output
    for trace_file in ./daikon-output/*; do
        mv "$trace_file" "./daikon-output/$(basename "$trace_file" .trace).dtrace"
    done
    cd ..
}

function runDaikon {
    cd $testDir/results
    for directory in ./daikon*; do
        java daikon.Daikon $directory/*.dtrace
        mv *.gz $directory
        # gunzip $directory/*.gz
        java daikon.PrintInvariants $directory/*.inv.gz > $directory/daikon_output.txt
    done

    for directory in ./dinv*; do
        java daikon.Daikon $directory/*.dtrace
        mv *.gz $directory
        # gunzip $directory/*.gz
        java daikon.PrintInvariants $directory/*.inv.gz > $directory/dinv_output.txt
    done
}

function cleanup {
    cd $testDir
    rm -rf results
    # shutdown
}    

function movelogs {
    cd $testDir
    shopt -s nullglob
    set -- *[gd].txt
    if [ "$#" -gt 0 ]
    then
        name=`date "+%m-%d-%y-%s"`
        mkdir old/$name
        mv *[gdt].txt old/$name
        mv *.dtrace old/$name
        mv *.gz old/$name
    fi
}

if [ "$1" == "-c" ];
then
    cleanup
    exit
fi
runTest
runLogMerger
runDaikon
if [ "$1" == "-d" ];
then
    exit
fi



