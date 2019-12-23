#!/usr/bin/env bash

BIN=./bitcoin-kline
STDLOG=output.log

RUNMODE=dev

if [[ "$1" ]]; then
    RUNMODE=$1
fi

echo "export RUNMODE=$RUNMODE"

export RUNMODE=$RUNMODE

chmod u+x $BIN

ID=$(/usr/sbin/pidof "$BIN")
if [[ "$ID" ]] ; then
    echo "kill -SIGINT $ID"
    kill -2 $ID
fi

while :
do
    ID=$(/usr/sbin/pidof "$BIN")
    if [[ "$ID" ]] ; then
        echo "service still running...wait"
        sleep 0.1
    else
        echo "Starting service nohup $BIN > $STDLOG 2>&1 &"

        nohup $BIN > $STDLOG 2>&1 &

        # sleep 2s
        echo "wait..."
        sleep 2

        ID=$(/usr/sbin/pidof "$BIN")
        if [[ "$ID" ]] ; then
            echo "service started $ID"
        else
            echo "service not start"
            cat $STDLOG
        fi

        break
    fi
done
