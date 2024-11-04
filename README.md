# vJailbreak
Helping VMware users migrate to Openstack

## v2v-helper
The main application that runs the migration. It is expected to run as a pod in a VM running in the target Openstack Environment.

## UI
This is the UI for vJailbreak.

## migration-controller
This is the k8s controller that schedules migrations. 

## v2v-cli
A CLI tool that starts the migration. It is not needed in the current version of vJailbreak.

## Sample Screenshots

### Form to start a migraiton
![alt text](assets/migrationform1.png)
![alt text](assets/migrationform2.png)

### Migration Progress
![alt text](assets/migrationprogress1.png)
![alt text](assets/migrationprogress2.png)


## Building
vJailbreak is intended to be run in a kubernetes environment (k3s) on the appliance VM. In order to build and deploy the kubernetes components, follow the instructions in `k8s/migration` to build and deploy the custom resources in the cluster.

In order to build v2v-helper,

    make v2v-helper

In order to build migration-controller,

    make vjail-controller

In order to build the UI,

    make ui

Change the image names in the makefile to push to another repository

## Usage

Download and install [ORAS](https://oras.land/docs/installation). Download the latest version of the vjailbreak image with the following command. 

    oras pull quay.io/platform9/vjailbreak:v0.1

This will download the vjailbreak qcow2 image locally. Upload it to your Openstack enviroment and create your appliance VM with it.

Then, ensure that your appliance can talk to your Openstack and VMware environments. This includes any setup required for VPNs, etc. If you do not have an Openstack environment, you can download the community edition of [Private Cloud Director](https://platform9.com/private-cloud-director/#experience) to get started.

Deploy all the following resources in the same namespace where you installed the Migration Controller. By default, it is `migration-system`.
1. Create the Creds objects. Ensure that after you create these objects, their status reflects that the credentials have been validated. If it is not validated, the migration will not proceed.

       apiVersion: vjailbreak.k8s.pf9.io/v1alpha1
       kind: OpenstackCreds
       metadata:
         name: sapmo1
         namespace: migration-system
       spec:
         OS_AUTH_URL: 
         OS_DOMAIN_NAME: 
         OS_USERNAME: 
         OS_PASSWORD:
         OS_REGION_NAME:  
         OS_TENANT_NAME:  
         OS_INSECURE: true/false <optional>
       ---
       apiVersion: vjailbreak.k8s.pf9.io/v1alpha1
       kind: VMwareCreds
       metadata:
         name: pnapbmc1
         namespace: migration-system
       spec:
         VCENTER_HOST: vcenter.phx.pnap.platform9.horse
         VCENTER_INSECURE:  true/false
         VCENTER_PASSWORD:
         VCENTER_USERNAME: 
  
  - OpenstackCreds use the variables from the openstack.rc file. All these fields are compulsory except OS_INSECURE
  - All the fields in VMwareCreds are compulsory

2. Create the mapping between networks in VMware and networks in Openstack

       apiVersion: vjailbreak.k8s.pf9.io/v1alpha1
       kind: NetworkMapping
       metadata:
         name: nwmap1
         namespace: migration-system
       spec:
         networks:
         - source: VM Network
           target: vlan3002
         - source: VM Network 2
           target: vlan3003
3. Create the mapping between datastores in VMware and volume types in Openstack

		apiVersion: vjailbreak.k8s.pf9.io/v1alpha1
		kind: StorageMapping
		metadata:
		  name: stmap1
		  namespace: migration-system
		spec:
		  storages:
		  - source: vcenter-datastore-1
		    target: lvm
		  - source: vcenter-datastore-2
		    target: ceph
4. Create the MigrationTemplate

       apiVersion: vjailbreak.k8s.pf9.io/v1alpha1
       kind: MigrationTemplate
       metadata:
         name: migrationtemplate-windows
         namespace: migration-system
       spec:
         networkMapping: name_of_networkMapping
         storageMapping: name_of_storageMapping
         osType: windows/linux <optional>
         source:
           datacenter: name_of_datacenter
           vmwareRef: name_of_VMwareCreds
         destination:
           openstackRef: name_of_OpenstackCreds

  - osType is optional. If not provided, the osType is retrieved from vcenter. If it could be automatically determined, migration will not proceed.

5. Finally, create the MigrationPlan

       apiVersion: vjailbreak.k8s.pf9.io/v1alpha1
       kind: MigrationPlan
       metadata:
         name: vm-migration-app1
         namespace: migration-system
       spec:
         migrationTemplate: migrationtemplate-windows
         retry: true/false <optional>
         advancedOptions:
           granularVolumeTypes: 
           - newvoltype1
           granularNetworks:
           - newnetworkname1
           - newnetworkname2
           granularPorts:
           - <port uuid 1>
           - <port uuid 2>
         migrationStrategy:
           type: hot/cold
           dataCopyStart: 2024-08-27T17:30:25.230Z
           vmCutoverStart: 2024-08-27T17:30:25.230Z
           vmCutoverEnd: 2024-08-28T17:30:25.230Z
           adminInitiatedCutOver: true/false
           performHealthChecks: true/false
           healthCheckPort: string
         virtualmachines:
           - - winserver2k12
             - winserver2k16
           - - winserver2k19
             - winserver2k22

  - retry: Optional. Retries one failed migration in a migration plan once. Set to false after a migration has been retried.
  - advancedOptions: This is an optional field for granular control over migration options. MigrationTemplate with mappings must still be present. These options override the ones in the template, if set. If you use these options, you must only have 1 VM present in the virtualmachines list.
    - granularVolumeTypes: In case you wish to provide different volume types to disks of a VM when they are all on the same datastore, you can speccify the volume type of each disk of your VM in order. You must define one volume type for one disk present on the VM
    - granularNetworks: In case you wish to override the default network mapping for a VM, you can provide a list of openstack network names to use in for each NIC on the VM, in order.
    - granularPorts: In case you wish to pre-create ports for a VM with certain configs and directly ptovide them to the target VM, you can define a list of port IDS to be used for each network on the VM. It will override options set in granularNetworks.
  - migrationStrategy: This is an optional field
    - type: 
      - cold: Cold indicates to power off VMs in migrationplan at the start of the migration. Quicker than hot
      - hot: Powers VM off just before cutover starts. Data copy occurs with the source VM powered on. May take longer
    - dataCopyStart: Optional.  ISO 8601 timestamp indicating when to start data copy
    - vmCutoverStart: Optional. ISO 8601 timestamp indicating when to start VM cutover
    - vmCutoverEnd: Optional. ISO 8601 timestamp indicating the latest time by when VM cutover can start. If this time has been passed before the cutover can start, migration will fail.
    - adminInitiatedCutOver: Set to true if you wish to manually trigger the cutover process. Default false
    - performHealthChecks: Set to false if you want to disable Ping and HTTP GET health check. Failing these checks does not clean up the targeted VM. Default true
    - healthCheckPort: Port to run the HTTP GET health check against. Default "443"
  - virtualmachines: Specify names of VMs to migrate. In this example the batch of VMs `winserver2k12` and `winserver2k16` migrate in parallel. `winserver2k19` and `winserver2k22` will wait for the first 2 to complete successfully, and then start in parallel. You can use this notation to specify whether VMs should migrate sequentially or parallelly within a plan.

Each VM migration will spawn a migration object. The status field contains a high level view of the progress of the migration of the VM. For more details about the migration, check the logs of the pod specified in the Migration object.