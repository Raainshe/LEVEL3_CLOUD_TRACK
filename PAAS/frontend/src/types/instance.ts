export interface RedisInstance {
  id: string
  name: string
  namespace: string
  redisReplicas: number
  sentinelReplicas: number
  status: string
  createdAt: string
  updatedAt: string
  externalHost?: string
  externalPort?: number
  redisCli?: string
}

export interface ListInstancesResponse {
  instances: RedisInstance[]
  count: number
}

export interface CreateInstanceRequest {
  name?: string
  namespace?: string
  redisReplicas: number
  sentinelReplicas: number
}
