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
echo "---> [0] OK 🍺"

echo "[1] Stop dockers"
eval $(docker-machine env ${app}) && docker-compose stop
echo "---> [1] OK 🍺"

echo "[2] Remove machines"
docker-machine rm ${app}
docker-machine rm ${worker}
echo "---> [2] OK 🍺"

echo "🍺 🍺 🍺 You've got everything removed."
