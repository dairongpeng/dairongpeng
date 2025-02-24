# Win10端口转发WSL Ubuntu SSH 22端口
Win10管理员权限下，直接运行，更进一步可以配置开机启动。

```shell
# setup_ssh_forwarding.ps1

$wsl_ip = (wsl hostname -I).Split(" ")[0]
netsh interface portproxy add v4tov4 listenport=2222 listenaddress=0.0.0.0 connectport=22 connectaddress=$wsl_ip
New-NetFireWallRule -Name "SSH_WSL" -Direction Inbound -LocalPort 2222 -Protocol TCP -Action Allow -ErrorAction SilentlyContinue
```
