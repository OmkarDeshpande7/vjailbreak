import { Paper, styled, IconButton, Tooltip, Button, Box, Typography } from "@mui/material"
import { DataGrid, GridColDef, GridRowSelectionModel, GridToolbarContainer } from "@mui/x-data-grid"
import { useEffect, useState } from "react"
import { useNavigate } from "react-router-dom"
import CustomSearchToolbar from "src/components/grid/CustomSearchToolbar"
import { FIVE_SECONDS, THIRTY_SECONDS } from "src/constants"
import { useMigrationsQuery } from "src/hooks/api/useMigrationsQuery"
import MigrationProgressWithPopover from "./MigrationProgressWithPopover"
import { deleteMigration } from "src/api/migrations/migrations"
import { useQueryClient } from "@tanstack/react-query"
import { MIGRATIONS_QUERY_KEY } from "src/hooks/api/useMigrationsQuery"
import DeleteIcon from '@mui/icons-material/DeleteOutlined';
import DeleteConfirmationDialog from "./DeleteConfirmationDialog"
import { getMigrationPlan, patchMigrationPlan } from "src/api/migration-plans/migrationPlans"
import { Migration } from "src/api/migrations/model"

const STATUS_ORDER = {
  'Running': 0,
  'Failed': 1,
  'Succeeded': 2,
  'Pending': 3
}

const DashboardContainer = styled("div")({
  display: "flex",
  justifyContent: "center",
  width: "100%",
  height: "100%",
  padding: "40px 20px",
  boxSizing: "border-box"
})

const StyledPaper = styled(Paper)({
  width: "100%",
  "& .MuiDataGrid-virtualScroller": {
    overflowX: "hidden"
  }
})

const columns: GridColDef[] = [
  {
    field: "name",
    headerName: "Name",
    valueGetter: (_, row) => row.metadata?.name,
    flex: 2,
  },
  {
    field: "status",
    headerName: "Status",
    valueGetter: (_, row) => row?.status?.phase || "Pending",
    flex: 1,
    sortComparator: (v1, v2) => {
      const order1 = STATUS_ORDER[v1] ?? Number.MAX_SAFE_INTEGER;
      const order2 = STATUS_ORDER[v2] ?? Number.MAX_SAFE_INTEGER;
      return order1 - order2;
    }
  },
  {
    field: "status.conditions",
    headerName: "Progress",
    valueGetter: (_, row) => row.status?.phase,
    flex: 2,
    renderCell: (params) => {
      const phase = params.row?.status?.phase
      const conditions = params.row?.status?.conditions
      return conditions ? (
        <MigrationProgressWithPopover
          phase={phase}
          conditions={params.row?.status?.conditions}
        />
      ) : null
    },
  },
  {
    field: "actions",
    headerName: "Actions",
    flex: 1,
    renderCell: (params) => {
      const phase = params.row?.status?.phase;
      const isDisabled = !phase || phase === "Running" || phase === "Pending";

      return (
        <Tooltip title={isDisabled ? "Cannot delete while migration is in progress" : "Delete migration"} >
          <span>
            <IconButton
              onClick={(e) => {
                e.stopPropagation();
                params.row.onDelete(params.row.metadata?.name);
              }}
              disabled={isDisabled}
              size="small"
              sx={{
                cursor: isDisabled ? 'not-allowed' : 'pointer',
                position: 'relative'
              }}
            >
              <DeleteIcon />
            </IconButton>
          </span>
        </Tooltip>
      );
    },
  },
]

const paginationModel = { page: 0, pageSize: 25 }

interface CustomToolbarProps {
  numSelected: number;
  onDeleteSelected: () => void;
}

const CustomToolbar = ({ numSelected, onDeleteSelected }: CustomToolbarProps) => {
  return (
    <GridToolbarContainer
      sx={{
        p: 2,
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center'
      }}
    >
      <div>
        <Typography variant="h6" component="h2">
          Migrations
        </Typography>
      </div>
      <Box sx={{ display: 'flex', gap: 2 }}>
        {numSelected > 0 && (
          <Button
            variant="outlined"
            color="error"
            startIcon={<DeleteIcon />}
            onClick={onDeleteSelected}
            sx={{ height: 40 }}
          >
            Delete Selected ({numSelected})
          </Button>
        )}
        <CustomSearchToolbar
          hideTitle
          placeholder="Search by Name, Status, or Progress"
        />
      </Box>
    </GridToolbarContainer>
  );
};

