# Test Server
This directory contains some basic HTML pages, meant for testing the Load Balancer.

## Running the server
To run this server we will use the Python's http pkg.

To run a server at port 5000, use:
```
python3 -m http.server 5000
```

Spin up multiple such servers, in this server directory to mimic an environment with multiple, distributed servers of the same application.

The Load Balancer should run at a different port(8000 is recommended) than all the ports used for the servers.