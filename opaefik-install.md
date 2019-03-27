# Opaefik: do it yourself

## Introduction
XXX

### Architecture
XXX

### Relevant system information

XXX let's show some info: ip, distribution and os
```
borja@dev-ubuntu:~$ hostname -i
192.168.249.25
```
```
borja@dev-ubuntu:~$ lsb_release -a
No LSB modules are available.
Distributor ID: Ubuntu
Description:    Ubuntu 18.04.2 LTS
Release:        18.04
Codename:       bionic
```
```
borja@dev-ubuntu:~$ uname -a
Linux dev-ubuntu 4.15.0-46-generic #49-Ubuntu SMP Wed Feb 6 09:33:07 UTC 2019 x86_64 x86_64 x86_64 GNU/Linux
```

## Components installation

XXX installation begins

### Installing Docker
XXX

Doc: https://docs.docker.com/install/linux/docker-ce/ubuntu/

Commands:

//Remove any previous installation
```
sudo apt-get remove docker docker-engine docker.io containerd runc
```

//Update apt
```
sudo apt-get update
```

//Install packages to allow apt to use a repository over HTTPS and some utilities
```
sudo apt-get install apt-transport-https ca-certificates curl gnupg-agent software-properties-common
```

//Add Docker repo to apt
```
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
```

//Configure Docker stable repository
```
sudo add-apt-repository \
   "deb [arch=amd64] https://download.docker.com/linux/ubuntu \
   $(lsb_release -cs) \
   stable"
```

//Do actual docker installation
```
sudo apt-get update
```

//Install updated versions of "Docker CE" and "containerd"
```
sudo apt-get install docker-ce docker-ce-cli containerd.io
```

//Verify installation with a hello world
```
sudo docker run hello-world
```

//You should see a message saying something like 
```
Hello from Docker!
This message shows that your installation appears to be working correctly.
```


   
### Installing Kubernetes

XXX Why have we choosen minikube

#### Install Minikube
Doc:
- https://kubernetes.io/docs/tasks/tools/install-minikube/
- https://kubernetes.io/docs/setup/minikube/


Commands:

//If curl is not installed
```
sudo apt install curl
```

//Download minikube
```
curl -Lo minikube https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64 && chmod +x minikube
```
//Put minikube into path
```
sudo cp minikube /usr/local/bin && rm minikube
```

//Start minikube without any additional virtualization. Use docker.
```
sudo minikube --vm-driver=none start
```

#### Install Kubectl
Doc: https://kubernetes.io/docs/tasks/tools/install-kubectl/

Commands:
//Enable apt via HTTPS
```
sudo apt-get update && sudo apt-get install -y apt-transport-https
```
//Get and trust on kubernetes apt repository key
```
curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key add -
```
//Add Kubernetes repository to apt
```
echo "deb https://apt.kubernetes.io/ kubernetes-xenial main" | sudo tee -a /etc/apt/sources.list.d/kubernetes.list
```
//Update installable components
```
sudo apt-get update
```
//Install Kubectl
```
sudo apt-get install -y kubectl
```

//Show configuration (verify clusters > cluster > server URL matches with machine IP address)
```
sudo cat ~/.kube/config
```

XXX KEEP ON VALIDATING FROM HERE


### Install Traefik to be used as ingress

Doc: https://github.com/containous/traefik/blob/master/old/docs/user-guide/kubernetes.md

//Deploy traefik as a daemon (runs on every server) with DEBUG loglevel
sudo kubectl apply -f traefik/traefik_daemon_debug.yaml

### Install OPA for policy evaluation

Doc: https://github.com/open-policy-agent/opa/blob/master/docs/book/deployments.md


//Define the authorization policy in a configmap
sudo kubectl create configmap example-policy --from-file opa/auth_policy.rego

//Deploy OPA
sudo kubectl apply -f opa/opa_deployment.yaml

//Define OPA service. It is published internally, but not in the node's IP
sudo kubectl apply -f opa/opa_service.yaml

