#!/bin/bash

if [ "$#" -ne 2 ]; then
	echo "run this script with usernames:"
	echo "    ./add_user.sh <user_name> <github_id>"
	echo "example:"
	echo "    ./add_user.sh aray machinaut"
	exit
fi

user=$1
github=$2
echo "adding user $user github $github"
set -x
sudo groupadd -f wheel
sudo userdel $user
sudo rm -rf /home/$user
sudo useradd $user
sudo usermod -aG docker $user
sudo usermod -aG wheel $user
sudo -u $user ssh-keygen -P '' -t rsa -b 4096 -f /home/$user/.ssh/id_rsa
sudo -u $user curl https://github.com/$github.keys -o /home/$user/.ssh/authorized_keys
sudo -u $user chmod 644 /home/$user/.ssh/authorized_keys
