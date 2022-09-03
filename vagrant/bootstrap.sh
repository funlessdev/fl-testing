#!/bin/bash

echo "[TASK 1] Update /etc/hosts file"
cat >>/etc/hosts<<EOF
192.168.56.100   core.funless.dev    core  
192.168.56.101   worker1.funless.dev worker1
192.168.56.102   worker2.funless.dev worker2
EOF

# echo "[TASK 2] Stop and Disable firewall"
# systemctl disable --now ufw >/dev/null 2>&1

echo "[TASK 2] Install Docker"
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER
newgrp docker

echo "[TASK 3] Install go"
wget https://go.dev/dl/go1.19.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.19.linux-amd64.tar.gz
echo "export PATH=$PATH:/usr/local/go/bin" >> /etc/profile
source /etc/profile

echo "[TASK 4] Install task"
sudo snap install task --classic

echo "[TASK 5] Clone fl-cli"
git clone https://github.com/funlessdev/fl-cli.git

echo "[TASK 6] Install fl-cli"
cd fl-cli
task install
