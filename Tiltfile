load('ext://min_k8s_version', 'min_k8s_version')
min_k8s_version('1.18.0')

trigger_mode(TRIGGER_MODE_MANUAL)

load('ext://namespace', 'namespace_create')
namespace_create('badgr')
k8s_resource(
  new_name = 'namespace',
  objects = ['badgr:namespace'],
  labels = ['badgr']
)

docker_build(
  'brigadecore/badgr', '.',
  only = [
    'internal/',
    'config.go',
    'go.mod',
    'go.sum',
    'main.go'
  ],
  ignore = ['**/*_test.go']
)
k8s_resource(
  workload = 'badgr',
  port_forwards = '31700:8080',
  resource_deps = ['redis'],
  labels = ['badgr']
)

k8s_resource(
  workload = 'badgr-redis-master',
  new_name = 'redis',
  labels = ['badgr']
)
k8s_resource(
  workload = 'redis',
  objects = [
    'badgr-redis:serviceaccount',
    'badgr-redis-configuration:configmap',
    'badgr-redis-health:configmap',
    'badgr-redis-scripts:configmap',
    'badgr-redis:secret',
  ],
)

k8s_yaml(
  helm(
    './charts/badgr',
    name = 'badgr',
    namespace = 'badgr',
    set = ['tls.enabled=false']
  )
)
