// Copyright © 2024 The vjailbreak authors

package openstack

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"vjailbreak/vm"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/blockstorage/extensions/volumeactions"
	"github.com/gophercloud/gophercloud/openstack/blockstorage/v3/volumes"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/bootfromvolume"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/volumeattach"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/flavors"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/networks"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/ports"
)

//go:generate mockgen -source=../openstack/openstackops.go -destination=../openstack/openstackops_mock.go -package=openstack

type OpenstackOperations interface {
	CreateVolume(name string, size int64, ostype string, uefi bool) (*volumes.Volume, error)
	WaitForVolume(volumeID string) error
	AttachVolumeToVM(volumeID string) error
	WaitForVolumeAttachment(volumeID string) error
	DetachVolumeFromVM(volumeID string) error
	SetVolumeUEFI(volume *volumes.Volume) error
	EnableQGA(volume *volumes.Volume) error
	SetVolumeImageMetadata(volume *volumes.Volume) error
	SetVolumeBootable(volume *volumes.Volume) error
	GetClosestFlavour(cpu int32, memory int32) (*flavors.Flavor, error)
	GetNetwork(networkname string) (*networks.Network, error)
	CreatePort(networkid *networks.Network, mac, vmname string) (*ports.Port, error)
	CreateVM(flavor *flavors.Flavor, networkIDs, portIDs []string, vminfo vm.VMInfo) (*servers.Server, error)
	DeleteVolume(volumeID string) error
	FindDevice(volumeID string) (string, error)
}

type OpenStackClients struct {
	BlockStorageClient *gophercloud.ServiceClient
	ComputeClient      *gophercloud.ServiceClient
	NetworkingClient   *gophercloud.ServiceClient
}

type OpenStackMetadata struct {
	UUID string `json:"uuid"`
}

const MaxCPU = 9999999
const MaxRAM = 9999999

// Number of intervals to wait for the volume to become available
const MaxIntervalCount = 6

func validateOpenStack() (*OpenStackClients, error) {
	opts, err := openstack.AuthOptionsFromEnv()
	if err != nil {
		return nil, fmt.Errorf("failed to get OpenStack auth options: %s", err)
	}
	providerClient, err := openstack.AuthenticatedClient(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate OpenStack client: %s", err)
	}

	endpoint := gophercloud.EndpointOpts{
		Region: os.Getenv("OS_REGION_NAME"),
	}

	blockStorageClient, err := openstack.NewBlockStorageV3(providerClient, endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to create block storage client: %s", err)
	}

	computeClient, err := openstack.NewComputeV2(providerClient, endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to create compute client: %s", err)
	}

	networkingClient, err := openstack.NewNetworkV2(providerClient, endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to create networking client: %s", err)
	}

	return &OpenStackClients{
		BlockStorageClient: blockStorageClient,
		ComputeClient:      computeClient,
		NetworkingClient:   networkingClient,
	}, nil
}

func NewOpenStackClients() (*OpenStackClients, error) {
	ostackclients, err := validateOpenStack()
	if err != nil {
		return nil, fmt.Errorf("failed to validate OpenStack connection: %s", err)
	}
	return ostackclients, nil
}

func getCurrentInstanceUUID() (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://169.254.169.254/openstack/latest/meta_data.json", nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %s", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get response: %s", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %s", err)
	}

	var metadata OpenStackMetadata
	if err := json.Unmarshal(body, &metadata); err != nil {
		return "", fmt.Errorf("failed to unmarshal metadata: %s", err)
	}

	return metadata.UUID, nil
}

// create a new volume
func (osclient *OpenStackClients) CreateVolume(name string, size int64, ostype string, uefi bool) (*volumes.Volume, error) {
	blockStorageClient := osclient.BlockStorageClient

	opts := volumes.CreateOpts{
		Size: int(float64(size) / (1024 * 1024 * 1024)),
		Name: name,
	}
	volume, err := volumes.Create(blockStorageClient, opts).Extract()
	if err != nil {
		return nil, fmt.Errorf("failed to create volume: %s", err)
	}

	err = osclient.WaitForVolume(volume.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for volume: %s", err)
	}
	if uefi {
		err = osclient.SetVolumeUEFI(volume)
		if err != nil {
			return nil, fmt.Errorf("failed to set volume uefi: %s", err)
		}
	}

	if ostype == "windows" {
		err = osclient.SetVolumeImageMetadata(volume)
		if err != nil {
			return nil, fmt.Errorf("failed to set volume image metadata: %s", err)
		}
	}

	err = osclient.EnableQGA(volume)
	if err != nil {
		return nil, err
	}

	return volume, nil
}

func (osclient *OpenStackClients) DeleteVolume(volumeID string) error {
	err := volumes.Delete(osclient.BlockStorageClient, volumeID, volumes.DeleteOpts{}).ExtractErr()
	if err != nil {
		return fmt.Errorf("failed to delete volume: %s", err)
	}
	return nil
}

