# MACPass

This program provides a way to authenticate users through an LDAP server and
check if they enter a MAC address correctly. In this case it output the 
correct MAC address associated with the user.

## How it should work

- First it ask the user his account and passoword
- Then try to authenticate the user in an ldap server
- If the auth is successful, it asks for a MAC address and verify it
- If the MAX add. is correct the it append the result to a file
- From there the deamon take on

I should probably use a file for configuring the program

# MACPassd

A program that interact with iptables to manage a firewall for MAC addresses.
It provides a way to insert MAC addresses that are allowed to go to the internet.

This should be run as a service, it's not a proper daemon.

## How it should work

- For a config file we read the input file
- In the input file [MACPAss](github.com/musianisamuele/macpass) provides a MAC
address followed by the username binded to it
- We allow the MAC address to navigate through iptables
- We periodically check if the MAC is still connected, if after a time the MAC
does not respond we block it in the firewall.
- We log everything :)

