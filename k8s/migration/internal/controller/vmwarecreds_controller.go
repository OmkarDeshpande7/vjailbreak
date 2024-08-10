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

package controller

import (
	"context"
	"fmt"
	"net/url"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	vjailbreakv1alpha1 "github.com/platform9/vjailbreak/k8s/migration/api/v1alpha1"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/session/cache"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

// VMwareCredsReconciler reconciles a VMwareCreds object
type VMwareCredsReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=vjailbreak.k8s.pf9.io,resources=vmwarecreds,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=vjailbreak.k8s.pf9.io,resources=vmwarecreds/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=vjailbreak.k8s.pf9.io,resources=vmwarecreds/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the VMwareCreds object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.4/pkg/reconcile
func (r *VMwareCredsReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctxlog := log.FromContext(ctx)

	// Get the VMwareCreds object
	vmwcreds := &vjailbreakv1alpha1.VMwareCreds{}
	if err := r.Get(ctx, req.NamespacedName, vmwcreds); err != nil {
		if apierrors.IsNotFound(err) {
			ctxlog.Info("Received ignorable event for a recently deleted VMWareCreds.")
			return ctrl.Result{}, nil
		}
		ctxlog.Error(err, fmt.Sprintf("Unexpected error reading VMWareCreds '%s' object", vmwcreds.Name))
		return ctrl.Result{}, err
	}

	if vmwcreds.ObjectMeta.DeletionTimestamp.IsZero() {
		ctxlog.Info(fmt.Sprintf("VMwareCreds '%s' CR is being created or updated", vmwcreds.Name))
		ctxlog.Info(fmt.Sprintf("Validating VMwareCreds '%s' object", vmwcreds.Name))
		if _, err := validateVMwareCreds(vmwcreds); err != nil {
			// Update the status of the VMwareCreds object
			vmwcreds.Status.VMwareValidationStatus = "Failed"
			vmwcreds.Status.VMwareValidationMessage = fmt.Sprintf("Error validating VMwareCreds '%s': %s", vmwcreds.Name, err)
			if err := r.Status().Update(ctx, vmwcreds); err != nil {
				ctxlog.Error(err, fmt.Sprintf("Error updating status of VMwareCreds '%s': %s", vmwcreds.Name, err))
				return ctrl.Result{}, err
			}
		} else {
			ctxlog.Info(fmt.Sprintf("Successfully authenticated to VMware '%s'", vmwcreds.Spec.VcenterHost))
			// Update the status of the VMwareCreds object
			vmwcreds.Status.VMwareValidationStatus = "Success"
			vmwcreds.Status.VMwareValidationMessage = "Successfully authenticated to VMware"
			if err := r.Status().Update(ctx, vmwcreds); err != nil {
				ctxlog.Error(err, fmt.Sprintf("Error updating status of VMwareCreds '%s': %s", vmwcreds.Name, err))
				return ctrl.Result{}, err
			}
		}
	}
	return ctrl.Result{}, nil
}

func validateVMwareCreds(vmwcreds *vjailbreakv1alpha1.VMwareCreds) (*vim25.Client, error) {
	host := vmwcreds.Spec.VcenterHost
	username := vmwcreds.Spec.VcenterUsername
	password := vmwcreds.Spec.VcenterPassword
	disableSSLVerification := vmwcreds.Spec.VcenterInsecure
	if host[:4] != "http" {
		host = "https://" + host
	}
	if host[len(host)-4:] != "/sdk" {
		host += "/sdk"
	}
	u, err := url.Parse(host)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %v", err)
	}
	u.User = url.UserPassword(username, password)
	// fmt.Println(u)
	// Connect and log in to ESX or vCenter
	// Share govc's session cache
	s := &cache.Session{
		URL:      u,
		Insecure: disableSSLVerification,
		Reauth:   true,
	}

	c := new(vim25.Client)
	err = s.Login(context.Background(), c, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to login: %v", err)
	}
	return c, nil
}

func GetVMwNetworks(ctx context.Context, vmwcreds *vjailbreakv1alpha1.VMwareCreds, datacenter, vmname string) ([]string, error) {
	var networks []string
	c, err := validateVMwareCreds(vmwcreds)
	if err != nil {
		return nil, fmt.Errorf("failed to validate vCenter connection: %v", err)
	}
	finder := find.NewFinder(c, false)
	dc, err := finder.Datacenter(ctx, datacenter)
	if err != nil {
		return nil, fmt.Errorf("failed to find datacenter: %v", err)
	}
	finder.SetDatacenter(dc)

	// Get the vm
	vm, err := finder.VirtualMachine(ctx, vmname)
	if err != nil {
		return nil, fmt.Errorf("failed to find vm: %v", err)
	}

	// Get the network name of the VM
	var o mo.VirtualMachine
	err = vm.Properties(ctx, vm.Reference(), []string{"config"}, &o)
	if err != nil {
		return nil, fmt.Errorf("failed to get VM properties: %v", err)
	}

	for _, device := range o.Config.Hardware.Device {
		switch nic := device.(type) {
		case *types.VirtualE1000e:
			nic = device.(*types.VirtualE1000e)
			networks = append(networks, nic.DeviceInfo.GetDescription().Summary)
		}
	}

	return networks, nil
}

func GetVMwDatastore(ctx context.Context, vmwcreds *vjailbreakv1alpha1.VMwareCreds, datacenter, vmname string) ([]string, error) {
	c, err := validateVMwareCreds(vmwcreds)
	if err != nil {
		return nil, fmt.Errorf("failed to validate vCenter connection: %v", err)
	}
	finder := find.NewFinder(c, false)
	dc, err := finder.Datacenter(ctx, datacenter)
	if err != nil {
		return nil, fmt.Errorf("failed to find datacenter: %v", err)
	}
	finder.SetDatacenter(dc)

	// Get the vm
	vm, err := finder.VirtualMachine(ctx, vmname)
	if err != nil {
		return nil, fmt.Errorf("failed to find vm: %v", err)
	}

	var vmProps mo.VirtualMachine
	err = vm.Properties(ctx, vm.Reference(), []string{"config"}, &vmProps)
	if err != nil {
		return nil, fmt.Errorf("failed to get VM properties: %v", err)
	}

	var datastores []string
	var ds mo.Datastore
	var dsref types.ManagedObjectReference
	for _, device := range vmProps.Config.Hardware.Device {
		switch device.(type) {
		case *types.VirtualDisk:
			disk := device.GetVirtualDevice()
			if b, ok := disk.Backing.(*types.VirtualDiskFlatVer2BackingInfo); ok {
				dsref = b.Datastore.Reference()
			} else if b, ok := disk.Backing.(*types.VirtualDiskSparseVer2BackingInfo); ok {
				dsref = b.Datastore.Reference()
			} else if b, ok := disk.Backing.(*types.VirtualDiskRawDiskMappingVer1BackingInfo); ok {
				dsref = b.Datastore.Reference()
			} else {
				return nil, fmt.Errorf("unsupported disk backing type: %T", disk.Backing)
			}
			err = property.DefaultCollector(c).RetrieveOne(ctx, dsref, []string{"name"}, &ds)
			if err != nil {
				return nil, fmt.Errorf("failed to get datastore: %v", err)
			}
			datastores = append(datastores, ds.Name)
		}
	}
	return datastores, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *VMwareCredsReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&vjailbreakv1alpha1.VMwareCreds{}).
		Complete(r)
}
