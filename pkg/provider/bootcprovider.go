package bootcprovider

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	messagebroker "bootcVK/messageBroker"
	utils "bootcVK/utils"

	dto "github.com/prometheus/client_model/go"
	"github.com/virtual-kubelet/virtual-kubelet/node/api"
	statsv1alpha1 "github.com/virtual-kubelet/virtual-kubelet/node/api/statsv1alpha1"
	"github.com/virtual-kubelet/virtual-kubelet/node/nodeutil"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE BootcProvider could connect to the DB or event bus or service that the IoT devices register to
// Must implement https://pkg.go.dev/github.com/virtual-kubelet/virtual-kubelet/node/nodeutil#BootcProvider
type BootcProvider struct {
	nodeName  string
	msgBroker *messagebroker.MessageBroker
}

func NewBootcProvider(nodeName string, cfg nodeutil.ProviderConfig) (*BootcProvider, error) {
	p := BootcProvider{}
	p.nodeName = nodeName

	user := os.Getenv("BROKER_USERNAME")
	password := os.Getenv("BROKER_PASSWORD")
	host := os.Getenv("BROKER_HOST")
	port := os.Getenv("BROKER_PORT")

	msgBroker, err := messagebroker.RegisterBroker(user, password, host, port)
	if err != nil {
		return nil, err
	}

	p.msgBroker = msgBroker
	p.ConfigureNode(nodeName, cfg.Node)

	return &p, nil
}

// All error for BootcProvider functions implemented need to be caught using github.com/virtual-kubelet/virtual-kubelet/errdefs

func (p *BootcProvider) CreatePod(ctx context.Context, pod *v1.Pod) error {
	fmt.Println("Called created fn")
	image := pod.Spec.Containers[0].Image

	err := utils.ValidateImage(image)
	if err != nil {
		// TODO doesnt seem to exit properly - create fn keeps being called
		// set error status for pod when do a get po
		return err
	}
	fmt.Println("Valid image")

	// NOTE use the rabbitmq rpc request and reply paradigm to wait for a response then update the pod spec
	err = p.msgBroker.SendSwitchCommand(image)
	if err != nil {
		return err
	}
	fmt.Println("Successfully sent message to create bootc application")

	now := metav1.NewTime(time.Now())
	pod.Status = v1.PodStatus{
		Phase:     v1.PodRunning,
		HostIP:    "1.2.3.4",
		PodIP:     "5.6.7.8",
		StartTime: &now,
		Conditions: []v1.PodCondition{
			{
				Type:   v1.PodInitialized,
				Status: v1.ConditionTrue,
			},
			{
				Type:   v1.PodReady,
				Status: v1.ConditionTrue,
			},
			{
				Type:   v1.PodScheduled,
				Status: v1.ConditionTrue,
			},
		},
	}

	return nil
}

func (p *BootcProvider) UpdatePod(ctx context.Context, pod *v1.Pod) error {
	err := utils.ValidateImage(pod.Spec.Containers[0].Image)
	if err != nil {
		return err
	}
	fmt.Println("Called update fn")
	return nil
}

func (p *BootcProvider) DeletePod(ctx context.Context, pod *v1.Pod) error {
	fmt.Println("Called delete fn")
	return nil
}

func (p *BootcProvider) GetPod(ctx context.Context, namespace, name string) (*v1.Pod, error) {
	fmt.Println("Called get pod fn")
	return nil, nil
}

func (p *BootcProvider) GetPodStatus(ctx context.Context, namespace, name string) (*v1.PodStatus, error) {
	return &v1.PodStatus{}, nil
}

func (p *BootcProvider) GetPods(context.Context) ([]*v1.Pod, error) {
	fmt.Println("Called get pods fn")
	return nil, nil
}

func (p *BootcProvider) GetContainerLogs(ctx context.Context, namespace, podName, containerName string, opts api.ContainerLogOpts) (io.ReadCloser, error) {
	return nil, nil
}

func (p *BootcProvider) RunInContainer(ctx context.Context, namespace, podName, containerName string, cmd []string, attach api.AttachIO) error {
	return nil
}

func (p *BootcProvider) AttachToContainer(ctx context.Context, namespace, podName, containerName string, attach api.AttachIO) error {
	return nil
}

func (p *BootcProvider) PortForward(ctx context.Context, namespace, pod string, port int32, stream io.ReadWriteCloser) error {
	return nil
}

func (p *BootcProvider) GetStatsSummary(context.Context) (*statsv1alpha1.Summary, error) {
	return &statsv1alpha1.Summary{}, nil
}

func (p *BootcProvider) GetMetricsResource(context.Context) ([]*dto.MetricFamily, error) {
	return []*dto.MetricFamily{}, nil
}
