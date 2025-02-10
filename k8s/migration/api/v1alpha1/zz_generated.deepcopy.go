//go:build !ignore_autogenerated

/*
Copyright 2024.

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	"k8s.io/api/core/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AdvancedOptions) DeepCopyInto(out *AdvancedOptions) {
	*out = *in
	if in.GranularVolumeTypes != nil {
		in, out := &in.GranularVolumeTypes, &out.GranularVolumeTypes
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.GranularNetworks != nil {
		in, out := &in.GranularNetworks, &out.GranularNetworks
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.GranularPorts != nil {
		in, out := &in.GranularPorts, &out.GranularPorts
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AdvancedOptions.
func (in *AdvancedOptions) DeepCopy() *AdvancedOptions {
	if in == nil {
		return nil
	}
	out := new(AdvancedOptions)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Migration) DeepCopyInto(out *Migration) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Migration.
func (in *Migration) DeepCopy() *Migration {
	if in == nil {
		return nil
	}
	out := new(Migration)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Migration) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MigrationList) DeepCopyInto(out *MigrationList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Migration, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MigrationList.
func (in *MigrationList) DeepCopy() *MigrationList {
	if in == nil {
		return nil
	}
	out := new(MigrationList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *MigrationList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MigrationPlan) DeepCopyInto(out *MigrationPlan) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MigrationPlan.
func (in *MigrationPlan) DeepCopy() *MigrationPlan {
	if in == nil {
		return nil
	}
	out := new(MigrationPlan)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *MigrationPlan) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MigrationPlanList) DeepCopyInto(out *MigrationPlanList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]MigrationPlan, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MigrationPlanList.
func (in *MigrationPlanList) DeepCopy() *MigrationPlanList {
	if in == nil {
		return nil
	}
	out := new(MigrationPlanList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *MigrationPlanList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MigrationPlanSpec) DeepCopyInto(out *MigrationPlanSpec) {
	*out = *in
	in.MigrationStrategy.DeepCopyInto(&out.MigrationStrategy)
	if in.VirtualMachines != nil {
		in, out := &in.VirtualMachines, &out.VirtualMachines
		*out = make([][]string, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = make([]string, len(*in))
				copy(*out, *in)
			}
		}
	}
	in.AdvancedOptions.DeepCopyInto(&out.AdvancedOptions)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MigrationPlanSpec.
func (in *MigrationPlanSpec) DeepCopy() *MigrationPlanSpec {
	if in == nil {
		return nil
	}
	out := new(MigrationPlanSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MigrationPlanStatus) DeepCopyInto(out *MigrationPlanStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MigrationPlanStatus.
func (in *MigrationPlanStatus) DeepCopy() *MigrationPlanStatus {
	if in == nil {
		return nil
	}
	out := new(MigrationPlanStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MigrationPlanStrategy) DeepCopyInto(out *MigrationPlanStrategy) {
	*out = *in
	in.DataCopyStart.DeepCopyInto(&out.DataCopyStart)
	in.VMCutoverStart.DeepCopyInto(&out.VMCutoverStart)
	in.VMCutoverEnd.DeepCopyInto(&out.VMCutoverEnd)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MigrationPlanStrategy.
func (in *MigrationPlanStrategy) DeepCopy() *MigrationPlanStrategy {
	if in == nil {
		return nil
	}
	out := new(MigrationPlanStrategy)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MigrationSpec) DeepCopyInto(out *MigrationSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MigrationSpec.
func (in *MigrationSpec) DeepCopy() *MigrationSpec {
	if in == nil {
		return nil
	}
	out := new(MigrationSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MigrationStatus) DeepCopyInto(out *MigrationStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]v1.PodCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MigrationStatus.
func (in *MigrationStatus) DeepCopy() *MigrationStatus {
	if in == nil {
		return nil
	}
	out := new(MigrationStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MigrationTemplate) DeepCopyInto(out *MigrationTemplate) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MigrationTemplate.
func (in *MigrationTemplate) DeepCopy() *MigrationTemplate {
	if in == nil {
		return nil
	}
	out := new(MigrationTemplate)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *MigrationTemplate) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MigrationTemplateDestination) DeepCopyInto(out *MigrationTemplateDestination) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MigrationTemplateDestination.
func (in *MigrationTemplateDestination) DeepCopy() *MigrationTemplateDestination {
	if in == nil {
		return nil
	}
	out := new(MigrationTemplateDestination)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MigrationTemplateList) DeepCopyInto(out *MigrationTemplateList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]MigrationTemplate, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MigrationTemplateList.
func (in *MigrationTemplateList) DeepCopy() *MigrationTemplateList {
	if in == nil {
		return nil
	}
	out := new(MigrationTemplateList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *MigrationTemplateList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MigrationTemplateSource) DeepCopyInto(out *MigrationTemplateSource) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MigrationTemplateSource.
func (in *MigrationTemplateSource) DeepCopy() *MigrationTemplateSource {
	if in == nil {
		return nil
	}
	out := new(MigrationTemplateSource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MigrationTemplateSpec) DeepCopyInto(out *MigrationTemplateSpec) {
	*out = *in
	out.Source = in.Source
	out.Destination = in.Destination
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MigrationTemplateSpec.
func (in *MigrationTemplateSpec) DeepCopy() *MigrationTemplateSpec {
	if in == nil {
		return nil
	}
	out := new(MigrationTemplateSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MigrationTemplateStatus) DeepCopyInto(out *MigrationTemplateStatus) {
	*out = *in
	if in.VMWare != nil {
		in, out := &in.VMWare, &out.VMWare
		*out = make([]VMInfo, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	in.Openstack.DeepCopyInto(&out.Openstack)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MigrationTemplateStatus.
func (in *MigrationTemplateStatus) DeepCopy() *MigrationTemplateStatus {
	if in == nil {
		return nil
	}
	out := new(MigrationTemplateStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Network) DeepCopyInto(out *Network) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Network.
func (in *Network) DeepCopy() *Network {
	if in == nil {
		return nil
	}
	out := new(Network)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NetworkMapping) DeepCopyInto(out *NetworkMapping) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NetworkMapping.
func (in *NetworkMapping) DeepCopy() *NetworkMapping {
	if in == nil {
		return nil
	}
	out := new(NetworkMapping)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *NetworkMapping) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NetworkMappingList) DeepCopyInto(out *NetworkMappingList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]NetworkMapping, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NetworkMappingList.
func (in *NetworkMappingList) DeepCopy() *NetworkMappingList {
	if in == nil {
		return nil
	}
	out := new(NetworkMappingList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *NetworkMappingList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NetworkMappingSpec) DeepCopyInto(out *NetworkMappingSpec) {
	*out = *in
	if in.Networks != nil {
		in, out := &in.Networks, &out.Networks
		*out = make([]Network, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NetworkMappingSpec.
func (in *NetworkMappingSpec) DeepCopy() *NetworkMappingSpec {
	if in == nil {
		return nil
	}
	out := new(NetworkMappingSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NetworkMappingStatus) DeepCopyInto(out *NetworkMappingStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NetworkMappingStatus.
func (in *NetworkMappingStatus) DeepCopy() *NetworkMappingStatus {
	if in == nil {
		return nil
	}
	out := new(NetworkMappingStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OpenstackCreds) DeepCopyInto(out *OpenstackCreds) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OpenstackCreds.
func (in *OpenstackCreds) DeepCopy() *OpenstackCreds {
	if in == nil {
		return nil
	}
	out := new(OpenstackCreds)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *OpenstackCreds) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OpenstackCredsList) DeepCopyInto(out *OpenstackCredsList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]OpenstackCreds, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OpenstackCredsList.
func (in *OpenstackCredsList) DeepCopy() *OpenstackCredsList {
	if in == nil {
		return nil
	}
	out := new(OpenstackCredsList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *OpenstackCredsList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OpenstackCredsSpec) DeepCopyInto(out *OpenstackCredsSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OpenstackCredsSpec.
func (in *OpenstackCredsSpec) DeepCopy() *OpenstackCredsSpec {
	if in == nil {
		return nil
	}
	out := new(OpenstackCredsSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OpenstackCredsStatus) DeepCopyInto(out *OpenstackCredsStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OpenstackCredsStatus.
func (in *OpenstackCredsStatus) DeepCopy() *OpenstackCredsStatus {
	if in == nil {
		return nil
	}
	out := new(OpenstackCredsStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OpenstackInfo) DeepCopyInto(out *OpenstackInfo) {
	*out = *in
	if in.VolumeTypes != nil {
		in, out := &in.VolumeTypes, &out.VolumeTypes
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Networks != nil {
		in, out := &in.Networks, &out.Networks
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OpenstackInfo.
func (in *OpenstackInfo) DeepCopy() *OpenstackInfo {
	if in == nil {
		return nil
	}
	out := new(OpenstackInfo)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Storage) DeepCopyInto(out *Storage) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Storage.
func (in *Storage) DeepCopy() *Storage {
	if in == nil {
		return nil
	}
	out := new(Storage)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StorageMapping) DeepCopyInto(out *StorageMapping) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StorageMapping.
func (in *StorageMapping) DeepCopy() *StorageMapping {
	if in == nil {
		return nil
	}
	out := new(StorageMapping)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *StorageMapping) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StorageMappingList) DeepCopyInto(out *StorageMappingList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]StorageMapping, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StorageMappingList.
func (in *StorageMappingList) DeepCopy() *StorageMappingList {
	if in == nil {
		return nil
	}
	out := new(StorageMappingList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *StorageMappingList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StorageMappingSpec) DeepCopyInto(out *StorageMappingSpec) {
	*out = *in
	if in.Storages != nil {
		in, out := &in.Storages, &out.Storages
		*out = make([]Storage, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StorageMappingSpec.
func (in *StorageMappingSpec) DeepCopy() *StorageMappingSpec {
	if in == nil {
		return nil
	}
	out := new(StorageMappingSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StorageMappingStatus) DeepCopyInto(out *StorageMappingStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StorageMappingStatus.
func (in *StorageMappingStatus) DeepCopy() *StorageMappingStatus {
	if in == nil {
		return nil
	}
	out := new(StorageMappingStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VMInfo) DeepCopyInto(out *VMInfo) {
	*out = *in
	if in.Datastores != nil {
		in, out := &in.Datastores, &out.Datastores
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Networks != nil {
		in, out := &in.Networks, &out.Networks
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VMInfo.
func (in *VMInfo) DeepCopy() *VMInfo {
	if in == nil {
		return nil
	}
	out := new(VMInfo)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VMwareCreds) DeepCopyInto(out *VMwareCreds) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VMwareCreds.
func (in *VMwareCreds) DeepCopy() *VMwareCreds {
	if in == nil {
		return nil
	}
	out := new(VMwareCreds)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *VMwareCreds) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VMwareCredsList) DeepCopyInto(out *VMwareCredsList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]VMwareCreds, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VMwareCredsList.
func (in *VMwareCredsList) DeepCopy() *VMwareCredsList {
	if in == nil {
		return nil
	}
	out := new(VMwareCredsList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *VMwareCredsList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VMwareCredsSpec) DeepCopyInto(out *VMwareCredsSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VMwareCredsSpec.
func (in *VMwareCredsSpec) DeepCopy() *VMwareCredsSpec {
	if in == nil {
		return nil
	}
	out := new(VMwareCredsSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VMwareCredsStatus) DeepCopyInto(out *VMwareCredsStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VMwareCredsStatus.
func (in *VMwareCredsStatus) DeepCopy() *VMwareCredsStatus {
	if in == nil {
		return nil
	}
	out := new(VMwareCredsStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VjailbreakNode) DeepCopyInto(out *VjailbreakNode) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VjailbreakNode.
func (in *VjailbreakNode) DeepCopy() *VjailbreakNode {
	if in == nil {
		return nil
	}
	out := new(VjailbreakNode)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *VjailbreakNode) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VjailbreakNodeList) DeepCopyInto(out *VjailbreakNodeList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]VjailbreakNode, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VjailbreakNodeList.
func (in *VjailbreakNodeList) DeepCopy() *VjailbreakNodeList {
	if in == nil {
		return nil
	}
	out := new(VjailbreakNodeList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *VjailbreakNodeList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VjailbreakNodeSpec) DeepCopyInto(out *VjailbreakNodeSpec) {
	*out = *in
	out.OpenstackCreds = in.OpenstackCreds
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VjailbreakNodeSpec.
func (in *VjailbreakNodeSpec) DeepCopy() *VjailbreakNodeSpec {
	if in == nil {
		return nil
	}
	out := new(VjailbreakNodeSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VjailbreakNodeStatus) DeepCopyInto(out *VjailbreakNodeStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VjailbreakNodeStatus.
func (in *VjailbreakNodeStatus) DeepCopy() *VjailbreakNodeStatus {
	if in == nil {
		return nil
	}
	out := new(VjailbreakNodeStatus)
	in.DeepCopyInto(out)
	return out
}
