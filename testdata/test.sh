#!/usr/bin/env bash

# List signals: kill -l

#pid=

pid=
trap 'echo "Received SIGUSR1"' SIGUSR1
trap "{ echo 'Received SIGTERM'; exit 131;}" SIGTERM
#trap '{ echo 'Interrupting'; [[ $pid ]] && kill "$pid"; exit 123}' SIGINT
#trap "{ echo 'Terminating'; exit 255; }" SIGTERM

echo "Message on stdout"

>&2 echo "Message on stderr"

sh -c 'Message on stdout from subshell'

read -t 1 -r fromstdin
echo "Message received on stdin: ${fromstdin}"

"$@" & pid=$!
wait "$pid"
res=$?
pid=

echo "test.sh completed res:$res"

exit "$res"
#"$@" & pid=$!
#wait
#pid=