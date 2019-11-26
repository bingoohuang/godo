# godo

call shell and sleep in go-routine in an infinite loop.

```bash
$ godo -h
Usage of godo:
  -nums string
    	numbers range, eg 1-3 (default "1")
  -setup string
    	setup num
  -shell string
    	shell to invoke
  -span string
    	time span to do sth, eg. 1h, 10m for fixed span, or 10s-1m for rand span among the range (default "10m")
```

shell demo, demo.sh

```bash
#!/usr/bin/env bash

godoNum=${GODO_NUM}

if ((godoNum == 0)); then
    echo 做点准备工作
elif ((godoNum == 1)); then
    echo 做点事情1
elif ((godoNum == 2)); then
    echo 做点事情2
elif ((godoNum == 3)); then
    echo 做点事情3
elif ((godoNum == 4)); then
    echo 做点事情4
elif ((godoNum == 5)); then
    echo 做点事情5
fi
```

call demo.sh :

```bash
nohup godo -shell ./demo.sh -setup 0  -span 10s-1m -nums 1-5  2>&1 >> godo.out &
```
