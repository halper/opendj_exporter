# OpenDJ Prometheus Exporter

This is a simple service that scrapes metrics from OpenDJ and exports them via HTTP for Prometheus consumption.

This exporter is mostly a modification of https://github.com/tomcz/openldap_exporter.

## Setting up OpenDJ for monitoring

Please refer to the [LDAP Based Monitoring](https://github.com/OpenIdentityPlatform/OpenDJ/wiki/Monitoring%2C-Logging%2C-and-Alerts#171-ldap-based-monitoring) for the details of monitoring in OpenDJ.

Once you've built the exporter (see below), you can install it on the same server as your OpenDJ instance, and run it as a service. You can then configure Prometheus to pull metrics from the exporter's `/metrics` endpoint on port 9330, and check to see that it is working via curl:

```
$> curl -s http://localhost:9330/metrics
...
# HELP opendj_administration_connector Administration connector
# TYPE opendj_administration_connector gauge
opendj_administration_connector{attr="ds-connectionhandler-num-connections"} 0
# HELP opendj_administration_connector_statistics Administration connector statistics
# TYPE opendj_administration_connector_statistics gauge
opendj_administration_connector_statistics{attr="abandonRequests"} 0
opendj_administration_connector_statistics{attr="addRequests"} 0
opendj_administration_connector_statistics{attr="addResponses"} 0
opendj_administration_connector_statistics{attr="bindRequests"} 12
...
# HELP opendj_jvm_memory_usage JVM memory usage
# TYPE opendj_jvm_memory_usage gauge
opendj_jvm_memory_usage{attr="code-heap-non-nmethods-bytes-used-after-last-collection"} 0
opendj_jvm_memory_usage{attr="code-heap-non-nmethods-current-bytes-used"} 1.371776e+06
opendj_jvm_memory_usage{attr="code-heap-non-profiled-nmethods-bytes-used-after-last-collection"} 0
opendj_jvm_memory_usage{attr="code-heap-non-profiled-nmethods-current-bytes-used"} 1.1173376e+07
opendj_jvm_memory_usage{attr="code-heap-profiled-nmethods-bytes-used-after-last-collection"} 0
...
# HELP opendj_ldap_connection_handler LDAP connection handler metrics
# TYPE opendj_ldap_connection_handler gauge
opendj_ldap_connection_handler{attr="ds-connectionhandler-num-connections"} 1
# HELP opendj_ldap_connection_handler_statistics LDAP connection handler statistics
# TYPE opendj_ldap_connection_handler_statistics gauge
opendj_ldap_connection_handler_statistics{attr="abandonRequests"} 0
opendj_ldap_connection_handler_statistics{attr="addRequests"} 0
opendj_ldap_connection_handler_statistics{attr="addResponses"} 0
opendj_ldap_connection_handler_statistics{attr="bindRequests"} 47
...
# HELP opendj_ldaps_connection_handler LDAPS connection handler metrics
# TYPE opendj_ldaps_connection_handler gauge
opendj_ldaps_connection_handler{attr="ds-connectionhandler-num-connections"} 0
# HELP opendj_ldaps_connection_handler_statistics LDAPS connection handler statistics
# TYPE opendj_ldaps_connection_handler_statistics gauge
opendj_ldaps_connection_handler_statistics{attr="abandonRequests"} 0
opendj_ldaps_connection_handler_statistics{attr="addRequests"} 0
opendj_ldaps_connection_handler_statistics{attr="addResponses"} 0
opendj_ldaps_connection_handler_statistics{attr="bindRequests"} 9192
...
# HELP opendj_userRoot_backend userRoot backend
# TYPE opendj_userRoot_backend gauge
opendj_userRoot_backend{attr="ds-backend-entry-count"} 1
# HELP opendj_userRoot_je_database userRoot JE database
# TYPE opendj_userRoot_je_database gauge
opendj_userRoot_je_database{attr="EnvironmentActiveLogSize"} 25277
opendj_userRoot_je_database{attr="EnvironmentAdminBytes"} 125
opendj_userRoot_je_database{attr="EnvironmentApplicationPermits"} 0
opendj_userRoot_je_database{attr="EnvironmentAvailableLogSize"} 5.8729750528e+10
...
```

## Configuration

You can configure `opendj_exporter` using multiple configuration sources at the same time. All configuration sources are optional, if none are provided then the default values will be used.

The precedence of these configuration sources is as follows (from the highest to the lowest):

1. Command line flags
2. Environment variables
3. YAML configuration file parameters
4. Default values

```
NAME:
   opendj_exporter - Export OpenDJ metrics to Prometheus

USAGE:
   opendj_exporter [global options] [arguments...]

VERSION:
    ()

GLOBAL OPTIONS:
   --adminListenAddr value               The address that administration connector is listening (default: "0.0.0.0") [$ADMN_LSTN]
   --adminPort value                     OpenDJ Administration port (default: 4444) [$ADMN_PORT]
   --configFile YAML_FILE, -c YAML_FILE  Optional configuration from a YAML_FILE
   --interval value, -i value            Scrape interval (default: 30s) [$INTERVAL]
   --jsonLog                             Output logs in JSON format (default: false) [$JSON_LOG]
   --ldapAddr value, -l value            Address and port of OpenDJ server (default: "localhost:389") [$LDAP_ADDR]
   --ldapListenAddr value                The address that LDAP connection handler is listening (default: "0.0.0.0") [$LDAP_LSTN]
   --ldapPass value                      OpenDJ bind password (optional) [$LDAP_PASS]
   --ldapPort value                      OpenDJ LDAP port (default: 389) [$LDAP_PORT]
   --ldapsListenAddr value               The address that LDAPS connection handler is listening (default: "0.0.0.0") [$LDAPS_LSTN]
   --ldapsPort value                     OpenDJ LDAPS port (default: 636) [$LDAPS_PORT]
   --ldapUser value, -u value            OpenDJ bind username (optional) [$LDAP_USER]
   --metrics value, -m value             Path on which to expose Prometheus metrics (default: "/metrics") [$METRICS_PATH]
   --promAddr value, -a value            Bind address for Prometheus HTTP metrics server (default: ":9330") [$PROM_ADDR]
   --help, -h                            show help (default: false)
   --version, -v                         print the version (default: false)
```

Example:

```
INTERVAL=10s /usr/sbin/opendj_exporter --promAddr ":8080" --config /etc/opendj/exporter.yaml
```

Where `exporter.yaml` looks like this:

```yaml
---
ldapAddr: 127.0.0.1:1389
ldapUser: cn=Directory Manager
ldapPass: password
ldapsPort: 1636
```

NOTES:


## Build

1. Install Go 1.19 from https://golang.org/
2. Build the binaries: `make build`
