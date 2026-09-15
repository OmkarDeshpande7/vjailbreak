import type { FormValues, RollingFormParams, StorageCopyMethod } from './types'

export enum CUTOVER_TYPES {
  'IMMEDIATE' = '0',
  'ADMIN_INITIATED' = '1',
  'TIME_WINDOW' = '2'
}

export enum OS_TYPES {
  'AUTO_DETECT' = 'default',
  'WINDOWS' = 'windowsGuest',
  'LINUX' = 'linuxGuest'
}

export const DATA_COPY_OPTIONS = [
  { value: 'cold', label: 'Power off live VMs, then copy' },
  { value: 'hot', label: 'Copy live VMs, then power off' },
  { value: 'mock', label: 'Do not Turn off the source VM'}
]

export const OS_TYPES_OPTIONS = [
  { value: OS_TYPES.AUTO_DETECT, label: 'Auto-detect' },
  { value: OS_TYPES.WINDOWS, label: 'Windows' },
  { value: OS_TYPES.LINUX, label: 'Linux' }
]

export const VM_CUTOVER_OPTIONS = [
  {
    value: CUTOVER_TYPES.IMMEDIATE,
    label: 'Cutover immediately after data copy'
  },
  { value: CUTOVER_TYPES.ADMIN_INITIATED, label: 'Admin initiated cutover' },
  { value: CUTOVER_TYPES.TIME_WINDOW, label: 'Cutover during time window' }
]

// ---------------------------------------------------------------------------
// NetworkAndStorageMappingStep constants
// ---------------------------------------------------------------------------

export const STORAGE_COPY_METHOD_OPTIONS = [
  { value: 'HotAdd', label: 'vJailbreak Accelerated Copy' },
  { value: 'StorageAcceleratedCopy', label: 'Storage Accelerated Copy' },
  { value: 'normal', label: 'Standard Copy' }
] as const

// Both entry points into a migration — the Migration Form and the Cluster Conversion
// (rolling migration) form — start on the same copy method. Keep this the single source
// of truth so the two forms cannot drift apart again.
export const DEFAULT_STORAGE_COPY_METHOD: StorageCopyMethod = 'HotAdd'

// Initial params for the two migration entry points. They are declared side by side so a
// default added to one is visibly missing from the other.
export const MIGRATION_FORM_DEFAULTS: Partial<FormValues> = {
  removeVMwareTools: true,
  storageCopyMethod: DEFAULT_STORAGE_COPY_METHOD
}

export const ROLLING_FORM_DEFAULTS: RollingFormParams = {
  removeVMwareTools: true,
  storageCopyMethod: DEFAULT_STORAGE_COPY_METHOD
}

// ---------------------------------------------------------------------------
// MigrationsTable constants
// ---------------------------------------------------------------------------

export const STATUS_ORDER: Record<string, number> = {
  Running: 0,
  Failed: 1,
  Succeeded: 2,
  Pending: 3
}

// ---------------------------------------------------------------------------
// MigrationForm / RollingMigrationForm defaults
// ---------------------------------------------------------------------------

export const DEFAULT_MIGRATION_OPTIONS = {
  dataCopyMethod: false,
  dataCopyStartTime: false,
  cutoverOption: false,
  cutoverStartTime: false,
  cutoverEndTime: false,
  postMigrationScript: false,
  useGPU: false,
  postMigrationAction: {
    suffix: false,
    folderName: false,
    renameVm: false,
    moveToFolder: false
  }
}

export const DRAWER_WIDTH = 1400

// ---------------------------------------------------------------------------
// VmsSelectionStep constants
// ---------------------------------------------------------------------------

export const MIGRATED_TOOLTIP_MESSAGE = 'This VM is migrating or already has been migrated.'
export const FLAVOR_NOT_FOUND_MESSAGE =
  'Appropriate flavor not found. Please assign a flavor before selecting this VM for migration or create a flavor.'
export const DEFAULT_PAGINATION_MODEL = { page: 0, pageSize: 5 }

// ---------------------------------------------------------------------------
// MigrationOptionsAlt constants
// ---------------------------------------------------------------------------

export const NEXT_SCRIPT_DELIMITER = '### NEXT SCRIPT ###'
