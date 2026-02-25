import { api } from '@/api/client'
import type { GetAuditLogsResponse } from '@/types/audit'

export interface GetAuditLogsParams {
  page?: number
  type?: string
  admin_only?: boolean
}

export function getAuditLogs(
  params: GetAuditLogsParams = {}
): Promise<GetAuditLogsResponse> {
  const search = new URLSearchParams()
  if (params.page != null && params.page > 0) {
    search.set('page', String(params.page))
  }
  if (params.type) {
    search.set('type', params.type)
  }
  if (params.admin_only === true) {
    search.set('admin_only', 'true')
  }
  const query = search.toString()
  const url = query ? `/api/audit-logs?${query}` : '/api/audit-logs'
  return api.get<GetAuditLogsResponse>(url).then((res) => res.data)
}
