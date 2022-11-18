Telnet
==============================

The telnet server is meant for debugging, which needs to comply with Redis standard ( https://redis.io/topics/protocol ) for interoperability.
This means you can use redis-cli with the tsxdb telnet server for testing and debugging.

In order to go into Redis-mode, send "COMMAND" as first (before auth) if the cli does not already to this for you.