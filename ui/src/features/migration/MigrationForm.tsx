import { Alert, AlertTitle, Box, Drawer, styled } from "@mui/material"
import { useQueryClient } from "@tanstack/react-query"
import axios from "axios"
import { flatten, uniq } from "ramda"
import { useEffect, useMemo, useState } from "react"
import { useNavigate } from "react-router-dom"
import { createMigrationPlanJson } from "src/api/migration-plans/helpers"
import { postMigrationPlan } from "src/api/migration-plans/migrationPlans"
import { MigrationPlan } from "src/api/migration-plans/model"
import { createMigrationTemplateJson } from "src/api/migration-templates/helpers"
import {
  getMigrationTemplate,
  patchMigrationTemplate,
  postMigrationTemplate,
  deleteMigrationTemplate,
} from "src/api/migration-templates/migrationTemplates"
import { MigrationTemplate, VmData } from "src/api/migration-templates/model"
import { getMigrations } from "src/api/migrations/migrations"
import { Migration } from "src/api/migrations/model"
import { createNetworkMappingJson } from "src/api/network-mapping/helpers"
import { postNetworkMapping } from "src/api/network-mapping/networkMappings"
import { createOpenstackCredsJson } from "src/api/openstack-creds/helpers"
import { OpenstackCreds } from "src/api/openstack-creds/model"
import {
  getOpenstackCredentials,
  postOpenstackCredentials,
  deleteOpenstackCredentials,
} from "src/api/openstack-creds/openstackCreds"
import { createStorageMappingJson } from "src/api/storage-mappings/helpers"
import { postStorageMapping } from "src/api/storage-mappings/storageMappings"
import { createVmwareCredsJson } from "src/api/vmware-creds/helpers"
import { VMwareCreds } from "src/api/vmware-creds/model"
import {
  getVmwareCredentials,
  postVmwareCredentials,
  deleteVmwareCredentials,
} from "src/api/vmware-creds/vmwareCreds"
import { THREE_SECONDS, TWENTY_SECONDS } from "src/constants"
import { MIGRATIONS_QUERY_KEY } from "src/hooks/api/useMigrationsQuery"
import { useInterval } from "src/hooks/useInterval"
import useParams from "src/hooks/useParams"
import { isNilOrEmpty } from "src/utils"
import Footer from "../../components/forms/Footer"
import Header from "../../components/forms/Header"
import MigrationOptions from "./MigrationOptionsAlt"
import NetworkAndStorageMappingStep from "./NetworkAndStorageMappingStep"
import SourceAndDestinationEnvStep from "./SourceAndDestinationEnvStep"
import VmsSelectionStep from "./VmsSelectionStep"
import { CUTOVER_TYPES, OS_TYPES } from "./constants"

const stringsCompareFn = (a, b) =>
  a.toLowerCase().localeCompare(b.toLowerCase())

const StyledDrawer = styled(Drawer)(() => ({
  "& .MuiDrawer-paper": {
    display: "grid",
    gridTemplateRows: "max-content 1fr max-content",
    width: "1034px",
  },
}))

const DrawerContent = styled("div")(({ theme }) => ({
  overflow: "auto",
  padding: theme.spacing(4, 6, 4, 4),
}))

export interface FormValues extends Record<string, unknown> {
  vmwareCreds?: {
    vcenterHost: string
    datacenter: string
    username: string
    password: string
  }
  openstackCreds?: OpenstackCreds
  vms?: VmData[]
  networkMappings?: { source: string; target: string }[]
  storageMappings?: { source: string; target: string }[]
  // Optional Params
  dataCopyMethod?: string
  dataCopyStartTime?: string
  cutoverOption?: string
  cutoverStartTime?: string
  cutoverEndTime?: string
  postMigrationScript?: string
  retryOnFailure?: boolean
  osType?: string
}

export interface SelectedMigrationOptionsType extends Record<string, unknown> {
  dataCopyMethod: boolean
  dataCopyStartTime: boolean
  cutoverOption: boolean
  cutoverStartTime: boolean
  cutoverEndTime: boolean
  postMigrationScript: boolean
  osType: boolean
}

// Default state for checkboxes
const defaultMigrationOptions = {
  dataCopyMethod: false,
  dataCopyStartTime: false,
  cutoverOption: false,
  cutoverStartTime: false,
  cutoverEndTime: false,
  postMigrationScript: false,
  osType: false,
}

const defaultValues: Partial<FormValues> = {}

export type FieldErrors = { [formId: string]: string }

interface MigrationFormDrawerProps {
  open: boolean
  onClose: () => void
  reloadMigrations?: () => void
}

