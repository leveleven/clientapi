# clientapi

## 依赖
```bash
wget -c https://golang.google.cn/dl/go1.18.9.linux-arm64.tar.gz -O - | sudo tar -xz -C /usr/local
export PATH=$PATH:/usr/local/go/bin

apt install lshw
```

## API调用方式

示例
```bash
// 获取硬件状态
curl http://127.0.0.1:8753/metrics

response:
{
  "Memory": {
    "Total": 4093992960,           // int B
    "Free": 1986834432,            // int B
    "Percent": 25.689988094166143  // int %
  },
  "CPU": {
    "Percent": 8.375634517766944,  // int %
    "Temp": 36                     // int °C
  }
  "Disk": {
    "status": health / unhealth / no_sata_disk
  }
}

// 获取网络配置
curl 127.0.0.1:8753

response:
{
  "Name": "eth0",                  // string
  "Address": "8c:14:7d:d3:5e:9a",  // string MAC
  "IP": "192.168.2.253",           // string
  "Netmask": 24                    // int
}

// 修改网络配置
curl 127.0.0.1:8753/netcfg 
    -H 'content-type:application/json' \
    -X POST
    -d "{\"address\":\"192.168.1.2\",\"netmask\":\"255.255.255.0\",\"gateway\":\"192.168.1.1\",\"dns\":[\"114.114.114.114\",\"8.8.8.8\"]}"

response:
{
	"info": "restart machine to apply new configure."
}
```