export interface AuditAction {
  action: string
  name: string
  namespace: string
  details?: string
}

export interface AuditLog {
  id: string
  user_email: string
  action: AuditAction
  admin_info: boolean
  timestamp: string
  request_method?: string
  request_path?: string
  client_ip?: string
  user_agent?: string
}

export interface GetAuditLogsResponse {
  audit_logs: AuditLog[]
  count: number
  total: number
  page: number
}

export const AUDIT_ACTION_TYPES = [
  'create',
  'update',
  'delete',
  'login',
  'register',
] as const

export type AuditActionType = (typeof AUDIT_ACTION_TYPES)[number]
