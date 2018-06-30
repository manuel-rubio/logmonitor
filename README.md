logmonitor
==========

`logmonitor` is a foreground process in charge to read a file passed as parameter and provide a monitorization output for stats and alerts.

The syntax of the command is:

```
logmonitor <access.log> <traffic_threshold>
```

where *access.log* is the path to a file that is being **actively** written-to and
*traffic_threshold* is an integer parameter described in the requirements below.
The task of this tool is to monitor the specified log file for HTTP traffic with
log lines like these:

```
194.179.0.18, 10.16.1.2, 10.129.0.10 [13/02/2016 16:45:01] "GET /some/path?param1=x&param2=y HTTP/1.1" 200 0.006065388
213.1.20.7 [13/02/2016 16:45:02] "POST /some/other/path HTTP/1.0" 201 0.012901348
```

Conformant log lines contain US ASCII characters only, and the fields are as
follows, from left to right, always separated by one or more space or
non-printable characters:
  * IPv4 addresses, from left to right, are the client's IP address, followed
    by optional additional addresses with the list of proxies forwarding the
    request to each other until the last one reaches us. The whole list of
    addresses, including the origin, are referred to as the proxy chain.
    Addresses are separated by commas (`,`) and any amount of space or
    non-printable characters.
  * The date and the time of the request in UTC as recorded by the server that
    processed it, enclosed in brackets, and separated by a single space
    character.
  * The request information enclosed in double quotes, with the method, the path
    including any URL parameters, and the protocol and its version, each
    separated by one space character.
  * The status code of the response.
  * The seconds it took the server to process and respond to the request.

The JSON output contains different fields. These fields are:

  * `timestamp`: a UNIX timestamp filled when the message is output as an integer.
  * `message_type`: the type of the message. It could be: `stats` or `alert`.
  * `get` (stats): The number of GET hits.
  * `post` (stats): The number of POST hits.
  * `hits` (stats): The number of GET and POST hits.
  * `forwarded_hits` (stats): The number of requests that have been proxied.
  * `most_used_proxy` (stats): The proxy that forwarded most requests.
  * `most_used_proxy_hits` (stats): The number of forwarded requests for the proxy
    that forwarded most requests.
  * `p95` (stats): The 95 percentile request time.
  * `bad_lines` (stats): The number of non-conformant log lines.
  * `alert_type` (alert): A string specifying the type of alert. The valid values
    are `traffic_above_threshold` or `traffic_below_threshold` depending
    on whether the cross happened upwards or downwards of the threshold,
    respectively.
  * `period` (alert): The relative period to which the alert applies. This will
    be always `minute`.
  * `threshold` (alert): The *traffic_threshold* value given as argument to our
    monitor.
  * `current_value` (alert): The amount of requests performed in the past minute.
  * `proxy_chain` (alert): The list of addresses in the proxy chain.
  * `inefficient_addresses` (alert): The list of addresses with more proxies than
     needed.
  * `efficient_proxy_chains` (alert): The list of efficient proxy chains.

Getting started
---------------

First, the requierements. You need:

  * Go 1.9.2 or later

To compile you can use the following command in the root directory (where this file is):

```
GOPATH=$(pwd) go install github.com/manuel-rubio/logmonitor
```

You'll get the binary `bin/logmonitor` ready to go.

Infrastructure
--------------

The system is running using several goroutines. These goroutines are in carge of:

  * `tail`: keeps reading the file, line by line. Sending each line back to the main goroutine.
  * `stats`: accepts log entry lines. It generates a new goroutine to handle the timer and ensure the output is generated exactly every 10 seconds.
  * `handle break`: to handle the Ctrl+C press and exits.
