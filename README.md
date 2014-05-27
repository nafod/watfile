watfile
=======

watfile is a simple, secure, and performant file uploader. The code in this repository currently powers the service at [watfile.com](https://watfile.com).

Setup
-----

First, you must install the dependencies [gcfg](https://code.google.com/p/gcfg/) and [mysql](https://github.com/go-sql-driver/mysql).

    go get code.google.com/p/gcfg
    go get github.com/go-sql-driver/mysql

Then, copy `watfile.conf.default` to `watfile.conf` and edit the configuration settings. In particular, make sure to include the relevant information to connect to the database.

Finally, you should be able to compile watfile and run it! Make sure to place `watfile.conf` in the same directory as the executable is running from. The executable itself will create any subdirectories it needs.
