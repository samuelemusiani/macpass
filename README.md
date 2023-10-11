# MACPass

This program provides a way to authenticate users through an LDAP server and
check if they enter a MAC address correctly. In this case it output the 
correct MAC address associated with the user.

# How it should work

- First it ask the user his account and passoword
- Then try to authenticate the user in an ldap server
- If the auth is successful, it asks for a MAC address and verify it
- If the MAX add. is correct the it append the result to a file
- From there the deamon take on

I should probably use a file for configuring the program
