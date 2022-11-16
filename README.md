# clientapi

## API调用方式

示例
```bash
curl http://127.0.0.1:8753/metrics

response:
{
  "Memory": {
    "Total": 4093992960,
    "Free": 1986834432,
    "Percent": 25.689988094166143
  },
  "CPU": {
    "Percent": 8.375634517766944,
    "Temp": 36
  }
}
```