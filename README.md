# godo

call shell and sleep in go-routine in an infinite loop.

```bash
$ godo -h
Usage of godo:
  -fla9 string  Flags config file, a scaffold one will created when it does not exist.
  -init Create initial ctl and exit
  -nums string  numbers range, eg 1-3 (default "1")
  -setup string setup num
  -shell string shell to invoke
  -span string  time span to do sth, eg. 1h, 10m for fixed span, or 10s-1m for rand span among the range (default "10m")
  -version (-v) Create initial ctl and exit
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
godo -shell ./demo.sh -setup 0  -span 10s-1m -nums 1-5  -d
```

```sh
$ godo -span 1m -shell "ps aux|awk 'NR == 1 || \$2==27535'"
2021/07/07 16:03:38 nothing to setup
2021/07/07 16:03:38 start to do sth
USER       PID %CPU %MEM    VSZ   RSS TTY      STAT START   TIME COMMAND
root     27535 58.9 10.8 2191460 864160 pts/1  Sl   14:43  47:15 beefs server -dir=/home/beefs/data -volume.max=80
2021/07/07 16:04:38 start to do sth
USER       PID %CPU %MEM    VSZ   RSS TTY      STAT START   TIME COMMAND
root     27535 60.0 12.9 2191460 1030644 pts/1 Sl   14:43  48:42 beefs server -dir=/home/beefs/data -volume.max=80
```
