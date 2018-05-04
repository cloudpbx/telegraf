# traceroute Input Plugin

The traceroute plugin provides routing information given end host.

### Configuration:

```toml
# NOTE: this plugin forks the traceroute command. You may need to set capabilities
# via setcap cap_net_raw+p /bin/traceroute
[[inputs.traceroute]]
  ## List of urls to traceroute
  urls = ["www.google.com"] # required
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
    - field1 (type, unit)
    - field2 (float, percent)

- traceroute_hop_data
  - tags:
    - target_fqdn
    - target_ip (IPv4 string)
    - hop_number
    - column_number (zero-indexed value representing which column of the traceroute output the data resides in)
  - fields:
    - hop_fqdn
    - hop_ip (IPv4 string)
    - rtt (round trip time in ms)

### Sample Queries:

This section should contain some useful InfluxDB queries that can be used to
get started with the plugin or to generate dashboards.  For each query listed,
describe at a high level what data is returned.

Get the max, mean, and min for the measurement in the last hour:
```
SELECT max(field1), mean(field1), min(field1) FROM measurement1 WHERE tag1=bar AND time > now() - 1h GROUP BY tag
```

### Example Output:

This section shows example output in Line Protocol format.  You can often use
`telegraf --input-filter <plugin-name> --test` or use the `file` output to get
this information.

```
measurement1,tag1=foo,tag2=bar field1=1i,field2=2.1 1453831884664956455
measurement2,tag1=foo,tag2=bar,tag3=baz field3=1i 1453831884664956455
```

Built by [mattfung](https://github.com/mattfung)

Sponsored by [CloudPBX](http://CloudPBX.ca) with generous support by the NSERC Experience Award.