func (osclient *OpenStackClients) WaitForVolume(volumeID string) error {
	for i := 0; i < MaxIntervalCount; i++ {
		volume, err := volumes.Get(osclient.BlockStorageClient, volumeID).Extract()
		if err != nil {
			return fmt.Errorf("failed to get volume: %s", err)
		}

		if volume.Status == "available" {
			return nil
		}
		time.Sleep(5 * time.Second) // Wait for 5 seconds before checking again
	}
	return fmt.Errorf("volume did not become available within %d seconds", MaxIntervalCount*5)
}

func (osclient *OpenStackClients) AttachVolumeToVM(volumeID string) error {
	instanceID, err := getCurrentInstanceUUID()
	if err != nil {
		return fmt.Errorf("failed to get instance ID: %s", err)
	}
	_, err = volumeattach.Create(osclient.ComputeClient, instanceID, volumeattach.CreateOpts{
		VolumeID:            volumeID,
		DeleteOnTermination: false,
	}).Extract()
	if err != nil {
		return fmt.Errorf("failed to attach volume to VM: %s", err)
	}

	log.Println("Waiting for volume attachment")
	err = osclient.WaitForVolumeAttachment(volumeID)
	if err != nil {
		return fmt.Errorf("failed to wait for volume attachment: %s", err)
	}

	return nil
}

func (osclient *OpenStackClients) FindDevice(volumeID string) (string, error) {
	files, err := os.ReadDir("/dev/disk/by-id/")
	if err != nil {
		return "", fmt.Errorf("failed to read directory: %s", err)
	}

	for _, file := range files {
		if strings.Contains(file.Name(), volumeID[:18]) {
			devicePath, err := filepath.EvalSymlinks(filepath.Join("/dev/disk/by-id/", file.Name()))
			if err != nil {
				return "", fmt.Errorf("failed to evaluate symlink: %s", err)
			}

			return devicePath, nil
		}
	}

	return "", nil
}

func (osclient *OpenStackClients) WaitForVolumeAttachment(volumeID string) error {
	for i := 0; i < MaxIntervalCount; i++ {
		devicePath, _ := osclient.FindDevice(volumeID)
		if devicePath != "" {
			return nil
		}
		time.Sleep(5 * time.Second) // Wait for 5 seconds before checking again
	}
	return fmt.Errorf("volume attachment not found within %d seconds", MaxIntervalCount*5)
}

func (osclient *OpenStackClients) DetachVolumeFromVM(volumeID string) error {
	instanceID, err := getCurrentInstanceUUID()
	if err != nil {
		return fmt.Errorf("failed to get instance ID: %s", err)
	}
	err = volumeattach.Delete(osclient.ComputeClient, instanceID, volumeID).ExtractErr()
	if err != nil {
		return fmt.Errorf("failed to detach volume from VM: %s", err)
	}
	return nil
}

func (osclient *OpenStackClients) EnableQGA(volume *volumes.Volume) error {
	options := volumeactions.ImageMetadataOpts{
		Metadata: map[string]string{
			"hw_qemu_guest_agent": "yes",
			"hw_video_model":      "virtio",
		},
	}
	err := volumeactions.SetImageMetadata(osclient.BlockStorageClient, volume.ID, options).ExtractErr()
	if err != nil {
		return fmt.Errorf("failed to detach volume from VM: %s", err)
	}
	return nil
}

func (osclient *OpenStackClients) SetVolumeUEFI(volume *volumes.Volume) error {
	options := volumeactions.ImageMetadataOpts{
		Metadata: map[string]string{
			"hw_firmware_type": "uefi",
		},
	}
	err := volumeactions.SetImageMetadata(osclient.BlockStorageClient, volume.ID, options).ExtractErr()
	if err != nil {
		return fmt.Errorf("failed to set volume image metadata hw_firmware_type to uefi: %s", err)
	}
	return nil
}

func (osclient *OpenStackClients) SetVolumeImageMetadata(volume *volumes.Volume) error {
	options := volumeactions.ImageMetadataOpts{
		Metadata: map[string]string{
			"hw_disk_bus": "virtio",
			"os_type":     "windows",
		},
	}
	err := volumeactions.SetImageMetadata(osclient.BlockStorageClient, volume.ID, options).ExtractErr()
	if err != nil {
		return fmt.Errorf("failed to set volume image metadata for windows: %s", err)
	}
	return nil
}

func (osclient *OpenStackClients) SetVolumeBootable(volume *volumes.Volume) error {
	options := volumeactions.BootableOpts{
		Bootable: true,
	}
	err := volumeactions.SetBootable(osclient.BlockStorageClient, volume.ID, options).ExtractErr()
	if err != nil {
		return fmt.Errorf("failed to set volume as bootable: %s", err)
	}
	return nil
}

