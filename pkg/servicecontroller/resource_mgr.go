package service

import (
	"k8s.io/kubernetes/pkg/cloudprovider/providers/gce"
	"k8s.io/ingress-gce/pkg/context"
	"k8s.io/ingress-gce/pkg/utils"
)

type ResourceManager struct {
	cloud *gce.GCECloud
	ctx *context.ControllerContext

}

func (rm *ResourceManager) CleanupILBResources(svcKey string) {
	// Given the service key, lookup the name of the ilb, forwarding rules, backend service as well as name of instance group
	// Delete each one of them
		svcNs, svcName := utils.SplitServiceKey(svcKey)
		lbName := rm.ctx.ClusterNamer.ILBBackend()
		fwRule := rm.ctx.ClusterNamer.ForwardingRule()


	}

func (rm *ResourceManager) EnsureILBResources(svcKey string) {
	// Same as above, find names of all of them and create each type of resource. If it exists, keep going or delete first?
}
