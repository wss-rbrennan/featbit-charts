{{- define "redis-dotnet-env" }}
{{ if (include "featbit.redis.sentinel.enabled" .) }}
- name: CACHE_TYPE
  value: RedisSentinelCache
- name: REDIS_SENTINEL_HOST_PORT_PAIRS
  value: {{ include "featbit.redis.config.0" . }}
- name: REDIS_SENTINEL_DB
  value: {{ (include "featbit.redis.db" .) | quote }}
- name: REDIS_SENTINEL_MASTER_SET
  value: {{ (include "featbit.redis.sentinel.masterSet" .) | quote }}  
{{ end }}
{{- end }}