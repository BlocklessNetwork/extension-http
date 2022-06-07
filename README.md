# blockless-extension-http

This this example of blockless driver write by golang. And the driver also support rust, c/c++ and others.
## How to build
```
$ go build go build -buildmode=c-shared -ldflags "-s -w" -o http_driver.so  
```