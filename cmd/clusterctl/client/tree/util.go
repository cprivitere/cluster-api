/*
Copyright 2020 The Kubernetes Authors.

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

package tree

import (
	"fmt"
	"sort"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/conditions"
	v1beta2conditions "sigs.k8s.io/cluster-api/util/conditions/v1beta2"
)

// GroupVersionVirtualObject is the group version for VirtualObject.
var GroupVersionVirtualObject = schema.GroupVersion{Group: "virtual.cluster.x-k8s.io", Version: clusterv1.GroupVersion.Version}

// GetReadyV1Beta2Condition returns the ReadyCondition for an object, if defined.
func GetReadyV1Beta2Condition(obj client.Object) *metav1.Condition {
	if getter, ok := obj.(v1beta2conditions.Getter); ok {
		return v1beta2conditions.Get(getter, clusterv1.ReadyV1Beta2Condition)
	}

	if objUnstructured, ok := obj.(*unstructured.Unstructured); ok {
		c, err := v1beta2conditions.UnstructuredGet(objUnstructured, clusterv1.ReadyV1Beta2Condition)
		if err != nil {
			return nil
		}
		return c
	}

	return nil
}

// GetAvailableV1Beta2Condition returns the AvailableCondition for an object, if defined.
func GetAvailableV1Beta2Condition(obj client.Object) *metav1.Condition {
	if getter, ok := obj.(v1beta2conditions.Getter); ok {
		return v1beta2conditions.Get(getter, clusterv1.AvailableV1Beta2Condition)
	}

	if objUnstructured, ok := obj.(*unstructured.Unstructured); ok {
		c, err := v1beta2conditions.UnstructuredGet(objUnstructured, clusterv1.AvailableV1Beta2Condition)
		if err != nil {
			return nil
		}
		return c
	}

	return nil
}

// GetMachineUpToDateV1Beta2Condition returns machine's UpToDate condition, if defined.
// Note: The UpToDate condition only exist on machines, so no need to support reading from unstructured.
func GetMachineUpToDateV1Beta2Condition(obj client.Object) *metav1.Condition {
	if getter, ok := obj.(v1beta2conditions.Getter); ok {
		return v1beta2conditions.Get(getter, clusterv1.MachineUpToDateV1Beta2Condition)
	}
	return nil
}

// GetReadyCondition returns the ReadyCondition for an object, if defined.
func GetReadyCondition(obj client.Object) *clusterv1.Condition {
	getter := objToGetter(obj)
	if getter == nil {
		return nil
	}
	return conditions.Get(getter, clusterv1.ReadyCondition)
}

// GetAllV1Beta2Conditions returns the other conditions (all the conditions except ready) for an object, if defined.
func GetAllV1Beta2Conditions(obj client.Object) []metav1.Condition {
	if getter, ok := obj.(v1beta2conditions.Getter); ok {
		return getter.GetV1Beta2Conditions()
	}

	if objUnstructured, ok := obj.(*unstructured.Unstructured); ok {
		conditionList, err := v1beta2conditions.UnstructuredGetAll(objUnstructured)
		if err != nil {
			return nil
		}
		return conditionList
	}

	return nil
}

// GetOtherConditions returns the other conditions (all the conditions except ready) for an object, if defined.
func GetOtherConditions(obj client.Object) []*clusterv1.Condition {
	getter := objToGetter(obj)
	if getter == nil {
		return nil
	}
	var conditions []*clusterv1.Condition
	for _, c := range getter.GetConditions() {
		if c.Type != clusterv1.ReadyCondition {
			conditions = append(conditions, &c)
		}
	}
	sort.Slice(conditions, func(i, j int) bool {
		return conditions[i].Type < conditions[j].Type
	})
	return conditions
}

func setAvailableV1Beta2Condition(obj client.Object, available *metav1.Condition) {
	if setter, ok := obj.(v1beta2conditions.Setter); ok {
		v1beta2conditions.Set(setter, *available)
	}
}

func setReadyV1Beta2Condition(obj client.Object, ready *metav1.Condition) {
	if setter, ok := obj.(v1beta2conditions.Setter); ok {
		v1beta2conditions.Set(setter, *ready)
	}
}

func setUpToDateV1Beta2Condition(obj client.Object, upToDate *metav1.Condition) {
	if setter, ok := obj.(v1beta2conditions.Setter); ok {
		v1beta2conditions.Set(setter, *upToDate)
	}
}

func setReadyCondition(obj client.Object, ready *clusterv1.Condition) {
	setter := objToSetter(obj)
	if setter == nil {
		return
	}
	conditions.Set(setter, ready)
}

func objToGetter(obj client.Object) conditions.Getter {
	if getter, ok := obj.(conditions.Getter); ok {
		return getter
	}

	objUnstructured, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return nil
	}
	getter := conditions.UnstructuredGetter(objUnstructured)
	return getter
}

func objToSetter(obj client.Object) conditions.Setter {
	if setter, ok := obj.(conditions.Setter); ok {
		return setter
	}

	objUnstructured, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return nil
	}
	setter := conditions.UnstructuredSetter(objUnstructured)
	return setter
}

// VirtualObject returns a new virtual object.
func VirtualObject(namespace, kind, name string) *NodeObject {
	return &NodeObject{
		TypeMeta: metav1.TypeMeta{
			Kind:       kind,
			APIVersion: GroupVersionVirtualObject.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
			Annotations: map[string]string{
				VirtualObjectAnnotation: "True",
			},
			UID: types.UID(fmt.Sprintf("%s, Kind=%s, %s/%s", GroupVersionVirtualObject.String(), kind, namespace, name)),
		},
		Status: NodeStatus{},
	}
}

// ObjectReferenceObject returns a new object referenced by the objectRef.
func ObjectReferenceObject(objectRef *corev1.ObjectReference) *NodeObject {
	return &NodeObject{
		TypeMeta: metav1.TypeMeta{
			Kind:       objectRef.APIVersion,
			APIVersion: objectRef.Kind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: objectRef.Namespace,
			Name:      objectRef.Name,
			UID:       types.UID(fmt.Sprintf("%s, Kind=%s, %s/%s", objectRef.APIVersion, objectRef.Kind, objectRef.Namespace, objectRef.Name)),
		},
		Status: NodeStatus{},
	}
}
