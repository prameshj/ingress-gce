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

package syncers

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud"
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/meta"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/ingress-gce/pkg/composite"
	negtypes "k8s.io/ingress-gce/pkg/neg/types"
	"k8s.io/legacy-cloud-providers/gce"
)

func TestEncodeDecodeEndpoint(t *testing.T) {
	ip := "10.0.0.10"
	instance := "somehost"
	port := "8080"

	retIp, retInstance, retPort := decodeEndpoint(encodeEndpoint(ip, instance, port))

	if ip != retIp || instance != retInstance || retPort != port {
		t.Fatalf("Encode and decode endpoint failed. Expect %q, %q, %q but got %q, %q, %q.", ip, instance, port, retIp, retInstance, retPort)
	}
}

func TestCalculateDifference(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		targetSet  map[string]sets.String
		currentSet map[string]sets.String
		addSet     map[string]sets.String
		removeSet  map[string]sets.String
	}{
		// unchanged
		{
			targetSet: map[string]sets.String{
				negtypes.TestZone1: sets.NewString("a", "b", "c"),
			},
			currentSet: map[string]sets.String{
				negtypes.TestZone1: sets.NewString("a", "b", "c"),
			},
			addSet:    map[string]sets.String{},
			removeSet: map[string]sets.String{},
		},
		// unchanged
		{
			targetSet:  map[string]sets.String{},
			currentSet: map[string]sets.String{},
			addSet:     map[string]sets.String{},
			removeSet:  map[string]sets.String{},
		},
		// add in one zone
		{
			targetSet: map[string]sets.String{
				negtypes.TestZone1: sets.NewString("a", "b", "c"),
			},
			currentSet: map[string]sets.String{},
			addSet: map[string]sets.String{
				negtypes.TestZone1: sets.NewString("a", "b", "c"),
			},
			removeSet: map[string]sets.String{},
		},
		// add in 2 zones
		{
			targetSet: map[string]sets.String{
				negtypes.TestZone1: sets.NewString("a", "b", "c"),
				negtypes.TestZone2: sets.NewString("e", "f", "g"),
			},
			currentSet: map[string]sets.String{},
			addSet: map[string]sets.String{
				negtypes.TestZone1: sets.NewString("a", "b", "c"),
				negtypes.TestZone2: sets.NewString("e", "f", "g"),
			},
			removeSet: map[string]sets.String{},
		},
		// remove in one zone
		{
			targetSet: map[string]sets.String{},
			currentSet: map[string]sets.String{
				negtypes.TestZone1: sets.NewString("a", "b", "c"),
			},
			addSet: map[string]sets.String{},
			removeSet: map[string]sets.String{
				negtypes.TestZone1: sets.NewString("a", "b", "c"),
			},
		},
		// remove in 2 zones
		{
			targetSet: map[string]sets.String{},
			currentSet: map[string]sets.String{
				negtypes.TestZone1: sets.NewString("a", "b", "c"),
				negtypes.TestZone2: sets.NewString("e", "f", "g"),
			},
			addSet: map[string]sets.String{},
			removeSet: map[string]sets.String{
				negtypes.TestZone1: sets.NewString("a", "b", "c"),
				negtypes.TestZone2: sets.NewString("e", "f", "g"),
			},
		},
		// add and delete in one zone
		{
			targetSet: map[string]sets.String{
				negtypes.TestZone1: sets.NewString("a", "b", "c"),
			},
			currentSet: map[string]sets.String{
				negtypes.TestZone1: sets.NewString("b", "c", "d"),
			},
			addSet: map[string]sets.String{
				negtypes.TestZone1: sets.NewString("a"),
			},
			removeSet: map[string]sets.String{
				negtypes.TestZone1: sets.NewString("d"),
			},
		},
		// add and delete in 2 zones
		{
			targetSet: map[string]sets.String{
				negtypes.TestZone1: sets.NewString("a", "b", "c"),
				negtypes.TestZone2: sets.NewString("a", "b", "c"),
			},
			currentSet: map[string]sets.String{
				negtypes.TestZone1: sets.NewString("b", "c", "d"),
				negtypes.TestZone2: sets.NewString("b", "c", "d"),
			},
			addSet: map[string]sets.String{
				negtypes.TestZone1: sets.NewString("a"),
				negtypes.TestZone2: sets.NewString("a"),
			},
			removeSet: map[string]sets.String{
				negtypes.TestZone1: sets.NewString("d"),
				negtypes.TestZone2: sets.NewString("d"),
			},
		},
	}

	for _, tc := range testCases {
		addSet, removeSet := calculateDifference(tc.targetSet, tc.currentSet)

		if !reflect.DeepEqual(addSet, tc.addSet) {
			t.Errorf("Failed to calculate difference for add, expecting %v, but got %v", tc.addSet, addSet)
		}

		if !reflect.DeepEqual(removeSet, tc.removeSet) {
			t.Errorf("Failed to calculate difference for remove, expecting %v, but got %v", tc.removeSet, removeSet)
		}
	}
}

