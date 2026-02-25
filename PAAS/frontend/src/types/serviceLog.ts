export interface ServiceLog {
  id: string
  instance_name: string
  namespace: string
  event_type: string
  from_status?: string
  to_status: string
  message: string
  details?: string
  timestamp: string
}

export interface GetServiceLogsResponse {
  service_logs: ServiceLog[]
  count: number
  total: number
  page: number
}

export interface GetServiceLogsParams {
  page?: number
  instance?: string
  namespace?: string
}

export interface GetInstanceServiceLogsParams {
  page?: number
  namespace?: string
}

export const SERVICE_LOG_EVENT_TYPES = [
  'status_change',
  'failure',
] as const

export type ServiceLogEventType = (typeof SERVICE_LOG_EVENT_TYPES)[number]
