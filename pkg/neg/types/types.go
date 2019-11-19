/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package types

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/meta"

	istioV1alpha3 "istio.io/api/networking/v1alpha3"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/ingress-gce/pkg/annotations"
)

type NetworkEndpointType string

const (
	VmIpPortEndpointType      = NetworkEndpointType("GCE_VM_IP_PORT")
	VmPrimaryIpEndpointType   = NetworkEndpointType("GCE_VM_PRIMARY_IP")
	NonGCPPrivateEndpointType = NetworkEndpointType("NON_GCP_PRIVATE_IP_PORT")
)

// SvcPortTuple is the tuple representing one service port
type SvcPortTuple struct {
	// Port is the service port number
	Port int32
	// Name is the service port name
	Name string
	// TargetPort is the service target port.
	// This can be a port number or named port
	TargetPort string
}

// String returns the string representation of SvcPortTuple
func (t SvcPortTuple) String() string {
	return fmt.Sprintf("%s/%v-%s", t.Name, t.Port, t.TargetPort)
}

// SvcPortTupleSet is a set of SvcPortTuple
type SvcPortTupleSet map[SvcPortTuple]struct{}

// NewSvcPortTupleSet returns SvcPortTupleSet with the input tuples
func NewSvcPortTupleSet(tuples ...SvcPortTuple) SvcPortTupleSet {
	set := SvcPortTupleSet{}
	set.Insert(tuples...)
	return set
}

// Insert inserts a SvcPortTuple into SvcPortTupleSet
func (set SvcPortTupleSet) Insert(tuples ...SvcPortTuple) {
	for _, tuple := range tuples {
		set[tuple] = struct{}{}
	}
}

// Get returns the SvcPortTuple with matching svc port if found
func (set SvcPortTupleSet) Get(svcPort int32) (SvcPortTuple, bool) {
	for tuple := range set {
		if svcPort == tuple.Port {
			return tuple, true
		}
	}
	return SvcPortTuple{}, false
}

// PortInfo contains information associated with service port
type PortInfo struct {
	// PortTuple is port tuple of a service.
	PortTuple SvcPortTuple

	// Subset name, subset is defined in Istio:DestinationRule, this is only used
	// when --enable-csm=true.
	Subset string

	// Subset label, should set together with Subset.
	SubsetLabels string

	// NegName is the name of the NEG
	NegName string
	// ReadinessGate indicates if the NEG associated with the port has NEG readiness gate enabled
	// This is enabled with service port is reference by ingress.
	// If the service port is only exposed as stand alone NEG, it should not be enbled.
	ReadinessGate bool
}

// PortInfoMapKey is the Key of PortInfoMap
type PortInfoMapKey struct {
	// ServicePort is the service port
	ServicePort int32

	// Istio:DestinationRule Subset, only used when --enable-csm=true
	Subset string
}

// PortInfoMap is a map of PortInfoMapKey:PortInfo
type PortInfoMap map[PortInfoMapKey]PortInfo

func NewPortInfoMap(namespace, name string, svcPortTupleSet SvcPortTupleSet, namer NetworkEndpointGroupNamer, readinessGate bool) PortInfoMap {
	ret := PortInfoMap{}
	for svcPortTuple := range svcPortTupleSet {
		ret[PortInfoMapKey{svcPortTuple.Port, ""}] = PortInfo{
			PortTuple:     svcPortTuple,
			NegName:       namer.NEG(namespace, name, svcPortTuple.Port),
			ReadinessGate: readinessGate,
		}
	}
	return ret
}

// NewPortInfoMapWithDestinationRule create PortInfoMap based on a gaven DesinationRule.
// Return error message if the DestinationRule contains duplicated subsets.
func NewPortInfoMapWithDestinationRule(namespace, name string, svcPortTupleSet SvcPortTupleSet, namer NetworkEndpointGroupNamer, readinessGate bool,
	destinationRule *istioV1alpha3.DestinationRule) (PortInfoMap, error) {
	ret := PortInfoMap{}
	var duplicateSubset []string
	for _, subset := range destinationRule.Subsets {
		for tuple := range svcPortTupleSet {
			key := PortInfoMapKey{tuple.Port, subset.Name}
			if _, ok := ret[key]; ok {
				duplicateSubset = append(duplicateSubset, subset.Name)
			}
			ret[key] = PortInfo{
				PortTuple:     tuple,
				NegName:       namer.NEGWithSubset(namespace, name, subset.Name, tuple.Port),
				ReadinessGate: readinessGate,
				Subset:        subset.Name,
				SubsetLabels:  labels.Set(subset.Labels).String(),
			}
		}
	}
	if len(duplicateSubset) != 0 {
		return ret, fmt.Errorf("Duplicated subsets: %s", strings.Join(duplicateSubset, ", "))
	}
	return ret, nil
}