func TestNetworkEndpointCalculateDifference(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		targetSet  map[string]negtypes.NetworkEndpointSet
		currentSet map[string]negtypes.NetworkEndpointSet
		addSet     map[string]negtypes.NetworkEndpointSet
		removeSet  map[string]negtypes.NetworkEndpointSet
	}{
		// unchanged
		{
			targetSet: map[string]negtypes.NetworkEndpointSet{
				negtypes.TestZone1: negtypes.NewNetworkEndpointSet(genNetworkEndpoint("a"), genNetworkEndpoint("b"), genNetworkEndpoint("c")),
			},
			currentSet: map[string]negtypes.NetworkEndpointSet{
				negtypes.TestZone1: negtypes.NewNetworkEndpointSet(genNetworkEndpoint("a"), genNetworkEndpoint("b"), genNetworkEndpoint("c")),
			},
			addSet:    map[string]negtypes.NetworkEndpointSet{},
			removeSet: map[string]negtypes.NetworkEndpointSet{},
		},
		// unchanged
		{
			targetSet:  map[string]negtypes.NetworkEndpointSet{},
			currentSet: map[string]negtypes.NetworkEndpointSet{},
			addSet:     map[string]negtypes.NetworkEndpointSet{},
			removeSet:  map[string]negtypes.NetworkEndpointSet{},
		},
		// add in one zone
		{
			targetSet: map[string]negtypes.NetworkEndpointSet{
				negtypes.TestZone1: negtypes.NewNetworkEndpointSet(genNetworkEndpoint("a"), genNetworkEndpoint("b"), genNetworkEndpoint("c")),
			},
			currentSet: map[string]negtypes.NetworkEndpointSet{},
			addSet: map[string]negtypes.NetworkEndpointSet{
				negtypes.TestZone1: negtypes.NewNetworkEndpointSet(genNetworkEndpoint("a"), genNetworkEndpoint("b"), genNetworkEndpoint("c")),
			},
			removeSet: map[string]negtypes.NetworkEndpointSet{},
		},
		// add in 2 zones
		{
			targetSet: map[string]negtypes.NetworkEndpointSet{
				negtypes.TestZone1: negtypes.NewNetworkEndpointSet(genNetworkEndpoint("a"), genNetworkEndpoint("b"), genNetworkEndpoint("c")),
				negtypes.TestZone2: negtypes.NewNetworkEndpointSet(genNetworkEndpoint("e"), genNetworkEndpoint("f"), genNetworkEndpoint("g")),
			},
			currentSet: map[string]negtypes.NetworkEndpointSet{},
			addSet: map[string]negtypes.NetworkEndpointSet{
				negtypes.TestZone1: negtypes.NewNetworkEndpointSet(genNetworkEndpoint("a"), genNetworkEndpoint("b"), genNetworkEndpoint("c")),
				negtypes.TestZone2: negtypes.NewNetworkEndpointSet(genNetworkEndpoint("e"), genNetworkEndpoint("f"), genNetworkEndpoint("g")),
			},
			removeSet: map[string]negtypes.NetworkEndpointSet{},
		},
		// remove in one zone
		{
			targetSet: map[string]negtypes.NetworkEndpointSet{},
			currentSet: map[string]negtypes.NetworkEndpointSet{
				negtypes.TestZone1: negtypes.NewNetworkEndpointSet(genNetworkEndpoint("a"), genNetworkEndpoint("b"), genNetworkEndpoint("c")),
			},
			addSet: map[string]negtypes.NetworkEndpointSet{},
			removeSet: map[string]negtypes.NetworkEndpointSet{
				negtypes.TestZone1: negtypes.NewNetworkEndpointSet(genNetworkEndpoint("a"), genNetworkEndpoint("b"), genNetworkEndpoint("c")),
			},
		},
		// remove in 2 zones
		{
			targetSet: map[string]negtypes.NetworkEndpointSet{},
			currentSet: map[string]negtypes.NetworkEndpointSet{
				negtypes.TestZone1: negtypes.NewNetworkEndpointSet(genNetworkEndpoint("a"), genNetworkEndpoint("b"), genNetworkEndpoint("c")),
				negtypes.TestZone2: negtypes.NewNetworkEndpointSet(genNetworkEndpoint("e"), genNetworkEndpoint("f"), genNetworkEndpoint("g")),
			},
			addSet: map[string]negtypes.NetworkEndpointSet{},
			removeSet: map[string]negtypes.NetworkEndpointSet{
				negtypes.TestZone1: negtypes.NewNetworkEndpointSet(genNetworkEndpoint("a"), genNetworkEndpoint("b"), genNetworkEndpoint("c")),
				negtypes.TestZone2: negtypes.NewNetworkEndpointSet(genNetworkEndpoint("e"), genNetworkEndpoint("f"), genNetworkEndpoint("g")),
			},
		},
		// add and delete in one zone
		{
			targetSet: map[string]negtypes.NetworkEndpointSet{
				negtypes.TestZone1: negtypes.NewNetworkEndpointSet(genNetworkEndpoint("a"), genNetworkEndpoint("b"), genNetworkEndpoint("c")),
			},
			currentSet: map[string]negtypes.NetworkEndpointSet{
				negtypes.TestZone1: negtypes.NewNetworkEndpointSet(genNetworkEndpoint("b"), genNetworkEndpoint("c"), genNetworkEndpoint("d")),
			},
			addSet: map[string]negtypes.NetworkEndpointSet{
				negtypes.TestZone1: negtypes.NewNetworkEndpointSet(genNetworkEndpoint("a")),
			},
			removeSet: map[string]negtypes.NetworkEndpointSet{
				negtypes.TestZone1: negtypes.NewNetworkEndpointSet(genNetworkEndpoint("d")),
			},
		},
		// add and delete in 2 zones
		{
			targetSet: map[string]negtypes.NetworkEndpointSet{
				negtypes.TestZone1: negtypes.NewNetworkEndpointSet(genNetworkEndpoint("a"), genNetworkEndpoint("b"), genNetworkEndpoint("c")),
				negtypes.TestZone2: negtypes.NewNetworkEndpointSet(genNetworkEndpoint("a"), genNetworkEndpoint("b"), genNetworkEndpoint("c")),
			},
			currentSet: map[string]negtypes.NetworkEndpointSet{
				negtypes.TestZone1: negtypes.NewNetworkEndpointSet(genNetworkEndpoint("b"), genNetworkEndpoint("c"), genNetworkEndpoint("d")),
				negtypes.TestZone2: negtypes.NewNetworkEndpointSet(genNetworkEndpoint("b"), genNetworkEndpoint("c"), genNetworkEndpoint("d")),
			},
			addSet: map[string]negtypes.NetworkEndpointSet{
				negtypes.TestZone1: negtypes.NewNetworkEndpointSet(genNetworkEndpoint("a")),
				negtypes.TestZone2: negtypes.NewNetworkEndpointSet(genNetworkEndpoint("a")),
			},
			removeSet: map[string]negtypes.NetworkEndpointSet{
				negtypes.TestZone1: negtypes.NewNetworkEndpointSet(genNetworkEndpoint("d")),
				negtypes.TestZone2: negtypes.NewNetworkEndpointSet(genNetworkEndpoint("d")),
			},
		},
	}

	for _, tc := range testCases {
		addSet, removeSet := calculateNetworkEndpointDifference(tc.targetSet, tc.currentSet)

		if !reflect.DeepEqual(addSet, tc.addSet) {
			t.Errorf("Failed to calculate difference for add, expecting %v, but got %v", tc.addSet, addSet)
		}

		if !reflect.DeepEqual(removeSet, tc.removeSet) {
			t.Errorf("Failed to calculate difference for remove, expecting %v, but got %v", tc.removeSet, removeSet)
		}
	}
}

