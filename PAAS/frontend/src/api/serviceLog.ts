import { api } from '@/api/client'
import type { GetServiceLogsResponse } from '@/types/serviceLog'

export interface GetServiceLogsParams {
  page?: number
  instance?: string
  namespace?: string
}

export interface GetInstanceServiceLogsParams {
  page?: number
  namespace?: string
}

export function getServiceLogs(
  params: GetServiceLogsParams = {}
): Promise<GetServiceLogsResponse> {
  const search = new URLSearchParams()
  if (params.page != null && params.page > 0) {
    search.set('page', String(params.page))
  }
  if (params.instance) {
    search.set('instance', params.instance)
  }
  if (params.namespace) {
    search.set('namespace', params.namespace)
  }
  const query = search.toString()
  const url = query ? `/api/service-logs?${query}` : '/api/service-logs'
  return api.get<GetServiceLogsResponse>(url).then((res) => res.data)
}

export function getInstanceServiceLogs(
  instanceId: string,
  params: GetInstanceServiceLogsParams = {}
): Promise<GetServiceLogsResponse> {
  const search = new URLSearchParams()
  if (params.page != null && params.page > 0) {
    search.set('page', String(params.page))
  }
  if (params.namespace) {
    search.set('namespace', params.namespace)
  }
  const query = search.toString()
  const url = query
    ? `/api/instances/${encodeURIComponent(instanceId)}/service-logs?${query}`
    : `/api/instances/${encodeURIComponent(instanceId)}/service-logs`
  return api.get<GetServiceLogsResponse>(url).then((res) => res.data)
}
