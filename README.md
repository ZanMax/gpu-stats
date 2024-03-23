# GPU stats
Go web service which return json data with info about GPU usage

<img src="https://github.com/ZanMax/zanmax.github.io/blob/master/public/img/gpustats.png?raw=true" width="481" height="470" alt="GPU stats">

## Installation
```bash
git clone https://github.com/ZanMax/gpu-stats.git
```

## Build
```bash
cd gpu-stats
```
```bash
go build .
```

## Usage
### Server
```bash
./gpu-stats
```
### Web client 
- gpu-stats.html

## Run as a service

- edit gpustats.service
- copy gpustats.service to /etc/systemd/system/
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