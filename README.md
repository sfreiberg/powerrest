# PowerRest

## About

Restful API for PowerDNS. It's written in Go (golang) and is extremely light on resources.

## Limitations
* PowerRest only supports MySQL at the moment. PostgreSQL support would be trivial but I don't have the bandwidth to test it at the moment.
* No authentication support. It should be used only in trusted environments or behind a reverse proxy that has authentication.

## Examples using Curl

### Domains

List domains

`curl "http://127.0.0.1/domains"`

Create new domain

`curl -X POST --data-binary '{ "name": "example.com" }' "http://127.0.0.1/v1/domains"`

Update a domain

`curl -X POST --data-binary '{ "name": "example.org" }' "http://127.0.0.1/v1/domains/1"`

Delete a domain

`curl -X DELETE "http://127.0.0.1/v1/domains/1"`

### Records

List records

`curl "http://127.0.0.1/records"`

Create new record

`curl -X POST --data-binary '{"domain_id":1,"name":"example.com","type":"A","content":"192.168.1.1","ttl":3600}' "http://127.0.0.1/v1/records"`

Update record

`curl -X POST --data-binary '{"domain_id":1,"name":"example.com","type":"A","content":"192.168.1.2","ttl":3600}' "http://127.0.0.1/v1/records/1"`

Delete record

`curl -X DELETE "http://127.0.0.1/v1/records/1"`