func TestEnsureNetworkEndpointGroup(t *testing.T) {
	var (
		testZone             = "test-zone"
		testNamedPort        = "named-port"
		testServiceName      = "test-svc"
		testServiceNameSpace = "test-ns"
		testNetwork          = cloud.ResourcePath("network", &meta.Key{Zone: testZone, Name: "test-network"})
		testSubnetwork       = cloud.ResourcePath("subnetwork", &meta.Key{Zone: testZone, Name: "test-subnetwork"})
	)

	fakeCloud := negtypes.NewFakeNetworkEndpointGroupCloud(testSubnetwork, testNetwork)

	testCases := []struct {
		description         string
		negName             string
		enableNonGCPMode    bool
		networkEndpointType negtypes.NetworkEndpointType
		expectedSubnetwork  string
		apiVersion          meta.Version
	}{
		{
			description:         "Create NEG of type GCE_VM_IP_PORT",
			negName:             "gcp-neg",
			enableNonGCPMode:    false,
			networkEndpointType: negtypes.VmIpPortEndpointType,
			expectedSubnetwork:  testSubnetwork,
			apiVersion:          meta.VersionGA,
		},
		{
			description:         "Create NEG of type NON_GCP_PRIVATE_IP_PORT",
			negName:             "non-gcp-neg",
			enableNonGCPMode:    true,
			networkEndpointType: negtypes.NonGCPPrivateEndpointType,
			expectedSubnetwork:  "",
			apiVersion:          meta.VersionGA,
		},
		{
			description:         "Create NEG of type GCE_VM_IP",
			negName:             "gcp-ip-neg",
			enableNonGCPMode:    false,
			networkEndpointType: negtypes.VmIpEndpointType,
			expectedSubnetwork:  testSubnetwork,
			apiVersion:          meta.VersionAlpha,
		},
	}
	for _, tc := range testCases {
		ensureNetworkEndpointGroup(
			testServiceNameSpace,
			testServiceName,
			tc.negName,
			testZone,
			testNamedPort,
			tc.networkEndpointType,
			fakeCloud,
			nil,
			nil,
			tc.apiVersion,
		)

		neg, err := fakeCloud.GetNetworkEndpointGroup(tc.negName, testZone, tc.apiVersion)
		if err != nil {
			t.Errorf("Failed to retrieve NEG %q: %v", tc.negName, err)
		}

		if neg.NetworkEndpointType != string(tc.networkEndpointType) {
			t.Errorf("Unexpected NetworkEndpointType, expecting %q but got %q", tc.networkEndpointType, neg.NetworkEndpointType)
		}

		if neg.Subnetwork != tc.expectedSubnetwork {
			t.Errorf("Unexpected Subnetwork, expecting %q but got %q", tc.expectedSubnetwork, neg.Subnetwork)
		}

		// Call ensureNetworkEndpointGroup with the same NEG.
		err = ensureNetworkEndpointGroup(
			testServiceNameSpace,
			testServiceName,
			tc.negName,
			testZone,
			testNamedPort,
			tc.networkEndpointType,
			fakeCloud,
			nil,
			nil,
			tc.apiVersion,
		)

		if err != nil {
			t.Errorf("Unexpected error when called with duplicated NEG: %v", err)
		}
	}
}