// Merge merges p2 into p1 PortInfoMap
// It assumes the same key (service port) will have the same target port and negName
// If not, it will throw error
// If a key in p1 or p2 has readiness gate enabled, the merged port info will also has readiness gate enabled
func (p1 PortInfoMap) Merge(p2 PortInfoMap) error {
	var err error
	for mapKey, portInfo := range p2 {
		mergedInfo := PortInfo{}
		if existingPortInfo, ok := p1[mapKey]; ok {
			if existingPortInfo.PortTuple != portInfo.PortTuple {
				return fmt.Errorf("for service port %v, port tuple in existing map is %q, but the merge map has %q", mapKey, existingPortInfo.PortTuple, portInfo.PortTuple)
			}
			if existingPortInfo.NegName != portInfo.NegName {
				return fmt.Errorf("for service port %v, NEG name in existing map is %q, but the merge map has %q", mapKey, existingPortInfo.NegName, portInfo.NegName)
			}
			if existingPortInfo.Subset != portInfo.Subset {
				return fmt.Errorf("for service port %v, Subset name in existing map is %q, but the merge map has %q", mapKey, existingPortInfo.Subset, portInfo.Subset)
			}
			mergedInfo.ReadinessGate = existingPortInfo.ReadinessGate
		}
		mergedInfo.PortTuple = portInfo.PortTuple
		mergedInfo.NegName = portInfo.NegName
		// Turn on the readiness gate if one of them is on
		mergedInfo.ReadinessGate = mergedInfo.ReadinessGate || portInfo.ReadinessGate
		mergedInfo.Subset = portInfo.Subset
		mergedInfo.SubsetLabels = portInfo.SubsetLabels

		p1[mapKey] = mergedInfo
	}
	return err
}

// Difference returns the entries of PortInfoMap which satisfies one of the following 2 conditions:
// 1. portInfo entry is not the present in p1
// 2. or the portInfo entry is not the same in p1 and p2.
func (p1 PortInfoMap) Difference(p2 PortInfoMap) PortInfoMap {
	result := make(PortInfoMap)
	for mapKey, p1PortInfo := range p1 {
		p2PortInfo, ok := p2[mapKey]
		if ok && reflect.DeepEqual(p1[mapKey], p2PortInfo) {
			continue
		}
		result[mapKey] = p1PortInfo
	}
	return result
}

func (p1 PortInfoMap) ToPortNegMap() annotations.PortNegMap {
	ret := annotations.PortNegMap{}
	for mapKey, portInfo := range p1 {
		ret[strconv.Itoa(int(mapKey.ServicePort))] = portInfo.NegName
	}
	return ret
}

func (p1 PortInfoMap) ToPortSubsetNegMap() annotations.PortSubsetNegMap {
	ret := annotations.PortSubsetNegMap{}
	for mapKey, portInfo := range p1 {
		if m, ok := ret[mapKey.Subset]; ok {
			m[strconv.Itoa(int(mapKey.ServicePort))] = portInfo.NegName
		} else {
			ret[mapKey.Subset] = map[string]string{strconv.Itoa(int(mapKey.ServicePort)): portInfo.NegName}
		}
	}
	return ret
}

// NegsWithReadinessGate returns the NegNames which has readiness gate enabled
func (p1 PortInfoMap) NegsWithReadinessGate() sets.String {
	ret := sets.NewString()
	for _, info := range p1 {
		if info.ReadinessGate {
			ret.Insert(info.NegName)
		}
	}
	return ret
}

// NegSyncerKey includes information to uniquely identify a NEG syncer
type NegSyncerKey struct {
	// Namespace of service
	Namespace string
	// Name of service
	Name string
	// PortTuple is the port tuple of the service backing the NEG
	PortTuple SvcPortTuple

	// Subset name, subset is defined in Istio:DestinationRule, this is only used
	// when --enable-csm=true.
	Subset string

	// Subset label, should set together with Subset.
	SubsetLabels string

	// NegType is the type of the network endpoints in this NEG.
	NegType NetworkEndpointType
	// Randomize indicates that the endpoints of the NEG can be picked at random, rather
	// than following the endpoints of the service. This only applies in the VM_PRIMARY_IP
	// NEG
	Randomize bool
}

func (key NegSyncerKey) String() string {
	return fmt.Sprintf("%s/%s-%s-%s", key.Namespace, key.Name, key.Subset, key.PortTuple.String())
}

// GetAPIVersion returns the compute API version to be used in order
// to create the negType specified in the given NegSyncerKey.
func (key NegSyncerKey) GetAPIVersion() meta.Version {
	switch key.NegType {
	case VmPrimaryIpEndpointType:
		return meta.VersionAlpha
	case NonGCPPrivateEndpointType:
		return meta.VersionAlpha
	default:
		return meta.VersionGA
	}
}

// EndpointPodMap is a map from network endpoint to a namespaced name of a pod
type EndpointPodMap map[NetworkEndpoint]types.NamespacedName