func (osclient *OpenStackClients) GetClosestFlavour(cpu int32, memory int32) (*flavors.Flavor, error) {
	allPages, err := flavors.ListDetail(osclient.ComputeClient, nil).AllPages()
	if err != nil {
		return nil, fmt.Errorf("failed to list flavors: %s", err)
	}

	allFlavors, err := flavors.ExtractFlavors(allPages)
	if err != nil {
		return nil, fmt.Errorf("failed to extract all flavors: %s", err)
	}

	log.Println("Current requirements:", cpu, "CPUs and", memory, "MB of RAM")

	bestFlavor := new(flavors.Flavor)
	bestFlavor.VCPUs = MaxCPU
	bestFlavor.RAM = MaxRAM
	// Find the smallest flavor that meets the requirements
	for _, flavor := range allFlavors {
		if flavor.VCPUs >= int(cpu) && flavor.RAM >= int(memory) {
			if flavor.VCPUs < bestFlavor.VCPUs || (flavor.VCPUs == bestFlavor.VCPUs && flavor.RAM < bestFlavor.RAM) {
				bestFlavor = &flavor
			}
		}
	}

	if bestFlavor.VCPUs != MaxCPU {
		log.Printf("The best flavor is:\nName: %s, ID: %s, RAM: %dMB, VCPUs: %d, Disk: %dGB\n",
			bestFlavor.Name, bestFlavor.ID, bestFlavor.RAM, bestFlavor.VCPUs, bestFlavor.Disk)
	} else {
		log.Println("No suitable flavor found.")
	}

	return bestFlavor, nil
}

func (osclient *OpenStackClients) GetNetwork(networkname string) (*networks.Network, error) {
	allPages, err := networks.List(osclient.NetworkingClient, nil).AllPages()
	if err != nil {
		return nil, fmt.Errorf("failed to list networks: %s", err)
	}

	allNetworks, err := networks.ExtractNetworks(allPages)
	if err != nil {
		return nil, fmt.Errorf("failed to extract all networks: %s", err)
	}

	for _, network := range allNetworks {
		if network.Name == networkname {
			return &network, nil
		}
	}
	return nil, fmt.Errorf("network not found")
}

func (osclient *OpenStackClients) CreatePort(network *networks.Network, mac, vmname string) (*ports.Port, error) {
	pages, err := ports.List(osclient.NetworkingClient, ports.ListOpts{
		NetworkID:  network.ID,
		MACAddress: mac,
	}).AllPages()
	if err != nil {
		return nil, fmt.Errorf("failed to list networks: %s", err)
	}

	portList, err := ports.ExtractPorts(pages)
	if err != nil {
		return nil, err
	}

	for _, port := range portList {
		if port.MACAddress == mac {
			log.Printf("Port with MAC address %s already exists, ID: %s\n", mac, port.ID)
			return &port, nil
		}
	}
	log.Printf("Port with MAC address %s does not exist, creating new port\n", mac)
	port, err := ports.Create(osclient.NetworkingClient, ports.CreateOpts{
		Name:       "port-" + vmname,
		NetworkID:  network.ID,
		MACAddress: mac,
	}).Extract()
	if err != nil {
		return nil, err
	}
	log.Println("Port created with ID: ", port.ID)
	return port, nil
}

func (osclient *OpenStackClients) CreateVM(flavor *flavors.Flavor, networkIDs, portIDs []string, vminfo vm.VMInfo) (*servers.Server, error) {
	blockDevice := bootfromvolume.BlockDevice{
		DeleteOnTermination: false,
		DestinationType:     bootfromvolume.DestinationVolume,
		SourceType:          bootfromvolume.SourceVolume,
		UUID:                vminfo.VMDisks[0].OpenstackVol.ID,
	}
	// Create the server
	openstacknws := []servers.Network{}
	for idx := range networkIDs {
		openstacknws = append(openstacknws, servers.Network{
			UUID: networkIDs[idx],
			Port: portIDs[idx],
		})
	}
	serverCreateOpts := servers.CreateOpts{
		Name:      vminfo.Name,
		FlavorRef: flavor.ID,
		Networks:  openstacknws,
	}

	createOpts := bootfromvolume.CreateOptsExt{
		CreateOptsBuilder: serverCreateOpts,
		BlockDevice:       []bootfromvolume.BlockDevice{blockDevice},
	}

	// Wait for disks to become available
	for _, disk := range vminfo.VMDisks {
		err := osclient.WaitForVolume(disk.OpenstackVol.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to wait for volume to become available: %s", err)
		}
	}

	server, err := servers.Create(osclient.ComputeClient, createOpts).Extract()
	if err != nil {
		return nil, fmt.Errorf("failed to create server: %s", err)
	}

	err = servers.WaitForStatus(osclient.ComputeClient, server.ID, "ACTIVE", 60)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for server to become active: %s", err)
	}

	log.Println("Server created with ID: ", server.ID)

	log.Println("Attaching Additional Disks")

	for _, disk := range vminfo.VMDisks[1:] {
		_, err := volumeattach.Create(osclient.ComputeClient, server.ID, volumeattach.CreateOpts{
			VolumeID:            disk.OpenstackVol.ID,
			DeleteOnTermination: false,
		}).Extract()
		if err != nil {
			return nil, fmt.Errorf("failed to attach volume to VM: %s", err)
		}
	}

	return server, nil
}
