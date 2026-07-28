# healthchecks Go compatibility implementation

This is an independent Go implementation of the replayed Healthchecks HTTP
contract. It uses only Go's standard library; it does not invoke the Django
reference or use an uptime-monitoring package.

## Ubuntu 24.04

```bash
sudo apt update
sudo apt install -y golang-go
cd target
go build -o healthchecks .
./healthchecks
```

The server listens on port 8000 by default. Set `PORT` to use another port.

## Local validation

In a second terminal, from `healthchecks/relang`:

```bash
python3 validate.py http://localhost:8000
```
