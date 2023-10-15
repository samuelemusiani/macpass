# MACPass

This program provides a way to authenticate users through an LDAP server and
check if they enter a MAC address correctly. In this case it output the 
correct MAC address associated with the user.

## How it (should) work

- First it asks the user his account and password
- Then try to authenticate the user in an ldap server specified in the config file
- If the auth is successful, it asks for a MAC address and verifies it
- If the MAC is correct then send a new request to macpassd through a Unix socket
- From there the macpassd deamon take on

The configuration file is `config.toml` and for now the only possible path for 
his location is in the same directory of the program

# MACPassd

A program that interact with iptables to manage a firewall for MAC addresses.
It provides a way to insert MAC addresses that are allowed to go to the internet.

This should be run as a service, it's not a proper daemon.

## How it (should) work

- The daemon listens through a Unix socket for incoming requests
- When a new request is red, the request provides a MAC address followed by the 
username bind to it
- The daemon allows the MAC address to navigate through iptables
- Periodically, a check is made to check if the MAC is still connected, if after 
a time the MAC does not respond we block it in the firewall. (TODO)
- We log everything :)
