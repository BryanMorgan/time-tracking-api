# Time Tracking API

<img align="right" width="128" src="https://user-images.githubusercontent.com/479339/74610686-49244e80-50aa-11ea-8a3d-dd4a11856d6c.png">

![build](https://github.com/BryanMorgan/time-tracking-api/workflows/build/badge.svg?branch=master&event=push)
[![Go Report Card](https://goreportcard.com/badge/github.com/BryanMorgan/time-tracking-api)](https://goreportcard.com/report/github.com/BryanMorgan/time-tracking-api)

Go API for tracking time. 

Manages time entries for tasks that are associated with projects. 

# Running
To run the API server you can run the Makefile target:

```make run``` 

which will start the dev server on the port configured in `config/dev.yml`

# Testing

## Unit Tests
Unit tests can be run using:

```make unit_test```

## Integration Tests
Integration tests are managed under the `integration_test` root folder and can be run using:

```make integration_test```

## Postman Tests
Additional functional tests are available using he [Postman](https://www.postman.com/) tool. 
These tests require the [newman](https://github.com/postmanlabs/newman) Postman command-line runner. Install using:

```npm install -g newman```

Also relies on the `database/bootstrap.sql` data to be present. To run the Postman tests locally, first start the web server:

```make run```

then run the Postman tests:

```make postman```
