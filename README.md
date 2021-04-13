# greeter


Testcase for KrakenD Urlencode bug

Usage:


```
$ krakend run --config krakend.json --debug
$ go run .

```

greeter Swagger UI at: <http://localhost:3000/docs>


## Endpoint (Backend greeter): /api/hello/{name}

Ok Case:

```
$ curl -X 'GET' \
  'http://localhost:3000/api/hello/World' \
  -H 'accept: application/json'

$ curl -X 'GET' \
  'http://localhost:3000/api/hello/World%2Fof%2Fbugs' \
  -H 'accept: application/json

```  

### Endpoint (Backend greeter): /api/hello/{name}/{bad1}/{bad2}


Wildcard notices: there is no (!) /api/hello/{name}/{bad1}

Calling this endpoint will always fail.


```
curl -X 'GET' \
  'http://localhost:3000/api/hello/World/of/bugs' \
  -H 'accept: application/json'

```  


### KrakenD Bug

Ok Case:

```
$ curl -X 'GET' \
  'http://localhost:8080/api/hello/World' \
  -H 'accept: application/json'

```  

Bugs - wrong endpoint is called:


```

# not existing endpoint - 404
$ curl -X 'GET' \
  'http://localhost:8080/api/hello/World%2Fof' \
  -H 'accept: application/json'

# wrong endpoint - 404
$ curl -X 'GET' \
  'http://localhost:8080/api/hello/World%2Fof%2Fbugs' \
  -H 'accept: application/json'

```  