func TestToZoneNetworkEndpointMapUtil(t *testing.T) {
	t.Parallel()
	_, transactionSyncer := newTestTransactionSyncer(negtypes.NewAdapter(gce.NewFakeGCECloud(gce.DefaultTestClusterValues())), negtypes.VmIpPortEndpointType)
	podLister := transactionSyncer.podLister

	// add all pods in default endpoint into podLister
	for i := 1; i <= 12; i++ {
		podLister.Add(&v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: testServiceNamespace,
				Name:      fmt.Sprintf("pod%v", i),
			},
		})
	}
	// pod6 is deleted
	podLister.Update(&v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:         testServiceNamespace,
			Name:              "pod6",
			DeletionTimestamp: &metav1.Time{},
		},
	})

	podLister.Update(&v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:         testServiceNamespace,
			Name:              "pod12",
			DeletionTimestamp: &metav1.Time{},
		},
	})

	zoneGetter := negtypes.NewFakeZoneGetter()
	testCases := []struct {
		desc                string
		portName            string
		endpointSets        map[string]negtypes.NetworkEndpointSet
		expectMap           negtypes.EndpointPodMap
		networkEndpointType negtypes.NetworkEndpointType
	}{
		{
			desc:                "non exist target port",
			portName:            "non-exists",
			endpointSets:        map[string]negtypes.NetworkEndpointSet{},
			expectMap:           negtypes.EndpointPodMap{},
			networkEndpointType: negtypes.VmIpPortEndpointType,
		},
		{
			desc:     "target port number",
			portName: "",
			endpointSets: map[string]negtypes.NetworkEndpointSet{
				negtypes.TestZone1: negtypes.NewNetworkEndpointSet(
					networkEndpointFromEncodedEndpoint("10.100.1.1||instance1||80"),
					networkEndpointFromEncodedEndpoint("10.100.1.2||instance1||80"),
					networkEndpointFromEncodedEndpoint("10.100.2.1||instance2||80"),
					networkEndpointFromEncodedEndpoint("10.100.1.3||instance1||80")),
				negtypes.TestZone2: negtypes.NewNetworkEndpointSet(
					networkEndpointFromEncodedEndpoint("10.100.3.1||instance3||80")),
			},
			expectMap: negtypes.EndpointPodMap{
				networkEndpointFromEncodedEndpoint("10.100.1.1||instance1||80"): types.NamespacedName{Namespace: testServiceNamespace, Name: "pod1"},
				networkEndpointFromEncodedEndpoint("10.100.1.2||instance1||80"): types.NamespacedName{Namespace: testServiceNamespace, Name: "pod2"},
				networkEndpointFromEncodedEndpoint("10.100.2.1||instance2||80"): types.NamespacedName{Namespace: testServiceNamespace, Name: "pod3"},
				networkEndpointFromEncodedEndpoint("10.100.3.1||instance3||80"): types.NamespacedName{Namespace: testServiceNamespace, Name: "pod4"},
				networkEndpointFromEncodedEndpoint("10.100.1.3||instance1||80"): types.NamespacedName{Namespace: testServiceNamespace, Name: "pod5"},
			},
			networkEndpointType: negtypes.VmIpPortEndpointType,
		},
		{
			desc:     "named target port",
			portName: testNamedPort,
			endpointSets: map[string]negtypes.NetworkEndpointSet{
				negtypes.TestZone1: negtypes.NewNetworkEndpointSet(
					networkEndpointFromEncodedEndpoint("10.100.2.2||instance2||81")),
				negtypes.TestZone2: negtypes.NewNetworkEndpointSet(
					networkEndpointFromEncodedEndpoint("10.100.4.1||instance4||81"),
					networkEndpointFromEncodedEndpoint("10.100.3.2||instance3||8081"),
					networkEndpointFromEncodedEndpoint("10.100.4.2||instance4||8081"),
					networkEndpointFromEncodedEndpoint("10.100.4.3||instance4||81")),
			},
			expectMap: negtypes.EndpointPodMap{
				networkEndpointFromEncodedEndpoint("10.100.2.2||instance2||81"):   types.NamespacedName{Namespace: testServiceNamespace, Name: "pod7"},
				networkEndpointFromEncodedEndpoint("10.100.4.1||instance4||81"):   types.NamespacedName{Namespace: testServiceNamespace, Name: "pod8"},
				networkEndpointFromEncodedEndpoint("10.100.4.3||instance4||81"):   types.NamespacedName{Namespace: testServiceNamespace, Name: "pod9"},
				networkEndpointFromEncodedEndpoint("10.100.3.2||instance3||8081"): types.NamespacedName{Namespace: testServiceNamespace, Name: "pod10"},
				networkEndpointFromEncodedEndpoint("10.100.4.2||instance4||8081"): types.NamespacedName{Namespace: testServiceNamespace, Name: "pod11"},
			},
			networkEndpointType: negtypes.VmIpPortEndpointType,
		},
		{
			desc:     "Non-GCP network endpoints",
			portName: "",
			endpointSets: map[string]negtypes.NetworkEndpointSet{
				negtypes.TestZone1: negtypes.NewNetworkEndpointSet(
					networkEndpointFromEncodedEndpoint("10.100.1.1||||80"),
					networkEndpointFromEncodedEndpoint("10.100.1.2||||80"),
					networkEndpointFromEncodedEndpoint("10.100.2.1||||80"),
					networkEndpointFromEncodedEndpoint("10.100.1.3||||80")),
				negtypes.TestZone2: negtypes.NewNetworkEndpointSet(
					networkEndpointFromEncodedEndpoint("10.100.3.1||||80")),
			},
			expectMap: negtypes.EndpointPodMap{
				networkEndpointFromEncodedEndpoint("10.100.1.1||||80"): types.NamespacedName{Namespace: testServiceNamespace, Name: "pod1"},
				networkEndpointFromEncodedEndpoint("10.100.1.2||||80"): types.NamespacedName{Namespace: testServiceNamespace, Name: "pod2"},
				networkEndpointFromEncodedEndpoint("10.100.2.1||||80"): types.NamespacedName{Namespace: testServiceNamespace, Name: "pod3"},
				networkEndpointFromEncodedEndpoint("10.100.3.1||||80"): types.NamespacedName{Namespace: testServiceNamespace, Name: "pod4"},
				networkEndpointFromEncodedEndpoint("10.100.1.3||||80"): types.NamespacedName{Namespace: testServiceNamespace, Name: "pod5"},
			},
			networkEndpointType: negtypes.NonGCPPrivateEndpointType,
		},
	}

	for _, tc := range testCases {
		retSet, retMap, err := toZoneNetworkEndpointMap(getDefaultEndpoint(), zoneGetter, tc.portName, podLister, "", tc.networkEndpointType)
		if err != nil {
			t.Errorf("For case %q, expect nil error, but got %v.", tc.desc, err)
		}

		if !reflect.DeepEqual(retSet, tc.endpointSets) {
			t.Errorf("For case %q, expecting endpoint set %v, but got %v.", tc.desc, tc.endpointSets, retSet)
		}

		if !reflect.DeepEqual(retMap, tc.expectMap) {
			t.Errorf("For case %q, expecting endpoint map %v, but got %v.", tc.desc, tc.expectMap, retMap)
		}
	}
}

