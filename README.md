# RabbitMQ Operator

RabbitMQ Operator 构建于 [Operator framework](https://github.com/operator-framework/operator-sdk) 之上，用于在 [Kubernetes](https://kubernetes.io/) 上快速搭建[RabbitMQ](https://www.rabbitmq.com)集群。

# Features

- 自动创建RabbitMQ集群
- 自动开启RabbitMQ Management Web UI
- 自动创建K8S集群内部的LB Service
- 自动创建RabbitMQ Management Web UI的 Ingress 服务

# Usage

```
# build your own operator image
git clone https://github.com/k8s-operators/rabbitmq-operator.git
cd rabbitmq-operator
operator-sdk generate k8s
operator-sdk build your-docker-name/rabbitmq-operator:v1.0.0
docker push your-docker-name/rabbitmq-operator:v1.0.0

# replace deploy/operator.yaml to use your operator image
# or just use jianzhiunique/rabbitmq-operator:v1.0.0

# deploy files under "deploy" dir to server
kubectl apply -f service_account.yaml
kubectl apply -f role.yaml
kubectl apply -f role_binding.yaml
kubectl apply -f operator.yaml
kubectl apply -f deploy/crds/lesolise.github.io_rabbitmqs_crd.yaml

# config cr.yaml
# for all fields, see pkg/apis/lesolise/v1/rabbitmq_types.go
# we apply some default values for you, see pkg/utils/check_cr.go

# the only config that you must specify is storage_class_name, 
# for rabbitmq, we recommend users to use local pv

```

# TODO

- 加入监控系统并与Prometheus
- 自动启用RabbitMQ延迟队列插件
- 集群扩容/缩容支持

 # Others
 
 https://www.rabbitmq.com/cluster-formation.html#peer-discovery-k8s
 
 https://github.com/rabbitmq/rabbitmq-peer-discovery-k8s
