POST Request
```sh
curl -X POST http://localhost:8080/temperature \
  -H "Content-Type: application/json" \
  -d '{"temperature": 23.5}'
```

GET Request
```sh
curl -X GET http://localhost:8080/temperature
```


GET Request with limit
```sh
curl -X GET http://localhost:8080/temperature?limit=10
```
