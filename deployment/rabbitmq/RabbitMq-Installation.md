Install rabbitmq operator
kubectl apply -f "https://github.com/rabbitmq/cluster-operator/releases/latest/download/cluster-operator.yml"

Apply rabbitmqcluster.yaml

- If deploying on cluster running on managed cloud use LoadBalancer as service type

- Otherwise create node port service and create a route/ingress

Get credentials from fleet-manager-default-user secret in rabbitmq-system namespace