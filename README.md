Snowflake
=========

`snowflake` is a HTTP server for generating incrementing ID numbers that are 
unique

This is a Go implementation of [Twitter Snowflake](https://blog.twitter.com/2010/announcing-snowflake)

Usage
-----

    curl http://server:port?count=n 

    where 1 ≤ n ≤ 500

each id will be a 64-bit number represented as a string, structured as follows:

    6  6         5         4         3         2         1         
    3210987654321098765432109876543210987654321098765432109876543210
    
    ttttttttttttttttttttttttttttttttttttttttttmmmmmmmmmmssssssssssss

where

* `s` (sequence) is a 12-bit integer that increments if called multiple times for the same millisecond
* `m` (machine id) is a 10-bit integer representing the server id
* `t` (time) is a 42-bit integer representing the current timestamp in milliseconds
    * the number of milliseconds to have elapsed since 1413817200000 or 2014-10-20T15:00:00.000Z

Licence
-------

Apache License 2.0

Credits
-------

Fork of [najeira/snowflake](https://github.com/najeira/snowflake)