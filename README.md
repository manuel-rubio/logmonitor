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
GOPATH=$(pwd) go install .../logmonitor
```

You'll get the binary `bin/logmonitor` ready to go.

If you want to use the logs generator (a tool designed to generate specific logs) you can compile it via this command:

```
GOPATH=$(pwd) go install .../genlogs
```

The command will be available in this path `bin/genlogs`.

Infrastructure
--------------

The system is running using several goroutines. These goroutines are in carge of:

  * `tail`: keeps reading the file, line by line. Sending each line back to the main goroutine.
  * `stats`: accepts log entry lines. It generates a new goroutine to handle the timer and ensure the output is generated exactly every 10 seconds.
  * `handle break`: to handle the Ctrl+C press and exits.
  * `proxy`: accepts log entry lines. It analyzes the proxy chain to ensure it's efficient, otherwise an alert is triggered.
  * `traffic`: accepts log entry lines. It analyze the amount of lines per minute and if the number exceeds the threshold an alert is triggered. In addition when the number of log lines are again under the threshold a new alert is triggered to report the situation was solved.

The amount of memory in use for `traffic` is always the same. Only one array of 60 integers where the amount of log lines are noted.

For `proxy` the amount of memory is variable and depends on the number of different proxies there are. The main map is generated using the _Proxy_ struct and inside the _parents_ and _children_ are arrays of pointers. For a normal node with 2 parents and 2 children the amount of memory in use should be 64 bytes. For 1000 proxies in memory the average amount should be 64kB. Depending on the number of links between them, of course.

For `stats` we use an array with a 100 elements sample. The amount of memory in this case is always the same. The most of the other statistics use the same data space except the proxy hits. The map in use for proxy hits store the name of the proxy with an integer for the counter of hits and decide which is the most used proxy. We use a max of 20 bytes per proxy stored (around 20kB for 1000 different proxies).

After running during hours with `genlogs` populating the log file with hundreds of lines the memory usage was stable at 6MB (with compressed memory at 180kB aprox. in a Mac computer).
