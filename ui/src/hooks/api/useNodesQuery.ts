import { useQuery } from "@tanstack/react-query"
import { getNodes } from "src/api/nodes/nodeMappings"
import { TWENTY_SECONDS } from "src/constants"

export const NODES_QUERY_KEY = ["nodes"]

export const useNodesQuery = () => {
  return useQuery({
    queryKey: NODES_QUERY_KEY,
    queryFn: () => getNodes(),
    refetchInterval: TWENTY_SECONDS,
  })
}
