# traceroute Input Plugin

The traceroute plugin provides routing information given end host.

### Configuration:

```toml
# NOTE: this plugin forks the traceroute command. You may need to set capabilities
# via setcap cap_net_raw+p /bin/traceroute
[[inputs.traceroute]]
  ## List of urls to traceroute
  urls = ["www.google.com"] # required
  ## per-traceroute timeout, in s. 0 == no timeout
  # response_timeout = 0.0
  ## wait time per probe in seconds (traceroute -w <WAITTIME>)
  # waittime = 5.0
  ## starting TTL of packet (traceroute -f <FIRST_TTL>)
  # first_ttl = 1
  ## maximum number of hops (hence TTL) traceroute will probe (traceroute -m <MAX_TTL>)
  # max_ttl = 30
  ## source interface/address to traceroute from (traceroute -i <INTERFACE/SRC_ADDR>)
  # interface = ""
```

### Metrics:

- traceroute
  - tags:
    - target_fqdn 
    - target_ip (IPv4 string)
  - fields:
    - result_code
        - 0:success
      	- 1:no such host
    - number_of_hops (int, # of hops made)
    - field2 (float, percent)

- traceroute_hop_data
  - tags:
    - target_fqdn
    - target_ip (IPv4 string)
    - column_number (zero-indexed value representing which column of the traceroute output the data resides in)
  - fields:
    - hop_number
    - hop_fqdn
    - hop_ip (IPv4 string)
    - hop_rtt_ms (round trip time in ms)

### Sample Queries:

Get traceroute information given host
```
SELECT *
FROM "traceroute"
WHERE "target_fqdn"='www.google.com'
```

Get average round trip team for each top given time
```
SELECT MEAN("hop_rtt_ms")
FROM "traceroute_hop_data"
WHERE "time"=1453831884664956455
GROUP BY "hop_number"
```

### Example Output:

#### traceroute
```
> traceroute,host=m1.cloudpbx.ca,target_fqdn=www.google.com,target_ip=172.217.0.100 number_of_hops=6i 1525474707000000000
```

#### traceroute_hop_data
```
> traceroute_hop_data,column_number=0,host=m1.cloudpbx.ca,target_fqdn=www.google.com,target_ip=172.217.0.100 hop_fqdn="165.227.32.253",hop_ip="165.227.32.253",hop_number=1i,hop_rtt_ms=3.5250000953674316 1525474707000000000
> traceroute_hop_data,column_number=1,host=m1.cloudpbx.ca,target_fqdn=www.google.com,target_ip=172.217.0.100 hop_fqdn="165.227.32.254",hop_ip="165.227.32.254",hop_number=1i,hop_rtt_ms=3.071000099182129 1525474707000000000
> traceroute_hop_data,column_number=2,host=m1.cloudpbx.ca,target_fqdn=www.google.com,target_ip=172.217.0.100 hop_fqdn="165.227.32.253",hop_ip="165.227.32.253",hop_number=1i,hop_rtt_ms=3.4200000762939453 1525474707000000000
> traceroute_hop_data,column_number=0,host=m1.cloudpbx.ca,target_fqdn=www.google.com,target_ip=172.217.0.100 hop_fqdn="138.197.249.78",hop_ip="138.197.249.78",hop_number=2i,hop_rtt_ms=3.4010000228881836 1525474707000000000
> traceroute_hop_data,column_number=1,host=m1.cloudpbx.ca,target_fqdn=www.google.com,target_ip=172.217.0.100 hop_fqdn="138.197.249.82",hop_ip="138.197.249.82",hop_number=2i,hop_rtt_ms=3.5429999828338623 1525474707000000000
> traceroute_hop_data,column_number=2,host=m1.cloudpbx.ca,target_fqdn=www.google.com,target_ip=172.217.0.100 hop_fqdn="138.197.249.78",hop_ip="138.197.249.78",hop_number=2i,hop_rtt_ms=3.3429999351501465 1525474707000000000
> traceroute_hop_data,column_number=0,host=m1.cloudpbx.ca,target_fqdn=www.google.com,target_ip=172.217.0.100 hop_fqdn="72.14.219.10",hop_ip="72.14.219.10",hop_number=3i,hop_rtt_ms=2.0139999389648438 1525474707000000000
> traceroute_hop_data,column_number=1,host=m1.cloudpbx.ca,target_fqdn=www.google.com,target_ip=172.217.0.100 hop_fqdn="162.243.190.33",hop_ip="162.243.190.33",hop_number=3i,hop_rtt_ms=3.315999984741211 1525474707000000000
> traceroute_hop_data,column_number=2,host=m1.cloudpbx.ca,target_fqdn=www.google.com,target_ip=172.217.0.100 hop_fqdn="162.243.190.33",hop_ip="162.243.190.33",hop_number=3i,hop_rtt_ms=2.9059998989105225 1525474707000000000
> traceroute_hop_data,column_number=0,host=m1.cloudpbx.ca,target_fqdn=www.google.com,target_ip=172.217.0.100 hop_fqdn="108.170.250.225",hop_ip="108.170.250.225",hop_number=4i,hop_rtt_ms=1.559000015258789 1525474707000000000
> traceroute_hop_data,column_number=1,host=m1.cloudpbx.ca,target_fqdn=www.google.com,target_ip=172.217.0.100 hop_fqdn="108.170.250.241",hop_ip="108.170.250.241",hop_number=4i,hop_rtt_ms=0.7829999923706055 1525474707000000000
> traceroute_hop_data,column_number=2,host=m1.cloudpbx.ca,target_fqdn=www.google.com,target_ip=172.217.0.100 hop_fqdn="108.170.250.225",hop_ip="108.170.250.225",hop_number=4i,hop_rtt_ms=1.5080000162124634 1525474707000000000
> traceroute_hop_data,column_number=0,host=m1.cloudpbx.ca,target_fqdn=www.google.com,target_ip=172.217.0.100 hop_fqdn="216.239.35.229",hop_ip="216.239.35.229",hop_number=5i,hop_rtt_ms=2.947000026702881 1525474707000000000
> traceroute_hop_data,column_number=1,host=m1.cloudpbx.ca,target_fqdn=www.google.com,target_ip=172.217.0.100 hop_fqdn="216.239.35.231",hop_ip="216.239.35.231",hop_number=5i,hop_rtt_ms=2.9040000438690186 1525474707000000000
> traceroute_hop_data,column_number=2,host=m1.cloudpbx.ca,target_fqdn=www.google.com,target_ip=172.217.0.100 hop_fqdn="216.239.35.229",hop_ip="216.239.35.229",hop_number=5i,hop_rtt_ms=2.5940001010894775 1525474707000000000
> traceroute_hop_data,column_number=0,host=m1.cloudpbx.ca,target_fqdn=www.google.com,target_ip=172.217.0.100 hop_fqdn="yyz10s13-in-f4.1e100.net",hop_ip="172.217.0.100",hop_number=6i,hop_rtt_ms=2.010999917984009 1525474707000000000
> traceroute_hop_data,column_number=1,host=m1.cloudpbx.ca,target_fqdn=www.google.com,target_ip=172.217.0.100 hop_fqdn="yyz10s13-in-f4.1e100.net",hop_ip="172.217.0.100",hop_number=6i,hop_rtt_ms=0.6510000228881836 1525474707000000000
> traceroute_hop_data,column_number=2,host=m1.cloudpbx.ca,target_fqdn=www.google.com,target_ip=172.217.0.100 hop_fqdn="yyz10s13-in-f4.1e100.net",hop_ip="172.217.0.100",hop_number=6i,hop_rtt_ms=0.6190000176429749 1525474707000000000
```


Built by [mattfung](https://github.com/mattfung)

Sponsored by [CloudPBX](http://CloudPBX.ca) with generous support by the NSERC Experience Award.