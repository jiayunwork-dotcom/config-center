-- 租户表
CREATE TABLE IF NOT EXISTS tenants (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    display_name VARCHAR(200),
    max_namespaces INTEGER NOT NULL DEFAULT 10,
    max_config_items INTEGER NOT NULL DEFAULT 1000,
    max_versions INTEGER NOT NULL DEFAULT 100,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 命名空间表
CREATE TABLE IF NOT EXISTS namespaces (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, name)
);

-- 分组表
CREATE TABLE IF NOT EXISTS groups (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    namespace_id BIGINT NOT NULL REFERENCES namespaces(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(namespace_id, name)
);

-- 配置项表
CREATE TABLE IF NOT EXISTS config_items (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    namespace_id BIGINT NOT NULL REFERENCES namespaces(id) ON DELETE CASCADE,
    group_id BIGINT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    key VARCHAR(255) NOT NULL,
    value TEXT NOT NULL,
    format VARCHAR(20) NOT NULL DEFAULT 'json', -- json, yaml, properties, toml
    environment VARCHAR(50) NOT NULL DEFAULT 'dev', -- dev, staging, prod
    level VARCHAR(20) NOT NULL DEFAULT 'group', -- public, namespace, group
    schema JSONB,
    current_version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(namespace_id, group_id, key, environment)
);

CREATE INDEX IF NOT EXISTS idx_config_items_namespace ON config_items(namespace_id);
CREATE INDEX IF NOT EXISTS idx_config_items_group ON config_items(group_id);
CREATE INDEX IF NOT EXISTS idx_config_items_env ON config_items(environment);
CREATE INDEX IF NOT EXISTS idx_config_items_level ON config_items(level);

-- 配置版本表
CREATE TABLE IF NOT EXISTS config_versions (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    config_item_id BIGINT NOT NULL REFERENCES config_items(id) ON DELETE CASCADE,
    version INTEGER NOT NULL,
    value TEXT NOT NULL,
    operator VARCHAR(100) NOT NULL DEFAULT 'system',
    change_type VARCHAR(50) NOT NULL DEFAULT 'update', -- create, update, rollback
    diff TEXT,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_config_versions_config ON config_versions(config_item_id);
CREATE INDEX IF NOT EXISTS idx_config_versions_version ON config_versions(config_item_id, version);

-- 灰度发布表
CREATE TABLE IF NOT EXISTS gray_releases (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    config_item_id BIGINT NOT NULL REFERENCES config_items(id) ON DELETE CASCADE,
    target_version INTEGER NOT NULL,
    strategy VARCHAR(20) NOT NULL, -- ip_list, percentage
    ip_list TEXT[],
    percentage INTEGER,
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, running, completed, rolled_back
    pushed_count INTEGER NOT NULL DEFAULT 0,
    total_count INTEGER NOT NULL DEFAULT 0,
    started_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_gray_releases_config ON gray_releases(config_item_id);
CREATE INDEX IF NOT EXISTS idx_gray_releases_status ON gray_releases(status);

-- 客户端连接表（Redis也存一份，这个做持久化记录）
CREATE TABLE IF NOT EXISTS client_connections (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    namespace_id BIGINT NOT NULL REFERENCES namespaces(id) ON DELETE CASCADE,
    client_id VARCHAR(100) NOT NULL,
    ip_address VARCHAR(50),
    connect_type VARCHAR(20) NOT NULL, -- longpoll, websocket
    last_pull_at TIMESTAMP,
    last_push_at TIMESTAMP,
    push_latency_ms INTEGER,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(namespace_id, client_id)
);

CREATE INDEX IF NOT EXISTS idx_client_connections_namespace ON client_connections(namespace_id);

-- 监控统计表
CREATE TABLE IF NOT EXISTS metrics (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    namespace_id BIGINT NOT NULL REFERENCES namespaces(id) ON DELETE CASCADE,
    metric_type VARCHAR(50) NOT NULL, -- pull_qps, push_success_rate, avg_latency
    value DOUBLE PRECISION NOT NULL,
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_metrics_namespace ON metrics(namespace_id, metric_type, timestamp);

-- 插入默认租户
INSERT INTO tenants (name, display_name, max_namespaces, max_config_items, max_versions)
VALUES ('default', '默认租户', 10, 1000, 100)
ON CONFLICT (name) DO NOTHING;

-- 插入公共命名空间（用于公共配置）
INSERT INTO namespaces (tenant_id, name, description)
SELECT id, 'public', '公共配置命名空间' FROM tenants WHERE name = 'default'
ON CONFLICT DO NOTHING;

-- 插入公共分组
INSERT INTO groups (tenant_id, namespace_id, name, description)
SELECT t.id, n.id, 'default', '默认公共分组'
FROM tenants t
JOIN namespaces n ON n.tenant_id = t.id
WHERE t.name = 'default' AND n.name = 'public'
ON CONFLICT DO NOTHING;
