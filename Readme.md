# HTTP Denial-of-Service protection system

Implement a simple HTTP Denial-of-Service protection system. The solution should reside in different modules (one for the server, one for the client).

1. Client
    * CLI application.
    * It should accept two arguments:
        - The number of clients
        - The number of concurrent workers/threads per client.
    * Each client should have a unique identifier (client ID).
    * Each concurrent worker should repeatedly do the following:
        - Send an HTTP request to a server with the client ID as a query parameter (
          e.g. http://localhost:8080/?clientId=3).
        - Wait some random amount of time.
    * The HTTP workers should run simultaneously (concurrently) without blocking each other.
    * The client should run until a key is pressed (e.g. Ctrl/Cmd+C), and after that it should gracefully drain all
      requests (wait for all of them to complete) and exit.
      An example:
      Assume the selected number of clients is 2 (e.g. clientId=1 and clientId=2), and the number of workers per client
      is 3. This means the app should start 6 concurrent workers, 3 of them should be sending the requests with client
      id 1, and the other 3 with the client id 2.
   
2. Server
    * It should expose an endpoint that for each incoming HTTP request does the following:
        - Using the client ID parameter check if this specific client has reached the maximum allowed number of requests
          per time frame (e.g. no more than 5 requests in 5 seconds).
        - The client should get an appropriate HTTP response depending on whether it has reached the threshold or not.
        - Note: The time frame starts on each client's first request and ends 5 seconds later (the time frame is fixed).
          After the time frame has ended, the client's first subsequent request opens a new time frame and so on.
    * The HTTP server should handle the requests concurrently.
