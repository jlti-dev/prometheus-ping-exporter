#!/bin/sh
if [[ -z "${GATEWAY}" ]]; then
        echo "Starting without changing Gateway. Consider setting env variable \"GATEWAY\""
else
        gw="${GATEWAY}"
        echo "redirecting Default Gateway to hostname $gw"
        ip="$(getent hosts ${gw}| awk '{ print $1 }')"
        echo "found ip $ip"
        if ping -c 1 $ip &> /dev/null; then
                ip route del default
                ip route add default via $ip
                echo "redirecting all traffic, deleted default and readded"
        else
                echo "ping to gateway failed."
                exit 1
        fi


fi

echo "starting ping-exporter"
/app/main
