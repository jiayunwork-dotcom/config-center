-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(100) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 角色绑定表
CREATE TABLE IF NOT EXISTS user_roles (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    namespace_id BIGINT REFERENCES namespaces(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL, -- admin, editor, viewer
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_user_roles_user ON user_roles(user_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_namespace ON user_roles(namespace_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_user_roles_unique ON user_roles(user_id, COALESCE(namespace_id, 0), role);

-- 审计日志表
CREATE TABLE IF NOT EXISTS audit_logs (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL DEFAULT 1,
    user_id BIGINT,
    username VARCHAR(100),
    action VARCHAR(50) NOT NULL,
    resource_type VARCHAR(50) NOT NULL, -- namespace, group, config, gray
    resource_id BIGINT,
    resource_name VARCHAR(255),
    old_value TEXT,
    new_value TEXT,
    ip_address VARCHAR(50),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_created ON audit_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_user ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource ON audit_logs(resource_type, resource_id);

-- 初始化admin用户: admin/admin123
-- bcrypt hash of "admin123"
INSERT INTO users (username, password_hash)
VALUES ('admin', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy')
ON CONFLICT (username) DO NOTHING;

-- 给admin用户授予全局admin角色（namespace_id为NULL表示全局）
INSERT INTO user_roles (user_id, namespace_id, role)
SELECT id, NULL, 'admin' FROM users WHERE username = 'admin'
ON CONFLICT (user_id, COALESCE(namespace_id, 0), role) DO NOTHING;
