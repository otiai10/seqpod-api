#!/bin/sh

# This script is a setup script for local development,
# requiring
#   1. VBoxManage (VirtualBox)
#     - https://www.virtualbox.org/
#   2. docker-machine
#     - https://www.docker.com/products/docker-toolbox
#   3. docker-compose
#     - https://www.docker.com/products/docker-toolbox

app="seqpod-app"
worker="seqpod-worker"
cwd=`pwd`

function check_cli() {
  if [ -z "$(which ${1})" ]; then
    echo "[ERROR] Command '${1}' not found"
    return 1
  fi
  echo "'${1}': OK"
  return 0
}

index=0
function section_start() {
  echo "---> [${index}] ${1}"
}

function section_end() {
  echo "---> [${index}] 🍺  DONE 🍺"
  echo ""
  index=`expr $index + 1`
}

section_start "Check required CLI"
check_cli "docker-machine"
check_cli "docker-compose"
check_cli "VBoxManage"
section_end

section_start "Delete machines if exist"
docker-machine rm -f ${app}
docker-machine rm -f ${worker}
section_end

section_start "Create machine for 'api' & 'mongodb'."
docker-machine create --driver virtualbox ${app}
docker-machine stop ${app}
section_end

section_start "Create machine for 'worker'"
docker-machine create --driver virtualbox ${worker}
docker-machine stop ${worker}
section_end

section_start "Mount directory to machines (equivalent to AWS EFS)"
# VBoxManage sharedfolder remove ${app} --name /var/app
VBoxManage sharedfolder add ${app} --name /var/app --hostpath ${cwd}/var/app --automount
# VBoxManage sharedfolder remove ${app} --name /var/machine
VBoxManage sharedfolder add ${app} --name /var/machine -hostpath ${HOME}/.docker/machine/machines/${worker} --automount
# VBoxManage sharedfolder remove ${worker} --name /var/app
VBoxManage sharedfolder add ${worker} --name /var/app --hostpath ${cwd}/var/app --automount
section_end

section_start "Start machines again"
docker-machine start ${app}
docker-machine start ${worker}
section_end

section_start "Setup NAT proxy for ${app}"
VBoxManage controlvm ${app} natpf1 web,tcp,,8080,,8080
section_end

echo "🍺 🍺 🍺  Congrats! You've got everything done successfully."
printf "Do you want to start API server and MongoDB? (y/n): "
read answer
if [ "${answer}" == "y" ]; then
  eval $(docker-machine env $app) && docker-compose up --build -d
  echo "http://localhost:8080/v0/status"
else
  echo "OK, you can issue this command for next, Good luck!\n> eval \$(docker-machine env ${app}) && docker-compose up"
fi
