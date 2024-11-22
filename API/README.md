# OVERVIEW
This project is implementation of rate limiting algorithms in API using Go language and Echo Framework

# HOW TO TEST
Make sure you'are in API director. ```cd API```

## Token Bucket Algorithm
```sh
go test -run TestTokenBucketMiddleware
```
## Leaky Bucket Algorithm
```sh
go test -run TestLeakyBucketMiddleware
```
## Fixed Window Counter Algorithm
```sh
go test -run TestFixedWindowMiddleware
```
## Sliding Window Logs Algorithm
```sh
go test -run TestSlidingWindowMiddleware
```