# Robot Dreams course homework 2: Load testing environment.

The main components are the following

- Nginx load balancer, 1GB of RAM, 1 CPU
- Two API servers, written in Go. 1 GB of RAM for each, 1 CPU for each
- PostgreSQL server, 5 GB of RAM, 1 CPU


Notable details:

- The go servers are using [the fasthttp package](https://github.com/valyala/fasthttp), with [the fasthttp router](github.com/valyala/fasthttprouter)
- The nginx balancer uses the Least connected load balancing method.
