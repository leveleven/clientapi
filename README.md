# clientapi

## 依赖
```bash
wget -c https://golang.google.cn/dl/go1.18.9.linux-arm64.tar.gz -O - | sudo tar -xz -C /usr/local
export PATH=$PATH:/usr/local/go/bin

apt install lshw
```

## API调用方式

示例

- 获取硬件状态
```bash
curl http://127.0.0.1:8753/metrics
```

response:
```json
{
  "data": {
    "Memory": {
      "Total": 4093992960,           // int B, -1表示获取异常
      "Free": 1986834432,            // int B, -1表示获取异常
      "Percent": 25.689988094166143  // int %, -1表示获取异常
    },
    "CPU": {
      "Percent": 8.375634517766944,  // int %, -1表示获取异常
      "Temp": 36                     // int °C, -1表示获取异常
    },
    "Disk": {
      "status": health / unhealth / no_sata_disk / error - 获取错误
    }
  },
  "error_code": 0 / 1,                // 1为有错误
  "error_msg": string
}
```

- 获取网络配置
``` bash
curl 127.0.0.1:8753
```

response:
```json
{
  "Name": "eth0",                  // string
  "Address": "8c:14:7d:d3:5e:9a",  // string MAC
  "IP": "192.168.2.253",           // string
  "Netmask": 24                    // int
}
```

- 修改网络配置
```bash
curl 127.0.0.1:8753/netcfg \
    -H 'content-type:application/json' \
    -X POST \
    -d "{\"address\":\"192.168.1.2\",\"netmask\":\"255.255.255.0\",\"gateway\":\"192.168.1.1\",\"dns\":[\"114.114.114.114\",\"8.8.8.8\"],\"need_reboot\":false}"
// need_reboot 为 bool 类型，0为需要重启，1为否
```

请求体：
```json
{
  "address": string,
  "netmask": string,
  "gateway": string,
  "dns":     []string,
  "need_reboot": bool
}
```

response:
```json
{
	"info": "restart machine to apply new configure."
}
```

- 获取SN码
```bash
curl 127.0.0.1:8753/sn
```

response:
```json
{
  "data": {
    "serial": "xjcc00000000"
  },
  "error_code": 0,
  "error_msg": ""
}
```

## 测试

### sn码测试
- 在/root/目录下面创建.env文件，并在文件中输入测试sn码，如：
```text
SN="xjcc123123123"
```
- 保存好文件后，重启clientapi服务生效
```bash
systemctl restart clientapi
```

## 更新日志

2023-2-7  API增加SN码获取，标签重启动作

2023-2-20 修改SN码获取途径，优化硬盘获取方式