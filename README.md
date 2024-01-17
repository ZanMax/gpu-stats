# GPU stats
Go web service which return json data with info about GPU usage

## Installation
```bash
go get github.com/ZanMax/gpu-stats
```

## Build
```bash
go build .
```

## Usage
```bash
./gpu-stats
```
## Run as a service

edit gpustats.service
```
Replace /path/to/your/binary with the full path to your compiled binary.
Set User and Group to appropriate values.
```

```bash
sudo systemctl daemon-reload
```

```bash
sudo systemctl start gpustats.service
```

enable your service to start on boot:
```bash
sudo systemctl enable gpustats.service
```

## Support
- Nvidia ( nvidia-smi )
- AMD ( rocm-smi )