export default function MigrationFormDrawer({
  open,
  onClose,
}: MigrationFormDrawerProps) {
  const navigate = useNavigate()
  const { params, getParamsUpdater } = useParams<FormValues>(defaultValues)
  const [error, setError] = useState<{ title: string; message: string } | null>(
    null
  )
  // Theses are the errors that will be displayed on the form
  const { params: fieldErrors, getParamsUpdater: getFieldErrorsUpdater } =
    useParams<FieldErrors>({})
  const queryClient = useQueryClient()

  // Migration Options - Checked or Unchecked state
  const {
    params: selectedMigrationOptions,
    getParamsUpdater: updateSelectedMigrationOptions,
  } = useParams<SelectedMigrationOptionsType>(defaultMigrationOptions)

  // Form Statuses
  const [validatingVmwareCreds, setValidatingVmwareCreds] = useState(false)
  const [validatingOpenstackCreds, setValidatingOpenstackCreds] =
    useState(false)
  const [submitting, setSubmitting] = useState(false)

  // Migration Resources
  const [vmwareCredentials, setVmwareCredentials] = useState<
    VMwareCreds | undefined
  >(undefined)
  const [openstackCredentials, setOpenstackCredentials] = useState<
    OpenstackCreds | undefined
  >(undefined)
  const [migrationTemplate, setMigrationTemplate] = useState<
    MigrationTemplate | undefined
  >(undefined)
  const [migrationPlan, setMigrationPlan] = useState<MigrationPlan | undefined>(
    undefined
  )
  const [migrations, setMigrations] = useState<Migration[] | undefined>(
    undefined
  )

  const [loadingVms, setLoadingVms] = useState(!isNilOrEmpty(migrationTemplate) && migrationTemplate?.status === undefined)

  const vmwareCredsValidated =
    vmwareCredentials?.status?.vmwareValidationStatus === "Succeeded"

  const openstackCredsValidated =
    openstackCredentials?.status?.openstackValidationStatus === "Succeeded"

  // Polling Conditions
  const shouldPollVmwareCreds =
    !!vmwareCredentials?.metadata?.name &&
    vmwareCredentials?.status === undefined

  const shouldPollOpenstackCreds =
    !!openstackCredentials?.metadata?.name &&
    openstackCredentials?.status === undefined

  const shouldPollMigrationTemplate =
    !!migrationTemplate?.metadata?.name &&
    migrationTemplate?.status === undefined


  const shouldPollMigrationPlan =
    !!migrationPlan?.metadata?.name && migrationPlan?.status === undefined

  useEffect(() => {
    const postCreds = async () => {
      try {
        setValidatingVmwareCreds(true)
        const body = createVmwareCredsJson(params.vmwareCreds)
        const response = await postVmwareCredentials(body)
        setVmwareCredentials(response)
      } catch {
        getFieldErrorsUpdater("vmwareCreds")(
          "Error validating VMware credentials"
        )
        setValidatingVmwareCreds(false)
      }
    }
    if (isNilOrEmpty(params.vmwareCreds)) return
    // Reset the VMwareCreds object if the user changes the credentials
    setVmwareCredentials(undefined)
    getFieldErrorsUpdater("vmwareCreds")("")
    postCreds()
  }, [params.vmwareCreds, getFieldErrorsUpdater])

  useEffect(() => {
    const postCreds = async () => {
      setValidatingOpenstackCreds(true)
      try {
        const body = createOpenstackCredsJson(params.openstackCreds)
        const response = await postOpenstackCredentials(body)
        setOpenstackCredentials(response)
      } catch (err) {
        console.error("Error validating Openstack credentials", err)
        getFieldErrorsUpdater("openstackCreds")(
          "Error validating Openstack credentials"
        )
        setValidatingOpenstackCreds(false)
      }
    }

    if (isNilOrEmpty(params.openstackCreds)) return
    // Reset the OpenstackCreds object if the user changes the credentials
    setOpenstackCredentials(undefined)
    getFieldErrorsUpdater("openstackCreds")("")
    postCreds()
  }, [params.openstackCreds, getFieldErrorsUpdater])

  useEffect(() => {
    const createMigrationTemplate = async () => {
      const body = createMigrationTemplateJson({
        datacenter: params.vmwareCreds?.datacenter,
        vmwareRef: vmwareCredentials?.metadata.name,
        openstackRef: openstackCredentials?.metadata.name,
      })
      const response = await postMigrationTemplate(body)
      setMigrationTemplate(response)
    }
    // Once the Openstack and VMware creds are validated, create the migration template

    if (!vmwareCredsValidated || !openstackCredsValidated) return
    createMigrationTemplate()
  }, [
    vmwareCredsValidated,
    openstackCredsValidated,
    params.vmwareCreds?.datacenter,
    vmwareCredentials?.metadata.name,
    openstackCredentials?.metadata.name,
  ])

  useInterval(
    async () => {
      if (shouldPollVmwareCreds) {
        try {
          const response = await getVmwareCredentials(
            vmwareCredentials?.metadata?.name
          )
          setVmwareCredentials(response)
          const validationStatus = response?.status?.vmwareValidationStatus
          if (validationStatus) {
            setValidatingVmwareCreds(false)
            if (validationStatus !== "Succeeded") {
              getFieldErrorsUpdater("vmwareCreds")(
                response?.status?.vmwareValidationMessage
              )
            }
          }
        } catch (err) {
          console.error("Error validating VMware credentials", err)
          getFieldErrorsUpdater("vmwareCreds")(
            "Error validating VMware credentials"
          )
        }
      }
    },
    THREE_SECONDS,
    shouldPollVmwareCreds
  )

  useInterval(
    async () => {
      if (shouldPollOpenstackCreds) {
        try {
          const response = await getOpenstackCredentials(
            openstackCredentials?.metadata?.name
          )
          setOpenstackCredentials(response)
          const validationStatus = response?.status?.openstackValidationStatus
          if (validationStatus) {
            setValidatingOpenstackCreds(false)
            if (validationStatus !== "Succeeded") {
              getFieldErrorsUpdater("openstackCreds")(
                response?.status?.openstackValidationMessage
              )
            }
          }
          setValidatingOpenstackCreds(false)
        } catch (err) {
          console.error("Error validating Openstack credentials", err)
          getFieldErrorsUpdater("openstackCreds")(
            "Error validating Openstack credentials"
          )
          setValidatingOpenstackCreds(false)
        }
      }
    },
    THREE_SECONDS,
    shouldPollOpenstackCreds
  )


  const fetchMigrationTemplate = async () => {
    try {
      setLoadingVms(true)

      const updatedMigrationTemplate = await getMigrationTemplate(
        migrationTemplate?.metadata?.name
      )
      setMigrationTemplate(updatedMigrationTemplate)
      setLoadingVms(false)
    } catch (err) {
      console.error("Error retrieving migration templates", err)
      getFieldErrorsUpdater("migrationTemplate")(
        "Error retrieving migration templates"
      )
      setLoadingVms(false)
    }
  }

  const refreshMigrationTemplate = async () => {
    try {
      setLoadingVms(true)

      const currentRefresh = migrationTemplate?.metadata?.labels?.refresh || "0"
      const nextRefreshValue = (parseInt(currentRefresh) + 1).toString()

      await patchMigrationTemplate(migrationTemplate?.metadata?.name, {
        metadata: {
          labels: {
            refresh: nextRefreshValue,
          },
        },
      })

      // Wait for 20 seconds before fetching, as the VM statuses are not updated immediately
      await new Promise(resolve => setTimeout(resolve, TWENTY_SECONDS))
      await fetchMigrationTemplate()

    } catch (err) {
      console.error("Error refreshing migration template", err)
      getFieldErrorsUpdater("migrationTemplate")(
        "Error refreshing migration template"
      )
      setLoadingVms(false)
    }
  }


  useInterval(
    async () => {
      if (shouldPollMigrationTemplate) {
        try {
          fetchMigrationTemplate()
        } catch (err) {
          console.error("Error retrieving migration templates", err)
          getFieldErrorsUpdater("migrationTemplate")(
            "Error retrieving migration templates"
          )
        }
      }
    },
    THREE_SECONDS,
    shouldPollMigrationTemplate
  )

  useEffect(() => {
    if (vmwareCredsValidated && openstackCredsValidated) return
    // Reset all the migration resources if the user changes the credentials
    setMigrationTemplate(undefined)
  }, [vmwareCredsValidated, openstackCredsValidated])

  const availableVmwareNetworks = useMemo(() => {
    if (params.vms === undefined) return []
    return uniq(flatten(params.vms.map((vm) => vm.networks || []))).sort(
      stringsCompareFn
    )
  }, [params.vms])

  const availableVmwareDatastores = useMemo(() => {
    if (params.vms === undefined) return []
    return uniq(flatten(params.vms.map((vm) => vm.datastores || []))).sort(
      stringsCompareFn
    )
  }, [params.vms])

  const createNetworkMapping = async (networkMappingParams) => {
    const body = createNetworkMappingJson({
      networkMappings: networkMappingParams,
    })

    try {
      const data = postNetworkMapping(body)
      return data
    } catch (err) {
      setError({
        title: "Error creating network mapping",
        message: axios.isAxiosError(err) ? err?.response?.data?.message : "",
      })
    }
  }

  const createStorageMapping = async (storageMappingsParams) => {
    const body = createStorageMappingJson({
      storageMappings: storageMappingsParams,
    })
    try {
      const data = postStorageMapping(body)
      return data
    } catch (err) {
      console.error("Error creating storage mapping", err)
      setError({
        title: "Error creating storage mapping",
        message: axios.isAxiosError(err) ? err?.response?.data?.message : "",
      })
    }
  }

  const updateMigrationTemplate = async (
    migrationTemplate,
    networkMappings,
    storageMappings
  ) => {
    const migrationTemplateName = migrationTemplate?.metadata?.name
    const updatedMigrationTemplateFields = {
      spec: {
        networkMapping: networkMappings.metadata.name,
        storageMapping: storageMappings.metadata.name,
        ...(selectedMigrationOptions.osType &&
          params.osType !== OS_TYPES.AUTO_DETECT && {
          osType: params.osType,
        }),
      },
    }
    try {
      const data = await patchMigrationTemplate(
        migrationTemplateName,
        updatedMigrationTemplateFields
      )
      return data
    } catch (err) {
      setError({
        title: "Error updating migration template",
        message: axios.isAxiosError(err) ? err?.response?.data?.message : "",
      })
    }
  }

  const createMigrationPlan = async (updatedMigrationTemplate) => {
    const vmsToMigrate = (params.vms || []).map((vm) => vm.name)
    const migrationFields = {
      migrationTemplateName: updatedMigrationTemplate?.metadata?.name,
      virtualmachines: vmsToMigrate,
      // Optional Migration Params
      type:
        selectedMigrationOptions.dataCopyMethod && params.dataCopyMethod
          ? params.dataCopyMethod
          : "hot",
      ...(selectedMigrationOptions.dataCopyStartTime &&
        params?.dataCopyStartTime && {
        dataCopyStart: params.dataCopyStartTime,
      }),
      ...(selectedMigrationOptions.cutoverOption &&
        params.cutoverOption === CUTOVER_TYPES.TIME_WINDOW &&
        params.cutoverStartTime && { vmCutoverStart: params.cutoverStartTime }),
      ...(selectedMigrationOptions.cutoverOption &&
        params.cutoverOption === CUTOVER_TYPES.TIME_WINDOW &&
        params.cutoverEndTime && { vmCutoverEnd: params.cutoverEndTime }),
      retry: params.retryOnFailure,
    }
    const body = createMigrationPlanJson(migrationFields)
    try {
      const data = await postMigrationPlan(body)
      return data
    } catch (err) {
      // Handle error
      console.error("Error creating migration plan", err)
      setError({
        title: "Error creating migration plan",
        message: axios.isAxiosError(err) ? err?.response?.data?.message : "",
      })
    }
  }

  const handleSubmit = async () => {
    setSubmitting(true)
    setError(null)

    // Create NetworkMapping
    const networkMappings = await createNetworkMapping(params.networkMappings)

    // Create StorageMapping
    const storageMappings = await createStorageMapping(params.storageMappings)

    if (!networkMappings || !storageMappings) {
      setSubmitting(false)
      return
    }

    // Update MigrationTemplate with NetworkMapping and StorageMapping resource names
    const updatedMigrationTemplate = await updateMigrationTemplate(
      migrationTemplate,
      networkMappings,
      storageMappings
    )

    // Create MigrationPlan
    const migrationPlanResource = await createMigrationPlan(
      updatedMigrationTemplate
    )
    setMigrationPlan(migrationPlanResource)
  }

  useInterval(
    async () => {
      if (shouldPollMigrationPlan) {
        try {
          const response = await getMigrations(migrationPlan?.metadata?.name)
          setMigrations(response)
        } catch (error) {
          console.error("Error getting MigrationPlan", { error })
          setSubmitting(false)
        }
      }
    },
    THREE_SECONDS,
    shouldPollMigrationPlan
  )

  useEffect(() => {
    if (migrations && migrations.length > 0 && !error) {
      setSubmitting(false)
      queryClient.invalidateQueries({ queryKey: MIGRATIONS_QUERY_KEY })
      onClose()
      navigate("/dashboard")
    }
  }, [migrations, error, onClose, navigate, queryClient])

  // Validate Selected Migration Options
  const migrationOptionValidated = useMemo(
    () =>
      Object.keys(selectedMigrationOptions).every((key) => {
        if (selectedMigrationOptions[key]) {
          if (
            key === "cutoverOption" &&
            params.cutoverOption === CUTOVER_TYPES.TIME_WINDOW
          ) {
            return (
              params.cutoverStartTime &&
              params.cutoverEndTime &&
              !fieldErrors["cutoverStartTime"] &&
              !fieldErrors["cutoverEndTime"]
            )
          }
          return params?.[key] && !fieldErrors[key]
        }
        return true
      }),
    [selectedMigrationOptions, params, fieldErrors]
  )

  const disableSubmit =
    !vmwareCredsValidated ||
    !openstackCredsValidated ||
    isNilOrEmpty(params.vms) ||
    isNilOrEmpty(params.networkMappings) ||
    isNilOrEmpty(params.storageMappings) ||
    !migrationOptionValidated

  const sortedOpenstackNetworks = useMemo(
    () =>
      (migrationTemplate?.status?.openstack?.networks || []).sort(
        stringsCompareFn
      ),
    [migrationTemplate?.status?.openstack?.networks]
  )
  const sortedOpenstackVolumeTypes = useMemo(
    () =>
      (migrationTemplate?.status?.openstack?.volumeTypes || []).sort(
        stringsCompareFn
      ),
    [migrationTemplate?.status?.openstack?.volumeTypes]
  )

  const handleClose = async () => {
    try {

      setMigrationTemplate(undefined)
      setVmwareCredentials(undefined)
      setOpenstackCredentials(undefined)
      setError(null)


      onClose()
      // Delete migration template if it exists
      if (migrationTemplate?.metadata?.name) {
        await deleteMigrationTemplate(migrationTemplate.metadata.name)
      }

      // Delete VMware credentials if they exist
      if (vmwareCredentials?.metadata?.name) {
        await deleteVmwareCredentials(vmwareCredentials.metadata.name)
      }

      // Delete OpenStack credentials if they exist
      if (openstackCredentials?.metadata?.name) {
        await deleteOpenstackCredentials(openstackCredentials.metadata.name)
      }

    } catch (err) {
      console.error("Error cleaning up resources", err)
      onClose()
    }
  }

  return (
    <StyledDrawer
      anchor="right"
      open={open}
      onClose={handleClose}
      ModalProps={{ keepMounted: false }}
    >
      <Header title="Migration Form" />
      <DrawerContent>
        <Box sx={{ display: "grid", gap: 4 }}>
          {error && (
            <Alert severity="error">
              <AlertTitle>{error.title}</AlertTitle>
              {error.message}
            </Alert>
          )}
          {/* Step 1 */}
          <SourceAndDestinationEnvStep
            params={params}
            onChange={getParamsUpdater}
            errors={fieldErrors}
            validatingVmwareCreds={validatingVmwareCreds}
            validatingOpenstackCreds={validatingOpenstackCreds}
            vmwareCredsValidated={
              vmwareCredentials?.status?.vmwareValidationStatus === "Succeeded"
            }
            openstackCredsValidated={
              openstackCredentials?.status?.openstackValidationStatus ===
              "Succeeded"
            }
          />
          {/* Step 2 */}
          <VmsSelectionStep
            vms={migrationTemplate?.status?.vmware || []}
            onChange={getParamsUpdater}
            error={fieldErrors["vms"]}
            loadingVms={loadingVms}
            onRefresh={refreshMigrationTemplate}
          />
          {/* Step 3 */}
          <NetworkAndStorageMappingStep
            vmwareNetworks={availableVmwareNetworks}
            vmWareStorage={availableVmwareDatastores}
            openstackNetworks={sortedOpenstackNetworks}
            openstackStorage={sortedOpenstackVolumeTypes}
            params={params}
            onChange={getParamsUpdater}
            networkMappingError={fieldErrors["networksMapping"]}
            storageMappingError={fieldErrors["storageMapping"]}
          />
          {/* Step 4 */}
          <MigrationOptions
            params={params}
            onChange={getParamsUpdater}
            selectedMigrationOptions={selectedMigrationOptions}
            updateSelectedMigrationOptions={updateSelectedMigrationOptions}
            errors={fieldErrors}
            getErrorsUpdater={getFieldErrorsUpdater}
          />
        </Box>
      </DrawerContent>
      <Footer
        submitButtonLabel={"Start Migration"}
        onClose={handleClose}
        onSubmit={handleSubmit}
        disableSubmit={disableSubmit || submitting}
        submitting={submitting}
      />
    </StyledDrawer>
  )
}
