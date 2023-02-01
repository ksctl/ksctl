#!/bin/bash
sudo apt update -y
sudo apt install haproxy -y

systemctl start haproxy
systemctl enable haproxy