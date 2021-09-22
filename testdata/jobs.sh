### envs
export addr=http://192.168.126.71:8080
export gap=10s

### sleep200us
blow ${addr}/sleep200us --report json -d1m -c{100,350,50}
sleep 30s && gurl -print b -raw ${addr}/ps

### sleep400us
blow ${addr}/sleep400us --report json  -d1m -c{100,350,50}
sleep 30s && gurl -print b -raw ${addr}/ps

### pi20 sleep0us
blow ${addr}/pi20/sleep0us --report json  -d1m -c{10,350,10}
sleep 30s && gurl -print b -raw ${addr}/ps

### pi20/sleep200u
blow ${addr}/pi20/sleep200us --report json  -d1m -c{10,350,10}
sleep 30s && gurl -print b -raw ${addr}/ps

### pi20/sleep400u
blow ${addr}/pi20/sleep400us --report json  -d1m -c{10,350,10}
sleep 30s && gurl -print b -raw ${addr}/ps
