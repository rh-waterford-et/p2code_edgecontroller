# About

> [!NOTE]
> This project is currently under development

Kuboot is a Kubernetes native scheduler capable of targeting edge devices leveraging bootable containers. Kuboot is a solution built to provide a controller for edge devices.  

> [P2CODE](https://p2code-project.eu/) has received funding from the European Union under grant agreement No. 101093069.

# Architecture

![Kuboot Architecture](/docs/images/kuboot-full-edge-controller-diagram.png "Kuboot Architecture")

# Installation

## Install Kuboot

1. Follow the [guide](deployment/rabbitmq/RabbitMq-Installation.md) to deploy a RabbitMQ instance 
2. Using the values from the secret generated, populate the environment variables in the [kuboot deployment](deployment/virtual-kubelet/virtual-kubelet.yaml)
3. Deploy Kuboot
   ``` kubectl apply -f deployment/virtual-kubelet ```
4. Ensure the kuboot node is deployed
    ``` kubectl get nodes ```
5. Ensure the kuboot deployment is healthy
    ``` kubectl get deployment -n virtual-kubelet ```

## Provision & Register Devices

1. Create amqp-config.yaml
``` mv agent/amqp-config.sample.yaml agent/amqp-config.yaml ```
2. Populate amqp-config.yaml with the values of the RabbitMQ cluster
3. Create login credentials for the device
``` mv agent/config.sample.toml config.toml ```
4. Build the image    
    - ``` sudo podman build -f agent/Dockerfile -t kuboot-agent . ```
    - ``` mkdir output ```
    - ``` sudo podman run --rm  -it --privileged --pull=newer --security-opt label=type:unconfined_t -v $(pwd)/config.toml:/config.toml:ro -v $(pwd)/output:/output -v /var/lib/containers/storage:/var/lib/containers/storage quay.io/centos-bootc/bootc-image-builder:latest --local --type raw --rootfs xfs kuboot-agent ```
5. Boot the edge device with the image generated
6. Log into the device with the credentials specified

# Demo

DevConf presentation and [demo](https://www.youtube.com/watch?v=6-x8tiwRO0s&list=PLU1vS0speL2bCGfDl6lCckdCzXb_gopJV&index=33)