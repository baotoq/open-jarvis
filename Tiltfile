# Prerequisite guard: fail fast if air is not installed
local('which air || (echo "ERROR: air not installed. Run: go install github.com/air-verse/air@latest" && exit 1)', quiet=True)

local_resource(
    'backend',
    serve_cmd='air',
    serve_dir='src/backend',
    deps=['src/backend'],
    ignore=['src/backend/tmp', 'src/backend/data'],
    readiness_probe=probe(
        tcp_socket=tcp_socket_action(port=8888),
        period_secs=2,
        failure_threshold=15,
    ),
    links=[link('http://localhost:8888', 'Backend API')],
    labels=['services'],
)

local_resource(
    'frontend',
    serve_cmd='npm run dev',
    serve_dir='src/frontend',
    deps=['src/frontend'],
    ignore=['src/frontend/.next', 'src/frontend/node_modules'],
    resource_deps=['backend'],
    links=[link('http://localhost:3000', 'Frontend')],
    labels=['services'],
)

local_resource(
    'open-browser',
    cmd='open http://localhost:3000',
    resource_deps=['frontend'],
    auto_init=True,
    labels=['dev'],
)
