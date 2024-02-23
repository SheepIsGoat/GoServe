# GoServe
Work in progress! The only directory that is correct currently is content_server, which could be more appropriately named app_server.

General Golang backend server designed to be generally extensible for building web apps.

Specs:
- Front end
    - windmill-dashboard-master template
    - htmx
    - alpine.js
- Back end
    - golang
    - echo
    - postgres
    - python (for model serving)



## create your postgres instance
```
chmod +x postgres.sh
./postgres.sh
```

## create and start the server
```
go build
./main
```