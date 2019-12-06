# RabbitMQ Operator

RabbitMQ Operator 构建于 [Operator framework](https://github.com/operator-framework/operator-sdk) 之上，用于在 [Kubernetes](https://kubernetes.io/) 上快速搭建[RabbitMQ](https://www.rabbitmq.com)集群。

# Features

- 自动创建RabbitMQ集群
- 自动开启RabbitMQ Management Web UI
- 自动创建K8S集群内部的LB Service
- 自动创建RabbitMQ Management Web UI的 Ingress 服务

# TODO

- 加入监控系统并与Prometheus
- 自动启用RabbitMQ延迟队列插件
- 集群扩容/缩容支持

 # Others
 
 https://www.rabbitmq.com/cluster-formation.html#peer-discovery-k8s
 
 https://github.com/rabbitmq/rabbitmq-peer-discovery-k8s