func TestRetrieveExistingZoneNetworkEndpointMap(t *testing.T) {
	zoneGetter := negtypes.NewFakeZoneGetter()
	negCloud := negtypes.NewFakeNetworkEndpointGroupCloud("test-subnetwork", "test-newtork")
	negName := "test-neg-name"
	irrelevantNegName := "irrelevant"
	testIP1 := "1.2.3.4"
	testIP2 := "1.2.3.5"
	testIP3 := "1.2.3.6"
	testIP4 := "1.2.3.7"
	testPort := int64(80)

	testCases := []struct {
		desc      string
		mutate    func(cloud negtypes.NetworkEndpointGroupCloud)
		expect    map[string]negtypes.NetworkEndpointSet
		expectErr bool
	}{
		{
			desc:      "neg not exists",
			mutate:    func(cloud negtypes.NetworkEndpointGroupCloud) {},
			expectErr: true,
		},
		{
			desc: "neg only exists in one of the zone",
			mutate: func(cloud negtypes.NetworkEndpointGroupCloud) {
				cloud.CreateNetworkEndpointGroup(&composite.NetworkEndpointGroup{Name: testNegName, Version: meta.VersionGA}, negtypes.TestZone1)
			},
			expectErr: true,
		},
		{
			desc: "neg only exists in one of the zone plus irrelevant negs",
			mutate: func(cloud negtypes.NetworkEndpointGroupCloud) {
				cloud.CreateNetworkEndpointGroup(&composite.NetworkEndpointGroup{Name: irrelevantNegName, Version: meta.VersionGA}, negtypes.TestZone2)
			},
			expectErr: true,
		},
		{
			desc: "empty negs exists in both zones",
			mutate: func(cloud negtypes.NetworkEndpointGroupCloud) {
				cloud.CreateNetworkEndpointGroup(&composite.NetworkEndpointGroup{Name: testNegName, Version: meta.VersionGA}, negtypes.TestZone2)
			},
			expect: map[string]negtypes.NetworkEndpointSet{
				negtypes.TestZone1: negtypes.NewNetworkEndpointSet(),
				negtypes.TestZone2: negtypes.NewNetworkEndpointSet(),
			},
			expectErr: false,
		},
		{
			desc: "one empty and one non-empty negs",
			mutate: func(cloud negtypes.NetworkEndpointGroupCloud) {
				cloud.AttachNetworkEndpoints(testNegName, negtypes.TestZone1, []*composite.NetworkEndpoint{
					{
						Instance:  negtypes.TestInstance1,
						IpAddress: testIP1,
						Port:      testPort,
					},
				}, meta.VersionGA)
			},
			expect: map[string]negtypes.NetworkEndpointSet{
				negtypes.TestZone1: negtypes.NewNetworkEndpointSet(negtypes.NetworkEndpoint{IP: testIP1, Node: negtypes.TestInstance1, Port: strconv.Itoa(int(testPort))}),
				negtypes.TestZone2: negtypes.NewNetworkEndpointSet(),
			},
			expectErr: false,
		},
		{
			desc: "one neg with multiple endpoints",
			mutate: func(cloud negtypes.NetworkEndpointGroupCloud) {
				cloud.AttachNetworkEndpoints(testNegName, negtypes.TestZone1, []*composite.NetworkEndpoint{
					{
						Instance:  negtypes.TestInstance2,
						IpAddress: testIP2,
						Port:      testPort,
					},
				}, meta.VersionGA)
			},
			expect: map[string]negtypes.NetworkEndpointSet{
				negtypes.TestZone1: negtypes.NewNetworkEndpointSet(
					negtypes.NetworkEndpoint{IP: testIP1, Node: negtypes.TestInstance1, Port: strconv.Itoa(int(testPort))},
					negtypes.NetworkEndpoint{IP: testIP2, Node: negtypes.TestInstance2, Port: strconv.Itoa(int(testPort))},
				),
				negtypes.TestZone2: negtypes.NewNetworkEndpointSet(),
			},
			expectErr: false,
		},
		{
			desc: "both negs with multiple endpoints",
			mutate: func(cloud negtypes.NetworkEndpointGroupCloud) {
				cloud.AttachNetworkEndpoints(testNegName, negtypes.TestZone2, []*composite.NetworkEndpoint{
					{
						Instance:  negtypes.TestInstance3,
						IpAddress: testIP3,
						Port:      testPort,
					},
					{
						Instance:  negtypes.TestInstance4,
						IpAddress: testIP4,
						Port:      testPort,
					},
				}, meta.VersionGA)
			},
			expect: map[string]negtypes.NetworkEndpointSet{
				negtypes.TestZone1: negtypes.NewNetworkEndpointSet(
					negtypes.NetworkEndpoint{IP: testIP1, Node: negtypes.TestInstance1, Port: strconv.Itoa(int(testPort))},
					negtypes.NetworkEndpoint{IP: testIP2, Node: negtypes.TestInstance2, Port: strconv.Itoa(int(testPort))},
				),
				negtypes.TestZone2: negtypes.NewNetworkEndpointSet(
					negtypes.NetworkEndpoint{IP: testIP3, Node: negtypes.TestInstance3, Port: strconv.Itoa(int(testPort))},
					negtypes.NetworkEndpoint{IP: testIP4, Node: negtypes.TestInstance4, Port: strconv.Itoa(int(testPort))},
				),
			},
			expectErr: false,
		},
		{
			desc: "irrelevant neg",
			mutate: func(cloud negtypes.NetworkEndpointGroupCloud) {
				cloud.AttachNetworkEndpoints(irrelevantNegName, negtypes.TestZone2, []*composite.NetworkEndpoint{
					{
						Instance:  negtypes.TestInstance3,
						IpAddress: testIP4,
						Port:      testPort,
					},
				}, meta.VersionGA)
			},
			expect: map[string]negtypes.NetworkEndpointSet{
				negtypes.TestZone1: negtypes.NewNetworkEndpointSet(
					negtypes.NetworkEndpoint{IP: testIP1, Node: negtypes.TestInstance1, Port: strconv.Itoa(int(testPort))},
					negtypes.NetworkEndpoint{IP: testIP2, Node: negtypes.TestInstance2, Port: strconv.Itoa(int(testPort))},
				),
				negtypes.TestZone2: negtypes.NewNetworkEndpointSet(
					negtypes.NetworkEndpoint{IP: testIP3, Node: negtypes.TestInstance3, Port: strconv.Itoa(int(testPort))},
					negtypes.NetworkEndpoint{IP: testIP4, Node: negtypes.TestInstance4, Port: strconv.Itoa(int(testPort))},
				),
			},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		tc.mutate(negCloud)
		out, err := retrieveExistingZoneNetworkEndpointMap(negName, zoneGetter, negCloud, meta.VersionGA)

		if tc.expectErr {
			if err == nil {
				t.Errorf("For test case %q, expecting error, but got nil", tc.desc)
			}
		} else {
			if err != nil {
				t.Errorf("For test case %q, endpointSets err = nil, but got %v", tc.desc, err)
			}
		}

		if !tc.expectErr {
			if !reflect.DeepEqual(out, tc.expect) {
				t.Errorf("For test case %q, endpointSets output = %+v, but got %+v", tc.desc, tc.expect, out)
			}
		}
	}
}

func TestMakeEndpointBatch(t *testing.T) {
	testCases := []struct {
		desc        string
		endpointNum int
		leftOverNum int
		negType     negtypes.NetworkEndpointType
	}{
		{
			"input with zero endpoints",
			0,
			0,
			negtypes.VmIpPortEndpointType,
		},
		{
			"input with 1 endpoints",
			1,
			0,
			negtypes.VmIpPortEndpointType,
		},
		{
			"input with 500 endpoints",
			500,
			0,
			negtypes.VmIpPortEndpointType,
		},
		{
			"input with 501 endpoints",
			501,
			1,
			negtypes.VmIpPortEndpointType,
		},
		{
			"input with 1000 endpoints",
			1000,
			500,
			negtypes.VmIpPortEndpointType,
		},
		{
			"input with 1000 endpoints",
			1000,
			500,
			negtypes.VmIpPortEndpointType,
		},
	}

	for _, tc := range testCases {
		endpointSet, endpointMap := genTestEndpoints(tc.endpointNum)
		out, err := makeEndpointBatch(endpointSet, tc.negType)

		if err != nil {
			t.Errorf("Expect err = nil, but got %v", err)
		}

		if endpointSet.Len() != tc.leftOverNum {
			t.Errorf("Expect endpoint set has %d endpoints left, but got %d", tc.leftOverNum, endpointSet.Len())
		}

		expectOutputEndpoints := tc.endpointNum
		if tc.endpointNum > MAX_NETWORK_ENDPOINTS_PER_BATCH {
			expectOutputEndpoints = MAX_NETWORK_ENDPOINTS_PER_BATCH
		}

		if expectOutputEndpoints != len(out) {
			t.Errorf("Expect %d endpoint(s) in output, but got %d", expectOutputEndpoints, len(out))
		}

		for key, endpoint := range out {
			if endpointSet.Has(key) {
				t.Errorf("Expect %q endpoint to exist in output endpoint map, but not", key)
			}
			expectEndpoint, ok := endpointMap[key]
			if !ok {
				t.Errorf("Expect %q endpoint to exist in expected endpoint map, but not", key)
			} else {
				if !reflect.DeepEqual(expectEndpoint, endpoint) {
					t.Errorf("Expect endpoint object %+v, but got %+v", expectEndpoint, endpoint)
				}
			}
		}
	}
}

func TestShouldPodBeInNeg(t *testing.T) {
	t.Parallel()

	_, transactionSyncer := newTestTransactionSyncer(negtypes.NewAdapter(gce.NewFakeGCECloud(gce.DefaultTestClusterValues())), negtypes.VmIpPortEndpointType)

	podLister := transactionSyncer.podLister

	namespace1 := "ns1"
	namespace2 := "ns2"
	name1 := "n1"
	name2 := "n2"

	podLister.Add(&v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace1,
			Name:      name1,
		},
	})

	// deleted pod
	podLister.Add(&v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:         namespace1,
			Name:              name2,
			DeletionTimestamp: &metav1.Time{},
		},
	})

	podLister.Add(&v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace2,
			Name:      name2,
		},
	})

	for _, tc := range []struct {
		desc      string
		namespace string
		name      string
		expect    bool
	}{
		{
			desc: "empty input",
		},
		{
			desc:      "non exists pod",
			namespace: "non exists",
			name:      "non exists",
			expect:    false,
		},
		{
			desc:      "pod exists and not deleted",
			namespace: namespace1,
			name:      name1,
			expect:    true,
		},
		{
			desc:      "pod exists and deleted",
			namespace: namespace1,
			name:      name2,
			expect:    false,
		},
		{
			desc:      "pod exists and not deleted 2",
			namespace: namespace2,
			name:      name2,
			expect:    true,
		},
	} {
		ret := shouldPodBeInNeg(podLister, tc.namespace, tc.name)
		if ret != tc.expect {
			t.Errorf("For test case %q, endpointSets output = %+v, but got %+v", tc.desc, tc.expect, ret)
		}
	}

}

