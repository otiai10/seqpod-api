#!/bin/sh

err=0
if [ -z "${VPC_ID}" ]; then
    echo "[!] env var VPC_ID is required"
    err=`expr $err + 1`
fi
if [ -z "${SUBNET_ID}" ]; then
    echo "[!] env var SUBNET_ID is required"
    err=`expr $err + 1`
fi
if [ ${err} -gt 0 ]; then
    echo "--> ABORTED"
    exit 1
fi

docker-machine create \
--driver amazonec2 \
--amazonec2-region ap-southeast-2 \
--amazonec2-vpc-id ${VPC_ID} \
--amazonec2-subnet-id ${SUBNET_ID} \
--amazonec2-security-group seqpod-app \
--amazonec2-instance-type m4.xlarge \
--engine-install-url=https://releases.rancher.com/install-docker/1.13.sh \
aws-seqpod-app

docker-machine create \
--driver amazonec2 \
--amazonec2-region ap-southeast-2 \
--amazonec2-vpc-id ${VPC_ID} \
--amazonec2-subnet-id ${SUBNET_ID} \
--amazonec2-security-group seqpod-worker \
--amazonec2-instance-type m4.2xlarge \
--engine-install-url=https://releases.rancher.com/install-docker/1.13.sh \
aws-seqpod-worker

echo "--> FINISHED"
