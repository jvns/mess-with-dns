curl -v -X POST -H "Content-Type: application/json" -H "X-API-Key: not-a-secret"\
    -d '{"name": "example.org.", "kind": "Native", "masters": [], "nameservers": ["ns1.example.org.", "ns2.example.org."]}' http://localhost:8081/api/v1/servers/localhost/zones
