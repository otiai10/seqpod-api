#!/bin/sh

app="seqpod-app"
worker="seqpod-worker"

function checkCLI() {
  if [ -z "$(which ${1})" ]; then
    echo "[ERROR] Command '${1}' not found"
    return 1
  fi
  echo "'${1}': OK"
  return 0
}

echo "[0] Check CLI tools"
checkCLI "docker-machine"
checkCLI "docker-compose"
checkCLI "VBoxManage"
echo "---> [0] OK 🍺"

echo "[1] Create machine for 'app' & 'mongodb'."
docker-machine create --driver virtualbox ${app}
docker-machine stop ${app}
echo "---> [1] OK 🍺"

echo "[2] Create machine for 'worker'"
docker-machine create --driver virtualbox ${worker}
docker-machine stop ${worker}
echo "---> [2] OK 🍺"

echo "[3] Mount directory to machines (equivalent to AWS EFS)"
VBoxManage sharedfolder add ${app} --name /var/app --hostpath ${pwd}/var/app --automount
VBoxManage sharedfolder add ${worker} --name /var/app --hostpath ${pwd}/var/app --automount
echo "---> [3] OK 🍺"

echo "[4] Start machines again"
docker-machine start ${app}
docker-machine start ${worker}
echo "---> [4] OK 🍺"

echo "🍺 🍺 🍺  Congrats! You've got everything done successfully."
printf "Do you want to start API server and MongoDB? (y/n): "
read answer
if [ "${answer}" == "y" ]; then
  eval $(docker-machine $app) && docker-compose up
else
  echo "OK, you can issue this command for next, Good luck!\n> eval $(docker-machine $app) && docker-compose up"
fi
