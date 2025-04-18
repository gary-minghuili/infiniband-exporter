#!/usr/bin/env bash

ip_address='10.4.101.1'
user='admin'
password='Canopy@123456'

session_cookie=$(curl -i "https://${ip_address}/admin/launch?script=rh&template=login&action=login" \
  --data-raw "d_user_id=user_id&t_user_id=string&c_user_id=string&e_user_id=true&f_user_id=${user}&f_password=${password}&Login=" \
  --insecure | grep -o "session=[^;]*")

curl "https://${ip_address}/admin/launch?script=json" \
  -X POST \
  -H "Content-Type: application/json" \
  -b "${session_cookie}" \
  --data-raw '{
    "execution_type": "sync",
    "commands": ["file ibdiagnet upload ibdiagnet_output.tgz scp://test:test1234@10.4.254.250/home/test/ib.tgz"]
  }' \
  --insecure