func genTestEndpoints(num int) (negtypes.NetworkEndpointSet, map[negtypes.NetworkEndpoint]*composite.NetworkEndpoint) {
	endpointSet := negtypes.NewNetworkEndpointSet()
	endpointMap := map[negtypes.NetworkEndpoint]*composite.NetworkEndpoint{}
	ip := "1.2.3.4"
	instance := "instance"
	for port := 0; port < num; port++ {
		key := negtypes.NetworkEndpoint{IP: ip, Node: instance, Port: strconv.Itoa(port)}
		endpointSet.Insert(key)
		endpointMap[key] = &composite.NetworkEndpoint{
			IpAddress: ip,
			Instance:  instance,
			Port:      int64(port),
		}
	}
	return endpointSet, endpointMap
}

func networkEndpointFromEncodedEndpoint(encodedEndpoint string) negtypes.NetworkEndpoint {
	ip, node, port := decodeEndpoint(encodedEndpoint)
	return negtypes.NetworkEndpoint{IP: ip, Node: node, Port: port}
}

func getDefaultEndpoint() *v1.Endpoints {
	instance1 := negtypes.TestInstance1
	instance2 := negtypes.TestInstance2
	instance3 := negtypes.TestInstance3
	instance4 := negtypes.TestInstance4
	return &v1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testServiceName,
			Namespace: testServiceNamespace,
		},
		Subsets: []v1.EndpointSubset{
			{
				Addresses: []v1.EndpointAddress{
					{
						IP:       "10.100.1.1",
						NodeName: &instance1,
						TargetRef: &v1.ObjectReference{
							Namespace: testServiceNamespace,
							Name:      "pod1",
						},
					},
					{
						IP:       "10.100.1.2",
						NodeName: &instance1,
						TargetRef: &v1.ObjectReference{
							Namespace: testServiceNamespace,
							Name:      "pod2",
						},
					},
					{
						IP:       "10.100.2.1",
						NodeName: &instance2,
						TargetRef: &v1.ObjectReference{
							Namespace: testServiceNamespace,
							Name:      "pod3",
						},
					},
					{
						IP:       "10.100.3.1",
						NodeName: &instance3,
						TargetRef: &v1.ObjectReference{
							Namespace: testServiceNamespace,
							Name:      "pod4",
						},
					},
				},
				NotReadyAddresses: []v1.EndpointAddress{
					{
						IP:       "10.100.1.3",
						NodeName: &instance1,
						TargetRef: &v1.ObjectReference{
							Namespace: testServiceNamespace,
							Name:      "pod5",
						},
					},
					{
						IP:       "10.100.1.4",
						NodeName: &instance1,
						TargetRef: &v1.ObjectReference{
							Namespace: testServiceNamespace,
							Name:      "pod6",
						},
					},
				},
				Ports: []v1.EndpointPort{
					{
						Name:     "",
						Port:     int32(80),
						Protocol: v1.ProtocolTCP,
					},
				},
			},
			{
				Addresses: []v1.EndpointAddress{
					{
						IP:       "10.100.2.2",
						NodeName: &instance2,
						TargetRef: &v1.ObjectReference{
							Namespace: testServiceNamespace,
							Name:      "pod7",
						},
					},
					{
						IP:       "10.100.4.1",
						NodeName: &instance4,
						TargetRef: &v1.ObjectReference{
							Namespace: testServiceNamespace,
							Name:      "pod8",
						},
					},
				},
				NotReadyAddresses: []v1.EndpointAddress{
					{
						IP:       "10.100.4.3",
						NodeName: &instance4,
						TargetRef: &v1.ObjectReference{
							Namespace: testServiceNamespace,
							Name:      "pod9",
						},
					},
				},
				Ports: []v1.EndpointPort{
					{
						Name:     testNamedPort,
						Port:     int32(81),
						Protocol: v1.ProtocolTCP,
					},
				},
			},
			{
				Addresses: []v1.EndpointAddress{
					{
						IP:       "10.100.3.2",
						NodeName: &instance3,
						TargetRef: &v1.ObjectReference{
							Namespace: testServiceNamespace,
							Name:      "pod10",
						},
					},
					{
						IP:       "10.100.4.2",
						NodeName: &instance4,
						TargetRef: &v1.ObjectReference{
							Namespace: testServiceNamespace,
							Name:      "pod11",
						},
					},
				},
				NotReadyAddresses: []v1.EndpointAddress{
					{
						IP:       "10.100.4.4",
						NodeName: &instance4,
						TargetRef: &v1.ObjectReference{
							Namespace: testServiceNamespace,
							Name:      "pod12",
						},
					},
				},
				Ports: []v1.EndpointPort{
					{
						Name:     testNamedPort,
						Port:     int32(8081),
						Protocol: v1.ProtocolTCP,
					},
				},
			},
		},
	}
}
