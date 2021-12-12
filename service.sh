#!/bin/sh
### BEGIN INIT INFO
# Provides:          main_go
# Required-Start:    $network $syslog
# Required-Stop:     $network $syslog
# Default-Start:     2 3 4 5
# Default-Stop:      0 1 6
# Short-Description: Start go server at boot time
# Description:       description
### END INIT INFO

case "$1" in
start)
    exec ./stats &
    ;;
stop)
    kill $(sudo lsof -t -i:80)
    ;;
*)
    echo $"Usage: $0 {start|stop}"
    exit 1
    ;;
esac
exit 0
