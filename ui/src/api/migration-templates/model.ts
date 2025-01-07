export interface GetMigrationTemplatesList {
  apiVersion: string
  items: MigrationTemplate[]
  kind: string
  metadata: GetMigrationTemplatesMetadata
}

export interface MigrationTemplate {
  apiVersion: string
  kind: string
  metadata: MigrationTemplateMetadata
  spec: MigrationTemplateSpec
  status: MigrationTemplateStatus
}

export interface MigrationTemplateMetadata {
  annotations: Annotations
  creationTimestamp: Date
  generation: number
  name: string
  namespace: string
  resourceVersion: string
  uid: string
  labels: Labels
}

export interface Annotations {
  "kubectl.kubernetes.io/last-applied-configuration": string
}

export interface Labels {
  refresh: string
}

export interface MigrationTemplateSpec {
  destination: Destination
  networkMapping: string
  source: Source
  storageMapping: string
}

export interface Destination {
  openstackRef: string
}

export interface Source {
  datacenter: string
  vmwareRef: string
}

export interface MigrationTemplateStatus {
  openstack: Openstack
  vmware: VmData[]
}

export interface Openstack {
  networks: string[]
  volumeTypes: string[]
}

export interface VmData {
  datastores: string[]
  name: string
  networks?: string[]
}

export interface GetMigrationTemplatesMetadata {
  continue: string
  resourceVersion: string
}
