# Technical Test at Scalingo

## Architecture


The project code is organized as follows:

1. `config/`: This directory contains the configuration of the application, including environment variables.
2. `models/`: This folder contains the model definition for our API such as the `Repo` and `Language`.
3. `server/`: This directory is used to define the API routes and their handlers.
4. `main.go`: This file is used to initialize the application and start the server.


## Decision Making

- I didn't add a way to get GitHub token from the user because I didn't have time to implement an OAuth flow.
- I use `goroutine` to parallelize the repositories data fetching and the languages data fetching, `Mutex` to synchronize the access to the shared data and `WaitGroup` to wait for all the goroutines to finish to avoid race conditions.
- I didn't implement a login system because the API will be public. 
- I try to use not a lot of dependencies to keep the project simple and focused.


## Execution

```bash
docker compose up
```

The default port is `5000`. You can access the API at `http://localhost:5000`.


## Example

- To get the last 100 repositories data:

```bash
$ curl  "http://localhost:5000/repos"
```

-  Filter the data with query params:
```bash
$ curl  "http://localhost:5000/repos?language=Python"
```

- Get the statistics of the repositories languages:

```bash
$ curl  "http://localhost:5000/stats"
```

## Improvements

- Implement a login system
- Implement a way to get the token from the user
- Have a Swagger documentation 
- Uniformize the version of Go in the Dockerfile and in the `go.mod` file
- Add some tests
