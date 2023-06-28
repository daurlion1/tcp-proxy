TCP Proxy
Assumptions and requirements of Proxy app.

  Assumptions:
1) the address to which connections are forwarded is considered constant for simplicity (i.e. we forward all connections to the same address), can be set by constants / obtained from the config - it does not matter, it is static for us
2) the data is not processed, just sent as is.
  
  Requirements:
1) processing of connections is carried out by the gorutin pool, if there are no free connections, we do not process new connections (there is no waiting queue)
2) data transfer should be carried out in both directions: client-server and server-client
3) inactive connections (for which there has been no data transfer for a long time) are closed
4) the connection is closed if it is closed by any of the endpoints (either the client or the server)
5) graceful shutdown - it is necessary to implement the correct completion of all processing gorutins (data that has already been received must be sent before completion, new data is not accepted)

Repository consists of 2 apps named server and proxy.
server - simple app for receiving messages from proxy.
proxy - app which receiving messages from clients and sending to server,
has own assumptions and requirements.

Guideline for start projects:
1) write command to terminal:  docker-compose up
   
   ![My Image](images/my-image.jpg)
2) connect as client to proxy write command to terminal : telnet localhost 8080

Checks


