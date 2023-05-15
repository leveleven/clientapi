#! /bin/bash

apt update && apt install lshw jq smartmontools -y

echo -e "auto eth0\niface eth0 inet dhcp" > /etc/network/interfaces
sed -i 's/managed=false/managed=true/g' /etc/NetworkManager/NetworkManager.conf

/etc/init.d/network-manager restart

chmod +x clientapi
cp clientapi /usr/local/bin/
cp clientapi.service /lib/systemd/system/

systemctl daemon-reload

systemctl enable clientapi
systemctl start clientapi
