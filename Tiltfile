def deploy_cert_manager(version):
    def deploy(version):
        return 'kubectl apply --validate=false -f https://github.com/cert-manager/cert-manager/releases/download/v'+version+'/cert-manager.yaml'

    def wait():
        return 'kubectl -n cert-manager wait --for=condition=available --timeout=180s --all deployments'
    print("deploy cert manager", version)
    local_resource('deploy-cert-manager', deploy(version))
    local_resource('wait-cert-manager', wait())

def manifests():
    return './bin/controller-gen crd rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases'

def generate():
    return './bin/controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./..."'

def docker_build_and_load():
    def docker_build():
        return 'make docker-build'
    def kind_load():
        return 'kind load docker-image controller:latest'
    def restart():
        return 'kubectl rollout restart deployment growi-manager-controller-manager -n growi-manager-system'
    return docker_build() + '&&' + kind_load() + '&&' + restart()
    
# Deploy Cert Manager
deploy_cert_manager(version='1.17.1')

# Generate manifests and go files
local_resource(
    'make-manifests',
    manifests(),
    deps=["api", "internal", "hooks"],
    ignore=['*/*/zz_generated.deepcopy.go']
)

local_resource(
    'make-generate',
    generate(),
    deps=["api", "hooks"],
    ignore=['*/*/zz_generated.deepcopy.go']
)

# Deploy CRD
local_resource(
    'apply-crd',
    manifests()+' && ./bin/kustomize build config/crd | kubectl apply -f -',
    deps=["api"],
    ignore=['*/*/zz_generated.deepcopy.go']
)

# Deploy manager with kustomize
watch_file('./config/')
k8s_yaml(kustomize('./config/dev', kustomize_bin='./bin/kustomize'))

# Docker build and load
local_resource(
    'rebuild',
    docker_build_and_load(),
    deps=['internal', 'api', 'cmd/main.go'],
    ignore=['*/*/zz_generated.deepcopy.go']
)

# Apply sample
local_resource(
    'apply-sample',
    'kubectl apply -f ./config/samples/app_v1_growi.yaml',
    deps=["./config/samples/app_v1_growi.yaml"]
)

