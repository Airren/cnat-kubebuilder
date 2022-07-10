#!/bin/bash
set -ex

pushd /home/ubuntu/mnt/deploy_conf
#kubectl apply -f tigera-operator.yaml
#kubectl apply -f calico_3.21.4.yaml
kubectl apply -f ./calico/tigera-operator.yaml
sleep 10s
kubectl apply -f ./calico/custom-resources.yaml
sleep 60s 

kubectl apply -f multus-daemonset.yml

sleep 30

kubectl apply -f cert-manager-1.8.0.yaml
sleep 30
# mv /etc/cni/net.d/70-multus.conf /etc/cni/net.d/00-multus.conf
kubectl apply -f ovn-daemonset.yaml
kubectl apply -f ovn4nfv-k8s-plugin.yaml

# kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v1.6.1/cert-manager.yaml
# kubectl apply -f cert-manager-1.6.1.yaml
# kubectl apply -f cert-manager-1.8.0.yaml
# kubectl apply -f cert-manager.yaml

popd

sleep 30

KUBE_EDITOR="sed -i s/\"natOutgoing: true\"/\"natOutgoing: false\"/g" kubectl edit ippools.crd.projectcalico.org default-ipv4-ippool
