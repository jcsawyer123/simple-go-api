PROFILING_PORT=6060 go run cmd/server/main.go

# CPU profile
go tool pprof http://localhost:6060/debug/pprof/profile

# Memory profile
go tool pprof http://localhost:6060/debug/pprof/heap

# Goroutine profile
go tool pprof http://localhost:6060/debug/pprof/goroutine



go install github.com/rakyll/hey@latest
hey -n 1000 -c 100 http://localhost:3000/your-endpoint



### Without Cache
❯ hey -n 1000 -c 100 -H "x-aims-auth-token: $AL_TOKEN" http://localhost:8080/api/test

Summary:
  Total:	20.4856 secs
  Slowest:	3.0480 secs
  Fastest:	0.7095 secs
  Average:	1.9667 secs
  Requests/sec:	48.8148

  Total data:	63000 bytes
  Size/request:	63 bytes

Response time histogram:
  0.710 [1]	|
  0.943 [6]	|■
  1.177 [17]	|■■
  1.411 [37]	|■■■■
  1.645 [97]	|■■■■■■■■■■■
  1.879 [189]	|■■■■■■■■■■■■■■■■■■■■■■
  2.113 [344]	|■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  2.346 [192]	|■■■■■■■■■■■■■■■■■■■■■■
  2.580 [79]	|■■■■■■■■■
  2.814 [27]	|■■■
  3.048 [11]	|■


Latency distribution:
  10% in 1.5199 secs
  25% in 1.7817 secs
  50% in 1.9841 secs
  75% in 2.1634 secs
  90% in 2.3823 secs
  95% in 2.5204 secs
  99% in 2.8223 secs

Details (average, fastest, slowest):
  DNS+dialup:	0.0004 secs, 0.7095 secs, 3.0480 secs
  DNS-lookup:	0.0001 secs, 0.0000 secs, 0.0031 secs
  req write:	0.0000 secs, 0.0000 secs, 0.0021 secs
  resp wait:	1.9662 secs, 0.7095 secs, 3.0479 secs
  resp read:	0.0000 secs, 0.0000 secs, 0.0005 secs

Status code distribution:
  [200]	1000 responses


### With Cache
❯ hey -n 1000 -c 100 -H "x-aims-auth-token: $AL_TOKEN" http://localhost:8080/api/test

Summary:
  Total:	7.6292 secs
  Slowest:	1.5804 secs
  Fastest:	0.1491 secs
  Average:	0.7119 secs
  Requests/sec:	131.0749

  Total data:	63000 bytes
  Size/request:	63 bytes

Response time histogram:
  0.149 [1]	|
  0.292 [7]	|■
  0.435 [37]	|■■■■
  0.578 [217]	|■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.722 [339]	|■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.865 [238]	|■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  1.008 [75]	|■■■■■■■■■
  1.151 [22]	|■■■
  1.294 [45]	|■■■■■
  1.437 [8]	|■
  1.580 [11]	|■


Latency distribution:
  10% in 0.4830 secs
  25% in 0.5699 secs
  50% in 0.6855 secs
  75% in 0.7955 secs
  90% in 0.9524 secs
  95% in 1.2172 secs
  99% in 1.4375 secs

Details (average, fastest, slowest):
  DNS+dialup:	0.0004 secs, 0.1491 secs, 1.5804 secs
  DNS-lookup:	0.0002 secs, 0.0000 secs, 0.0032 secs
  req write:	0.0000 secs, 0.0000 secs, 0.0011 secs
  resp wait:	0.7114 secs, 0.1490 secs, 1.5750 secs
  resp read:	0.0000 secs, 0.0000 secs, 0.0004 secs

Status code distribution:
  [200]	1000 responses
