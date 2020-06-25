# prometheus-ping-exporter
Ping Exporter for Prometheus written in Go
## Motivation
My motivation to implement this Exporter was, that i needed to ping some public accessable hosts, but also some hosts, that are not publicly accessible, but only via a client connected to a VPN.

## How it works
The docker-compose file shows both usages.
- With a preconfigured gateway (accessing hosts in a VPN)
- No preconfigured gateway (accessing public available hosts)

To add a Gateway other than the default assigned by docker, you need to add the environment variable "GATEWAY".

For each host to Ping, there are 3 environment variables:
- IP_{{counter}} = The IP-Adress to ping. Will be labeled in Prometheus as "host"
- NAME_{{counter}} = The Name belonging to this server. You can use any kind of string, i recommend not to use spaces. This will be labeled as "name". If no name is set, it will be set to IP_{{counter}}
- GROUP_{{counter}} = The group the server belongs to. If no group is set, the group is set to "default"

I had the problem, that i had many hosts belonging to the same subnet, and i wanted to ping all of them, but aggregate them in Prometheus.

## Metrics
The Exporter basically has 7 metrics, it exposes.
- Packets = all send packets
- Success = those packets, who got a response
- Fail = those packets missing a response
- Last = the last round trip time 
- Avg = the average round trip time
- high = the highest round trip time
- low = the lowest round trip time

These metrics are divided by "Total" meaning every metric since the start of the exporter and "Scrape" meaning the metric since the last scrape request.
If you scrape this targets every 15 seconds, it will basically mean, that packets equals 15, but it can be 16 or 14, depending on network and host consumptions.
The average is calculated as (avg * counter + last ) / (counter + 1). So a missing ping will not affect the average of pings.

##Contribution
If you like to contribute, feel free to open an issue or send a pull request.
