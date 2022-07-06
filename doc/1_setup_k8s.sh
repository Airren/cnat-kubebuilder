#!/bin/bash -e

#---------------------------- Parse Argument -----------------------------------

sudo swapoff -a
intel_http_proxy='http://child-prc.intel.com:913'
proxy_flag='off'
k8s_version='1.24.2'
docker_version='20.10.17'
docker_user=
docker_pass=
api_server_ip=
while getopts ":a:f:v:u:p:h" options; do
  case "$options" in
  "a")
    api_server_ip=$OPTARG
    echo ">>>>>>>>>>>>>> api server ip ${api_server_ip}"
    ;;
  "f")
    echo ">>>>>>>>>>>>>> proxy off"
    proxy_flag=$OPTARG
    ;;
  "v")
    k8s_version=$OPTARG
    echo ">>>>>>>>>>>>>> specific K8s version to ${k8s_version}"
    ;;
  "u")
    docker_user=-u$OPTARG
    echo ">>>>>>>>>>>>>> specific docker user to ${docker_user}"
    ;;
  "p")
    docker_pass=-p$OPTARG
    echo ">>>>>>>>>>>>>> specific docker pass ok"
    ;;
  "h")
    echo '
      -f [off/on] proxy off 
      -v specify the k8s version
      -h help
    '
    exit 1
    ;;
  ":")
    echo 'unknown options $OPTARG'
    exit 1
    ;;
  *)
    echo "unknown error while processing options"
    exit 1
    ;;
  esac
done

if [ $proxy_flag == 'off' ]; then
  echo "WARNING: Suggest use -f option to set proxy for apt & docker"
fi

set -x

#-------------------- Set proxy for APT & Docker -------------------------------
function set_proxy_for_apt() {
  cat <<EOF | sudo tee /etc/apt/apt.conf.d/proxy.conf
Acquire::http::Proxy "$intel_http_proxy";
Acquire::https::Proxy "$intel_http_proxy";
EOF
  echo ">>>>>>>>>>>>>>>>>>>>>>> Set Proxy for Apt <<<<<<<<<<<<<<<<<<<<<<<<<<<<<"
}
if [ $proxy_flag == 'on' ]; then
  set_proxy_for_apt
fi

function set_proxy_for_docker() {
  sudo mkdir -p /etc/systemd/system/docker.service.d
  cat <<EOF | sudo tee /etc/systemd/system/docker.service.d/http-proxy.conf
[Service]
Environment="HTTP_PROXY=$intel_http_proxy"
Environment="HTTPS_PROXY=$intel_http_proxy"
Environment="NO_PROXY=localhost,127.0.0.0/8,10.293.154.0/16"
EOF
  sudo systemctl daemon-reload
  sudo systemctl restart docker
  echo ">>>>>>>>>>>>>>>>>>>>>>> Set Proxy for Docker <<<<<<<<<<<<<<<<<<<<<<<<<<"
}

#-------------------------------------------------------------------------------
#                          Pre Configure of K8s
#-------------------------------------------------------------------------------
#----------------- Let iptables see bridged traffic ----------------------------
function iptables_conf() {
  cat <<EOF | sudo tee /etc/modules-load.d/k8s.conf
br_netfilter
EOF

  cat <<EOF | sudo tee /etc/sysctl.d/k8s.conf
net.bridge.bridge-nf-call-ip6tables = 1
net.bridge.bridge-nf-call-iptables = 1
EOF

  sudo sysctl --system
  sudo modprobe br_netfilter
  # if the after command has response mean availavle
  lsmod | grep br_netfilter
}

#---------------------- Install Container Runtime ------------------------------
function docker_install() {
   sudo apt remove docker docker-engine docker.io containerd runc
 sudo apt-get remove -y docker-ce=5:${docker_version}~3-0~ubuntu-jammy \
  docker-ce-cli=5:${docker_version}~3-0~ubuntu-jammy containerd.io
   sudo apt install -y docker.io
  sudo rm -f /usr/share/keyrings/docker-archive-keyring.gpg
  curl  \
    -fsSL https://download.docker.com/linux/ubuntu/gpg |
    sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg

  echo \
    "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/\
docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu \
  $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list \
    >/dev/null

  sudo apt update

 #sudo apt-get install -y docker-ce=5:${docker_version}~3-0~ubuntu-jammy \
 # docker-ce-cli=5:${docker_version}~3-0~ubuntu-jammy containerd.io
 # sudo apt install docker.io


 # sudo apt-get install -y docker-ce=5:${docker_version}~3-0~ubuntu-focal \
  #  docker-ce-cli=5:${docker_version}~3-0~ubuntu-focal containerd.io
  sudo usermod -aG docker $USER

  # runtime config
  sudo mkdir -p /etc/docker
  cat <<EOF | sudo tee /etc/docker/daemon.json
{
  "exec-opts": ["native.cgroupdriver=systemd"],
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "100m"
  },
  "storage-driver": "overlay2"
}
EOF
  sudo systemctl enable docker
  sudo systemctl daemon-reload
  sleep 15s
  sudo systemctl restart docker

  if [ $proxy_flag == "on" ]; then
    set_proxy_for_docker
  fi

  if [ ! -z $docker_user ]; then
    sudo docker login ${docker_user} ${docker_pass}
  fi
}