//As service opa published just internally (inside the cluster), we need to know where is has been put by kubernetes: Cluster ip and port
sudo kubectl get service opa

//You can test the OPA deployment and policy configuration with the following request. Please be sure to set the IP address according to your environment. Either cluster IP and the mapped port, or the pod ip (cluster-ip value from former command) and pod port
curl -X POST -d @opa/user_json.json -H "Content-Type: application/json" 192.168.249.25:31756

### Install Mysql for additional user information

Doc: https://kubernetes.io/docs/tasks/run-application/run-single-instance-stateful-application/

//Create persistent volume and volumen claim
sudo kubectl create -f sql/mysql-pv.yaml

//Create deployment and service
sudo kubectl create -f sql/mysql-deployment.yaml

//Make mysql accessible in the node (to simplify access from opaefik)
sudo kubectl apply -f sql/mysql-service-updated.yaml

XXX REVISANDO AQUÍ

//Add user information to mysql
sudo kubectl run -it --rm --image=mysql:5.6 --restart=Never mysql-client -- mysql -h mysql -ppassword < sql/user-information.sql

### Build Opaefik into docker

XXX include golang instalation, download of sql driver, compilation of opaefik, put into kubernetes opaefik deployment and service (opaefik_kube_system.yaml) and validate.

### Creating a simple service to protect

//Go to the directory
cd dune

//Show contents of the "service". It is na apache httpd that returns some content
printf "************************\n*** LISTING CONTENTS ***\n************************\n";find .;printf "************************\n*** SHOWING CONTENTS ***\n************************\n";for f in `find .`; do printf "***\nFile: $f\n" ;cat $f; done

//Show docker configuration
cat Dockerfile

//Build the docker image
docker build -t dune-stuff:1.1 .

//Show deployment information. Explain deployment, service and ingress
cat dune.yaml

//Deploy the service
kubectl apply -f dune/dune.yaml 

//Go back to your home
cd

## Verify environment
Users:

- thawat
- patreides
- vharkonen
- pdevries

URLs:

- /authenticated/any
- /groups/mentat
- /house/atreides
- /mac/vharkonen
- /public

Authorized users:

- /authenticated/any => ALL
- /groups/mentat => thawat | pdevries
- /house/atreides => thawat | patreides
- /mac/vharkonen => vharkonen
- /public => ALL (even without authentication header)


Commands:

- authenticated
 - curl dune.minikube/authenticated/any # KO
 - curl -H 'X-WebAuth-User:thawat' dune.minikube/authenticated/any
 - curl -H 'X-WebAuth-User:patreides' dune.minikube/authenticated/any
 - curl -H 'X-WebAuth-User:vharkonen' dune.minikube/authenticated/any
 - curl -H 'X-WebAuth-User:pdevries' dune.minikube/authenticated/any

- mentat
 - curl -H 'X-WebAuth-User:thawat' dune.minikube/group/mentat
 - curl -H 'X-WebAuth-User:patreides' dune.minikube/group/mentat # KO
 - curl -H 'X-WebAuth-User:vharkonen' dune.minikube/group/mentat # KO
 - curl -H 'X-WebAuth-User:pdevries' dune.minikube/group/mentat

- atreides
 - curl -H 'X-WebAuth-User:thawat' dune.minikube/house/atreides
 - curl -H 'X-WebAuth-User:patreides' dune.minikube/house/atreides
 - curl -H 'X-WebAuth-User:vharkonen' dune.minikube/house/atreides # KO
 - curl -H 'X-WebAuth-User:pdevries' dune.minikube/house/atreides # KO

- vharkonen
 - curl -H 'X-WebAuth-User:thawat' dune.minikube/mac/vharkonen # KO
 - curl -H 'X-WebAuth-User:patreides' dune.minikube/mac/vharkonen # KO
 - curl -H 'X-WebAuth-User:vharkonen' dune.minikube/mac/vharkonen
 - curl -H 'X-WebAuth-User:pdevries' dune.minikube/mac/vharkonen # KO

- public
 - curl dune.minikube/public
