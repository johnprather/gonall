# gonall

A golang "onall" tool for shell use when wanting to run a command on
multiple hosts.

## Usage

gonall reads a list of hostnames from stdin, using ssh to run the specified
command on each. All output lines are prepended with the originating server's
hostname.  Stdout and stderr are used appropriately to allow easy separation
of valid result output and error text.

gonall attempts to politely stream both the server list from stdin and
stdout/stderr from the ssh jobs.  This should make it suitable for using
with either long/infinite-running processes providing the server list to stdin,
or long/infinite-running ssh jobs sending output.

```
Usage:

  gonall [options] command < hostfile

  ... | gonall [options] <command>

Description:

 Options:

  -d int
    	delay (ms) between job starts (default 10)
  -n int
    	number parallel ssh jobs (default 6)
  -p	prompt for a password to use
  -t int
    	timeout (secs) for connect to server
  -u string
    	username to use
```

One can (and probably should) create a config file.

The config file is a yaml file which needs to be created at _$HOME/.gonall.yml_.
It contains an array of configurations, which will be respected in a
first-come-first-serve manner (like ~/.config for openssh) to define which
options will be used for which hosts.

This example configuration would cause gonall to use user www for all target
hosts, with a connect timeout of 30 seconds:

```
- Host: "*"
  User: www
  Timeout: 30
```

The below example demonstrates a configure where hosts matching _ops*_ would
use user "sre", hosts matching _oob*_ would use user "admin" on port 22000,
and all other hosts would be connected to using user "www".

```
- Host: "ops*"
  User: sre
- Host: "oob*"
  User: admin
  Port: 22000
- Host: "*"
  User: www
```

The below more complex example sets up ssh proxy hosts.  A ProxyHost value of
"none" results in no proxy host being used.  The below starts with a specific
rule for our proxy hosts to ensure that when we do try throw their hostnames
at gonall, we connect to them directly.  The next more general rule will set,
for the domain, an proxy server (except for the proxy servers due to
the previous rule) and proxy username to use for all hosts on the domain.

```
- Host: "secure*.my-domain.tld"
  User: "joe.smith"
  ProxyHost: none

- Host: "*.my-domain.tld"
  User: "sre"
  ProxyHost: "secure1.my-domain.tld"
  ProxyUser: "joe.smith"
```

And you can even nest proxies.  The config below should result in connections
that pass through secure1.my-domain.tld, then secure1.internal-domain.tld,
to connect to \*.internal.my-domain.tld hosts.

__Note__ that this functionality allows you to create cyclic proxy
dependencies, resulting in the program crashing out of infinite recursion.

```
- Host: "secure*.internal-domain.tld"
  User: "joe.smith"
  ProxyHost: "secure1.my-domain.tld"

- Host: "secure*.my-domain.tld"
  User: "joe.smith"
  ProxyHost: none

- Host: "*.internal-domain.tld"
  User: sre
  ProxyHost: "secure1.internal-domain.tld"

- Host: "*.my-domain.tld"
  User: www
  ProxyHost: "secure1.my-domain.tld"
```
