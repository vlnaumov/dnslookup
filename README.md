# dnslookup

```
Usage:
  -domain string
        Domain to lookup
  -in string
        Input file name (csv)
  -out string
        Output file name (default "/dev/stdout")
  -parall int
        Maximum parallelism (default 1)

go run main.go -in ./ip.csv -out ./result.txt -parall 5 -domain "google.com"
```
