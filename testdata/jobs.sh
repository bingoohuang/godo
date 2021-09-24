### envs
export addr=http://192.168.126.71:8080
export gap=10s

### sleep200/400us
blow ${addr}/sleep{200,400,200}us --report json -d1m -c{100,350,50}
sleep 30s && gurl -print b -raw ${addr}/ps

### pi20 sleep0/200/400us
blow ${addr}/pi20/sleep{0,400,200}us --report json  -d1m -c{10,350,10}
sleep 30s && gurl -print b -raw ${addr}/ps
