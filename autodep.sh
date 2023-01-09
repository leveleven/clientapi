#! /bin/bash

apt install lshw jq smartmontools -y

cp clientapi /usr/local/bin/
cp clientapi.service /lib/systemd/system/

systemctl daemon-reload

systemctl enable clientapi
systemctl start clientapi
