

# Build

For a specific platform:
```
env GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o dnspulse_exporter
```

Or for the native platform:
```
go build -ldflags "-s -w" -o dnspulse_exporter
```

