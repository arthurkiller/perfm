# RedisBenchmark on Perfm

## Feature
* support `redis-benchmark` like usage, you can easily switch to redisbenchmark-perfm
* add **RAND** & **RAND2** support, for more flexible testing case built
* build with `go`, easily to use and modify
* better result print with histogram

## Build
* you need to have `go` installed locally
* run `go build`
* set `env GOOS=linux` for cross platform build

## Usage
```
Usage of ./redis:
  -a string
        cluster auth
  -batching int
        pipeline mode, control batching size inside MULTI
  -c int
        number of total tests runned, this will dissable duration option
  -command string
        testing command, you can add __RAND__ or __RAND2__ as random field to the command, but for each field, only the last __RAND__ will be replaced (default "tr.getbit foo __RAND__")
  -d int
        testing duration in second (default 30)
  -h string
        cluster host (default "127.0.0.1")
  -p int
        number of parallel (default 4)
  -port int
        cluster port (default 6379)
  -r int
        random range for __RAND__ (default 100000000)
  -r2 int
        random range for __RAND2__ (default 100000000)
```

## Examples

* teseting redis set with `set` for 100000 random key and 10000000 random field range
    
```bash
  ./redis -h r-welcome.redis.zhangbei.rds.aliyuncs.com -a hello:world -d 30 -r 10000000 -r2 100000 -command "set foo-__RAND2__ bar-__RAND__" -p 20 -c 1
```

* teseting redis set with `del` for 100000 random key and 10000000 random field range
    
```bash
./redis -h r-welcome.redis.zhangbei.rds.aliyuncs.com -a hello:world -d 30 -r 10000000 -r2 100000 -command "del foo-__RAND2__ bar-__RAND__" -p 20 -c 1
```

* teseting tairroaring bitmap with `tr.getbit` for 100000 random key and 10000000 random field range
    
```bash
./redis -h r-welcome.redis.zhangbei.rds.aliyuncs.com -a hello:world -d 30 -r 10000000 -r2 100000 -command "tr.getbit foo-__RAND2__ __RAND__" -p 20 -c 1
```

* teseting tairroaring bitmap with `tr.getbits` for 100000 random key and 100 fields under 10000000 random range
    
```bash
./redis -h r-welcome.redis.zhangbei.rds.aliyuncs.com -a hello:world -d 30 -r 10000000 -r2 100000 -command "tr.getbits foo-__RAND2__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__ __RAND__" -p 20
```

* teseting tair roaring bitmap with `tr.bitopcard xor`
    
```bash
./redis -h r-welcome.redis.zhangbei.rds.aliyuncs.com -a hello:world -d 60 -r 15000 -r2 10000 -command "tr.bitopcard XOR foo-__RAND2__ foo-__RAND2__" -p 10
```

## Welcome to try perfm to build more benchmark cases
