#!/bin/bash

echo "[TASK 1] Update /etc/hosts file"
cat >>/etc/hosts<<EOF
192.168.56.100   core.funless.dev    core  
192.168.56.101   worker1.funless.dev worker1
192.168.56.102   worker2.funless.dev worker2
EOF

# echo "[TASK 2] Stop and Disable firewall"
# systemctl disable --now ufw >/dev/null 2>&1