#------------------------- Load Local Images -----------------------------------
function load_local_images() {
  mkdir -p /home/ubuntu/mnt/images
  pushd /home/ubuntu/mnt/images/
  wget -r -N -nd http://sdewan.sh.intel.com:8888/nodus-v6.0.0
  find /home/ubuntu/mnt/images/ -name "*.tar" | xargs -n1 sudo docker load -i
  popd
  echo ">>>>>>>>>>>>>>>> Load Local Docker Images <<<<<<<<<<<<<<<<<<<<<<<<<<<<<"
}

function kube_install() {
  #sudo apt-get update
  sudo apt-get install -y apt-transport-https ca-certificates curl \
    net-tools ipvsadm

  if [ $proxy_flag == 'on' ]; then
    sudo curl -x $intel_http_proxy \
      -fsSLo /usr/share/keyrings/kubernetes-archive-keyring.gpg \
      https://packages.cloud.google.com/apt/doc/apt-key.gpg

  else
    sudo curl -fsSLo /usr/share/keyrings/kubernetes-archive-keyring.gpg \
      https://packages.cloud.google.com/apt/doc/apt-key.gpg
  fi

  echo "deb [signed-by=/usr/share/keyrings/kubernetes-archive-keyring.gpg] \
  https://apt.kubernetes.io/ kubernetes-xenial main" |
    sudo tee /etc/apt/sources.list.d/kubernetes.list

  sudo apt-get update
  sudo apt-get install -y \
    kubelet=${k8s_version}-00 kubeadm=${k8s_version}-00 \
    kubectl=${k8s_version}-00
  sudo apt-mark hold kubelet kubeadm kubectl

  if [ ! -z $docker_user ]; then
    # temp solution to avoid docker pull from docker hub reach limit issue
    sudo cp /root/.docker/config.json /var/lib/kubelet/
  fi
}

# --------------------------- K8s Cluster Init ---------------------------------
function k8s_init() {
  ifconfig | awk '/inet/{print $2}' | cut -f2 -d ":"
  ipaddr=$(ifconfig | awk '/inet/{print $2}' | cut -f2 -d ":" |
    awk 'NR==2 {print $1}')

  echo "auto detected IP for api server:" $ipaddr

  if [ ! -z ${api_server_ip} ]; then
    echo "Use user specfied api server IP ${api_server_ip}"
    ipaddr=${api_server_ip}
  fi
  echo "---------------K8s API Ip is $ipaddr------------------------"
  echo "---------------K8s API Ip is $ipaddr------------------------"
  echo "---------------K8s API Ip is $ipaddr------------------------"

  sudo kubeadm init --kubernetes-version=${k8s_version} \
    --pod-network-cidr=10.233.64.0/18 --apiserver-advertise-address=${ipaddr} --cri-socket=/var/run/cri-dockerd.sock

  mkdir -p $HOME/.kube
  sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
  sudo chown $(id -u):$(id -g) $HOME/.kube/config
  # let master also be a work node
  kubectl taint nodes --all node-role.kubernetes.io/master-
  nodename=$(kubectl get node -o jsonpath='{.items[0].metadata.name}')
  # kubectl taint node $nodename node-role.kubernetes.io/master:NoSchedule-
  kubectl label --overwrite node $nodename ovn4nfv-k8s-plugin=ovn-control-plane
  kubectl create namespace sdewan-system

  # kubectl autocomplete
  cat <<EOF | sudo tee ~/.alias
source <(kubectl completion bash)
alias k="kubectl"
complete -o default -F __start_kubectl k
EOF
  echo 'source ~/.alias' >>~/.bashrc

  # helm install
  sudo snap install helm --classic
  sudo apt install -y jq

  sleep 30s
}

# apt non interactive
export DEBIAN_FRONTEND=noninteractive

iptables_conf
docker_install
# load_local_images
kube_install
k8s_init
