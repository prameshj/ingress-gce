/*
Copyright 2019 The Kubernetes Authors.

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
	"crypto/sha256"
	"encoding/hex"
	"sort"

	"k8s.io/api/core/v1"
	negtypes "k8s.io/ingress-gce/pkg/neg/types"
	"k8s.io/ingress-gce/pkg/utils"
)

const (
	maxSubsetSize = 25
)

// NodeInfo stores node metadata used to sort nodes and pick a subset.
type NodeInfo struct {
	index      int
	hashedName string
	skip       bool
}

func getHashedName(nodeName, salt string) string {
	hashSum := sha256.Sum256([]byte(nodeName + ":" + salt))
	return hex.EncodeToString(hashSum[:])
}

// PickSubsets takes a list of nodes, hash salt, count and retuns a
// subset of size - 'count'. If the input list is smaller than the
// desired subset count, the entire list is returned. The hash salt
// is used so that a different subset is returned even when the same
// node list is passed in, for a different salt value.
func PickSubsets(nodes []*v1.Node, salt string, count int) []*v1.Node {
	if len(nodes) < count {
		return nodes
	}
	subsets := make([]*v1.Node, 0, count)
	info := make([]*NodeInfo, len(nodes))
	for i, node := range nodes {
		info[i] = &NodeInfo{i, getHashedName(node.Name, salt), false}
	}
	// sort alphabetically, based on the hashed string
	sort.Slice(info, func(i, j int) bool {
		return info[i].hashedName < info[j].hashedName
	})
	for _, val := range info {
		subsets = append(subsets, nodes[val.index])
		if len(subsets) == count {
			break
		}
	}
	return subsets
}

// PickSubsetsNoRemovals ensures that there are no node removals from current subset unless the node no longer exists.
func PickSubsetsNoRemovals(nodes []*v1.Node, salt string, count int, current []negtypes.NetworkEndpoint) []*v1.Node {
	if len(nodes) < count {
		return nodes
	}
	subset := make([]*v1.Node, 0, count)
	info := make([]*NodeInfo, len(nodes))
	for i, node := range nodes {
		info[i] = &NodeInfo{i, getHashedName(node.Name, salt), false}
	}
	// sort alphabetically, based on the hashed string
	sort.Slice(info, func(i, j int) bool {
		return info[i].hashedName < info[j].hashedName
	})
	// Pick all nodes from existing subset if still available.
	for _, ep := range current {
		for _, nodeInfo := range info {
			curHashName := getHashedName(ep.Node, salt)
			if nodeInfo.hashedName == curHashName {
				subset = append(subset, nodes[nodeInfo.index])
				nodeInfo.skip = true
			} else if nodeInfo.hashedName > curHashName {
				break
			}
		}
	}
	if len(subset) >= count {
		subset = subset[:count+1]
		return subset
	}
	for _, val := range info {
		if val.skip {
			continue
		}
		subset = append(subset, nodes[val.index])
		if len(subset) == count {
			break
		}
	}
	return subset
}
func getSubsetPerZone(nodes []*v1.Node, zoneGetter negtypes.ZoneGetter, svcID string, currentMap map[string]negtypes.NetworkEndpointSet) (map[string]negtypes.NetworkEndpointSet, error) {
	result := make(map[string]negtypes.NetworkEndpointSet)
	zoneMap := make(map[string][]*v1.Node)
	for _, node := range nodes {
		zone, err := zoneGetter.GetZoneForNode(node.Name)
		if err != nil {
			continue
		}
		zoneMap[zone] = append(zoneMap[zone], node)
	}
	// This algorithm picks atmost 'perZoneSubset' number of nodes from each zone.
	// If there are fewer nodes in one zone, more nodes are NOT picked from other zones.
	// TODO fix this.
	perZoneSubset := maxSubsetSize / len(zoneMap)
	var currentList []negtypes.NetworkEndpoint

	for zone, nodesInZone := range zoneMap {
		result[zone] = negtypes.NewNetworkEndpointSet()
		if zmap, ok := currentMap[zone]; ok {
			currentList = zmap.List()
		} else {
			currentList = nil
		}
		subset := PickSubsetsNoRemovals(nodesInZone, svcID, perZoneSubset, currentList)
		for _, node := range subset {
			result[zone].Insert(negtypes.NetworkEndpoint{Node: node.Name, IP: utils.GetNodePrimaryIP(node)})
		}
	}
	return result, nil
}
