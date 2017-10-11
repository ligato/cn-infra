#!/usr/bin/env bash
#
TMP_FILE="/tmp/out"
TMP_FILE2="/tmp/unprocessed"
processedLines=0 
exitCode=0
PREV_IFS="$IFS"

# test whether output of the command contains expected lines
# arguments
# 1-st command to run
# 2-nd array of expected strings in the command output
# 3-rd argument is an optional array of unexpected strings in the command output
# 4-th argument is an optional command runtime limit (value 0 does not cause stopping of the run.)
function testOutput {
IFS="${PREV_IFS}"

    #run the command
    if [ $# -ge 4 ]; then
        if [ -e ${TMP_FILE} ]; then
            # if exists file /tmp/out we assume that the command still runs, do not start it again
            echo > /dev/null
        else
            $1 > ${TMP_FILE} 2>&1 &
            CMD_PID=$!
        fi
        # do we set runtime limit?
        if [ $4 -ne 0 ]; then
            sleep $4
            kill $CMD_PID
        else
            # The command continues to run - his PID is in the variable CMD_PID
            sleep 20
        fi
        cat ${TMP_FILE} | awk "NR > $processedLines" > ${TMP_FILE2}
    else
        $1 > ${TMP_FILE} 2>&1
        cat ${TMP_FILE} > ${TMP_FILE2}
    fi

IFS="
"
    echo "Testing $1"
    rv=0
    # loop through expected lines
    for i in $2
    do
        if grep -- "${i}" ${TMP_FILE2} > /dev/null ; then
            echo "OK - '$i'"
        else
            echo "Not found - '$i'"
            rv=1
        fi
    done
    # loop through unexpected lines
    if [[ ! -z $3 ]] ; then
        for i in $3
        do
            if grep -- "${i}" ${TMP_FILE2} > /dev/null ; then
                echo "IS NOT OK - '$i'"
                rv=1
            fi
        done
    fi

    # if an error occurred print the output
    if [[ ! $rv -eq 0 ]] ; then
        cat ${TMP_FILE2}
        exitCode=1
    fi

    echo "================================================================"
    if [ $# -ge 4 ]; then
        if [ $4 -ne 0 ]; then
            rm ${TMP_FILE}
            rm ${TMP_FILE2}
        else
            # The command continues to run - the output is still redirected to the file ${TMP_FILE}
            # read -n1 -r -p "Press any key to continue..." key
            processedLines=`wc -l ${TMP_FILE} | cut --delimiter=" " -f1`
        fi
    else
        rm ${TMP_FILE}
        rm ${TMP_FILE2}
    fi
    return ${rv}
}

source scripts/docker_start_stop_functions.sh
echo "som tu"
