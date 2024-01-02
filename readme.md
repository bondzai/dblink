# Hexagonal architecture in Go.

## Application structure
```
└── Messenger
   ├── cmd
   │   └── main.go
   ├── go.mod
   ├── go.sum
   └── internal
       ├── adapters
       │   ├── handler
       │   │   └── http.go
       │   └── repository
       │       ├── postgres.go
       │       └── redis.go
       └── core
           ├── domain
           │   └── model.go
           ├── ports
           │   └── ports.go
           └── services
               └── services.go

```

Credit: https://www.golinuxcloud.com/hexagonal-architectural-golang/