export default function Dashboard() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [deleteDialog, setDeleteDialog] = useState<{ open: boolean, migrationName: string | null, selectedMigrations?: Migration[] }>({
    open: false,
    migrationName: null
  });
  const [isDeleting, setIsDeleting] = useState(false);
  const [deleteError, setDeleteError] = useState<string | null>(null);
  const [selectedRows, setSelectedRows] = useState<GridRowSelectionModel>([]);

  const { data: migrations } = useMigrationsQuery(undefined, {
    refetchInterval: (query) => {
      const migrations = query?.state?.data || []
      const hasPendingMigration = !!migrations.find(
        (m) => m.status === undefined
      )
      return hasPendingMigration ? FIVE_SECONDS : THIRTY_SECONDS
    },
  })

  const handleDeleteClick = (migrationName: string) => {
    setDeleteError(null);
    setDeleteDialog({
      open: true,
      migrationName
    });
  };

  const handleDeleteClose = () => {
    setDeleteError(null);
    setDeleteDialog({
      open: false,
      migrationName: null
    });
  };

  const handleSelectionChange = (newSelection: GridRowSelectionModel) => {
    setSelectedRows(newSelection);
  };

  const handleDeleteSelected = () => {
    const selectedMigrations = migrations?.filter(
      m => selectedRows.includes(m.metadata?.name)
    );
    if (!selectedMigrations?.length) return;

    setDeleteDialog({
      open: true,
      migrationName: null,
      selectedMigrations
    });
  };

  const handleDeleteConfirm = async () => {
    if (!deleteDialog.selectedMigrations?.length) return;

    setIsDeleting(true);
    setDeleteError(null);

    const results = await Promise.allSettled(
      deleteDialog.selectedMigrations.map(async (migration) => {
        try {
          await handleDeleteMigration(migration);
          return { success: true, name: migration.metadata.name };
        } catch (error) {
          return {
            success: false,
            name: migration.metadata.name,
            error: error instanceof Error ? error.message : 'Unknown error'
          };
        }
      })
    );

    const failures = results.filter(
      (r): r is PromiseRejectedResult => r.status === 'rejected'
    );

    if (failures.length) {
      setDeleteError(
        `Failed to delete some migrations: ${failures
          .map(f => f.reason.name)
          .join(', ')}`
      );
    }

    queryClient.invalidateQueries({ queryKey: MIGRATIONS_QUERY_KEY });
    setIsDeleting(false);
    handleDeleteClose();
    setSelectedRows([]);
  };

  const handleDeleteMigration = async (migration: Migration) => {
    try {
      const migrationPlan = await getMigrationPlan(migration.spec.migrationPlan)

      const updatedVirtualMachines = migrationPlan.spec.virtualmachines[0].filter(
        vm => vm !== migration.spec.vmName
      )

      await patchMigrationPlan(migration.spec.migrationPlan, {
        spec: {
          virtualmachines: updatedVirtualMachines
        }
      })

      await deleteMigration(migration.metadata.name)

    } catch (error) {
      console.error("Error removing VM from migration plan", error)
    }
  }

  useEffect(() => {
    if (!!migrations && migrations.length === 0) {
      navigate("/onboarding")
    }
  }, [migrations, navigate])

  const migrationsWithActions = migrations?.map(migration => ({
    ...migration,
    onDelete: handleDeleteClick
  })) || []

  const isRowSelectable = (params) => {
    const phase = params.row?.status?.phase;
    return !(!phase || phase === "Running" || phase === "Pending");
  };

  return (
    <DashboardContainer>
      <StyledPaper>
        <DataGrid
          rows={migrationsWithActions}
          columns={columns}
          initialState={{
            pagination: { paginationModel },
            sorting: {
              sortModel: [{ field: 'status', sort: 'asc' }],
            },
          }}
          pageSizeOptions={[25, 50, 100]}
          localeText={{ noRowsLabel: "No Migrations Available" }}
          getRowId={(row) => row.metadata?.name}
          checkboxSelection
          isRowSelectable={isRowSelectable}
          onRowSelectionModelChange={handleSelectionChange}
          rowSelectionModel={selectedRows}
          slots={{
            toolbar: () => (
              <CustomToolbar
                numSelected={selectedRows.length}
                onDeleteSelected={handleDeleteSelected}
              />
            ),
          }}
        />
      </StyledPaper>
      <DeleteConfirmationDialog
        open={deleteDialog.open}
        migrationName={deleteDialog.migrationName}
        selectedMigrations={deleteDialog.selectedMigrations}
        onClose={handleDeleteClose}
        onConfirm={handleDeleteConfirm}
        isDeleting={isDeleting}
        error={deleteError}
      />
    </DashboardContainer>
  )
}
