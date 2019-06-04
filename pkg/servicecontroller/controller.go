package service

import (
	"k8s.io/ingress-gce/pkg/utils"
	"k8s.io/client-go/kubernetes"
	"k8s.io/ingress-gce/pkg/context"
	"sync"
	"k8s.io/client-go/tools/cache"
	"k8s.io/ingress-gce/pkg/controller"
	"k8s.io/ingress-gce/pkg/instances"
	"k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/cloudprovider/providers/gce"
	"github.com/golang/glog"
	"fmt"
	"reflect"
)

type ILBController struct {
	ctx *context.ControllerContext
	client         kubernetes.Interface
	// hasSynced returns true if all associated sub-controllers have synced.
	// Abstracted into a func for testing.
	hasSynced func() bool
	nodes      *controller.NodeController
	nodeLister cache.Indexer
	resourceMgr *ResourceManager // Manages all the resources associated with an InternalLoadbalancer Service
	serviceCache // Map of all the relevant services we have so far
	serviceLister  cache.Indexer
	stopCh     chan struct{}
	// stopLock is used to enforce only a single call to Stop is active.
	// Needed because we allow stopping through an http endpoint and
	// allowing concurrent stoppers leads to stack traces.
	stopLock sync.Mutex
	svcQueue utils.TaskQueue
}

func wantsILB(service *v1.Service) (bool, string) {
	if service == nil {
		return false, ""
	}
	// TODO check this same name in the cache, see if that spec was using ILB. We need to check because a service type
	// could be modified from type LoadBalancer to something else or from Internal to External Load Balancer.

	if service.Spec.Type != v1.ServiceTypeLoadBalancer {
		return false, fmt.Sprintf("Type : %s", service.Spec.Type)
	}
	ltype, _ := gce.GetLoadBalancerAnnotationType(service)
	if ltype == gce.LBTypeInternal {
		return true, fmt.Sprintf("Type : %s, %s", service.Spec.Type, ltype)
	}
	return false, fmt.Sprintf("Type : %s, %s", service.Spec.Type, gce.LBTypeInternal, ltype)
}

func NewController(ctx *context.ControllerContext, stopCh chan struct{}) *ILBController {
	instancePool := instances.NewNodePool(ctx.Cloud, ctx.ClusterNamer)
	ilbc := &ILBController {
		ctx: ctx,
		client: ctx.KubeClient,
		nodes: controller.NewNodeController(ctx, instancePool),
		nodeLister: ctx.NodeInformer.GetIndexer(),
		serviceLister: ctx.ServiceInformer.GetIndexer(),
		}
	ilbc.svcQueue = utils.NewPeriodicTaskQueue("ilb", "services", ilbc.sync)
	ctx.ServiceInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			addSvc := obj.(*v1.Service)
			svcKey := utils.ServiceKeyFunc(addSvc.Namespace, addSvc.Name)
			if result, out := wantsILB(addSvc); !result {
				glog.V(4).Infof("Ignoring add for non-lb service %s based on %v", svcKey, out)
				return
			}
			glog.V(3).Infof("Service %s added, enqueuing", svcKey)
			ilbc.ctx.Recorder(addSvc.Namespace).Eventf(addSvc, v1.EventTypeNormal, "ADD", svcKey)
			ilbc.svcQueue.Enqueue(obj)
		},

		DeleteFunc: func(obj interface{}) {
			delSvc := obj.(*v1.Service)
			svcKey := utils.ServiceKeyFunc(delSvc.Namespace, delSvc.Name)
			if result, out := wantsILB(delSvc); !result {
				glog.V(4).Infof("Ignoring delete for service %v based on %v", svcKey, out)
				return
			}

			glog.V(3).Infof("Service %s deleted, enqueueing", svcKey)
			ilbc.svcQueue.Enqueue(obj)
		},
		UpdateFunc: func(old, cur interface{}) {
			curSvc := cur.(*v1.Service)
			svcKey := utils.ServiceKeyFunc(curSvc.Namespace, curSvc.Name)
			if result, newtype := wantsILB(curSvc); !result {
				oldSvc := old.(*v1.Service)
				// If old service created ILB, we need to delete GCP LB resources.
				if result, oldtype := wantsILB(oldSvc); result {
					glog.V(4).Infof("Service %s type was changed from %s, enqueuing", svcKey, oldtype, newtype)
					ilbc.svcQueue.Enqueue(cur)
					return
				}
				return
			}
			if !reflect.DeepEqual(old, cur) {
				glog.V(3).Infof("Periodic enqueueing of %v", svcKey)
			} else {
				glog.V(3).Infof("Service %v changed, enqueuing", svcKey)
			}
			ilbc.svcQueue.Enqueue(cur)
		},

	})
	ctx.AddHealthCheck("ServiceController healthy", ilbc.IsHealthy)
	return ilbc
}

func (ilbc *ILBController) sync(key string) error {
	// Check whether this service is being added/deleted or modified. Can we have an explicit function to handle each
	// event instead of inferring it here?

	// Lookup the service name provided in key.
	svc, exists, err := ilbc.ctx.Services().GetByKey(key)
	if err != nil {
		return fmt.Errorf("Failed to lookup service for key %s : %s", key, err)
	}
	if !exists {
		glog.V(2).Infof("Service %s does not exist, cleaning up resources", key)
		//TODO TRIGGER CLEANUP
		return nil
	}
	result, out := wantsILB(svc)
	if result {
		glog.V(2).Infof("Added service %s of type %s, creating resources", key, out)
		//TODO create resources
	} else {
		glog.V(2).Infof("Added service %s of type %s, deleting ILB resources, if any", key, out)
	}
	return nil
}

func (ilbc *ILBController) IsHealthy() error {
	return fmt.Errorf("Testing")
}