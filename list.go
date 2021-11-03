package main

const LISTMESSAGE = `emailremote
===========

Runs a commend for a remote user be check an email account and return
the results.

Commands are sent on the subject line. Each command has only one option
except LIST.

Commands
--------

-   LIST
    -   Returns this README
-   PING
    -   Ping remote host. ipv4 or ipv6
-   TRACE
    -   Traceroute to remote host. ipv4 or ipv6
-   MTR
    -   MTR -c 10 -r
-   RVIEWS
    -   Queries route-views. ipv4 only
-   CIDR
    -   Parse a CIDR in to subnets

Addresses
---------

Addresses must be well-formed as there is minimum bounds checking.

RVIEWS requires a CIDR block. Example 8.8.8.0/24
`

func List(f string) {
	_ = ReturnEmail(f, "LIST", LISTMESSAGE) // Ignore errors
	//if err != nil{
	//return err
	//}
}
