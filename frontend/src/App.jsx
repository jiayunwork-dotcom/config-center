import React, { useState, useEffect } from 'react'
import { BrowserRouter as Router, Routes, Route, Navigate, useNavigate } from 'react-router-dom'
import { Layout, Menu, Breadcrumb, Select, Space, Button, Dropdown, Avatar, Tag, message } from 'antd'
import { SettingOutlined, AppstoreOutlined, DashboardOutlined, UserOutlined, LogoutOutlined, HistoryOutlined, SafetyOutlined, AuditOutlined } from '@ant-design/icons'
import ConfigTree from './components/ConfigTree'
import ConfigEditor from './components/ConfigEditor'
import ConfigItemList from './components/ConfigItemList'
import Dashboard from './components/Dashboard'
import LoginPage from './components/LoginPage'
import AuditLog from './components/AuditLog'
import ApprovalManagement from './components/ApprovalManagement'
import { getToken, getUser, setUser, clearToken, authApi, permissionApi } from './api'
import './App.css'

const { Header, Sider, Content } = Layout
const { Option } = Select

function AuthGuard({ children }) {
  const navigate = useNavigate()
  const token = getToken()
  if (!token) {
    return <Navigate to="/login" replace />
  }
  return children
}

function AppLayout() {
  const navigate = useNavigate()
  const [selectedKey, setSelectedKey] = useState('configs')
  const [selectedConfig, setSelectedConfig] = useState(null)
  const [selectedGroup, setSelectedGroup] = useState(null)
  const [selectedNamespace, setSelectedNamespace] = useState(null)
  const [viewMode, setViewMode] = useState('editor')
  const [environment, setEnvironment] = useState('dev')
  const [treeRefreshKey, setTreeRefreshKey] = useState(0)
  const [currentUser, setCurrentUser] = useState(getUser())

  useEffect(() => {
    loadMe()
  }, [])

  const loadMe = async () => {
    try {
      const me = await authApi.me()
      setCurrentUser(me)
      setUser(me)
    } catch (e) {}
  }

  const handleLogout = () => {
    clearToken()
    message.success('已退出登录')
    navigate('/login', { replace: true })
  }

  const isAdmin = permissionApi.isAdmin(currentUser)

  const menuItems = [
    { key: 'configs', icon: <SettingOutlined />, label: '配置管理' },
    { key: 'dashboard', icon: <DashboardOutlined />, label: '监控看板' },
    { key: 'audit', icon: <HistoryOutlined />, label: '审计日志' },
    ...(isAdmin ? [{ key: 'approvals', icon: <AuditOutlined />, label: '审批管理' }] : []),
  ]

  const userMenuItems = [
    {
      key: 'user',
      icon: <UserOutlined />,
      label: (
        <span>
          {currentUser?.username || '未登录'}
          {currentUser?.is_global_admin && (
            <Tag color="red" style={{ marginLeft: 8 }}>Admin</Tag>
          )}
        </span>
      ),
      disabled: true,
    },
    { type: 'divider' },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: '退出登录',
      onClick: handleLogout,
    },
  ]

  const handleConfigSelect = (config) => {
    setSelectedConfig(config)
    setViewMode('editor')
  }

  const handleGroupSelect = (group, namespace) => {
    setSelectedGroup(group)
    setSelectedNamespace(namespace)
    setViewMode('list')
  }

  const handleRefreshTree = () => {
    setTreeRefreshKey(prev => prev + 1)
  }

  return (
    <Layout className="app-layout">
      <Header className="header">
        <div className="logo">
          <AppstoreOutlined />
          <span style={{ marginLeft: 8, color: 'white', fontSize: 18, fontWeight: 'bold' }}>
            配置中心
          </span>
        </div>
        <Menu
          theme="dark"
          mode="horizontal"
          selectedKeys={[selectedKey]}
          items={menuItems}
          onClick={({ key }) => { setSelectedKey(key); setSelectedConfig(null); setSelectedGroup(null) }}
          style={{ flex: 1, minWidth: 0 }}
        />
        <Space className="env-switch">
          {selectedKey === 'configs' && (
            <>
              <span style={{ color: 'rgba(255,255,255,0.85)' }}>环境:</span>
              <Select
                value={environment}
                onChange={setEnvironment}
                style={{ width: 120 }}
                size="middle"
              >
                <Option value="dev">开发环境</Option>
                <Option value="staging">测试环境</Option>
                <Option value="prod">生产环境</Option>
              </Select>
            </>
          )}
          <Dropdown
            menu={{ items: userMenuItems }}
            placement="bottomRight"
            trigger={['click']}
          >
            <Space style={{ cursor: 'pointer', padding: '0 8px' }}>
              <Avatar
                size="small"
                icon={<UserOutlined />}
                style={{ backgroundColor: '#1677ff' }}
              />
              <span style={{ color: 'rgba(255,255,255,0.85)' }}>
                {currentUser?.username || '未登录'}
              </span>
            </Space>
          </Dropdown>
        </Space>
      </Header>
      <Layout>
        {selectedKey === 'configs' && (
          <Sider width={280} className="sider">
            <ConfigTree
              key={treeRefreshKey}
              environment={environment}
              onSelect={handleConfigSelect}
              onGroupSelect={handleGroupSelect}
              isAdmin={isAdmin}
              currentUser={currentUser}
            />
          </Sider>
        )}
        <Content className="main-content">
          <Breadcrumb style={{ marginBottom: 16 }}>
            <Breadcrumb.Item>首页</Breadcrumb.Item>
            <Breadcrumb.Item>
              {selectedKey === 'configs' ? '配置管理' : selectedKey === 'dashboard' ? '监控看板' : selectedKey === 'approvals' ? '审批管理' : '审计日志'}
            </Breadcrumb.Item>
            {selectedKey === 'configs' && selectedNamespace && (
              <Breadcrumb.Item>{selectedNamespace.name}</Breadcrumb.Item>
            )}
            {selectedKey === 'configs' && selectedGroup && viewMode === 'list' && (
              <Breadcrumb.Item>{selectedGroup.name}</Breadcrumb.Item>
            )}
            {selectedKey === 'configs' && selectedConfig && viewMode === 'editor' && (
              <Breadcrumb.Item>{selectedConfig.key}</Breadcrumb.Item>
            )}
          </Breadcrumb>
          {selectedKey === 'configs' ? (
            viewMode === 'list' ? (
              <ConfigItemList
                namespace={selectedNamespace}
                group={selectedGroup}
                environment={environment}
                onRefresh={handleRefreshTree}
                onSelectConfig={handleConfigSelect}
                currentUser={currentUser}
              />
            ) : (
              <ConfigEditor
                config={selectedConfig}
                environment={environment}
                onConfigChange={handleRefreshTree}
                currentUser={currentUser}
              />
            )
          ) : selectedKey === 'dashboard' ? (
            <Dashboard environment={environment} />
          ) : selectedKey === 'approvals' ? (
            <ApprovalManagement />
          ) : (
            <AuditLog />
          )}
        </Content>
      </Layout>
    </Layout>
  )
}

function App() {
  return (
    <Router>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route
          path="/*"
          element={
            <AuthGuard>
              <AppLayout />
            </AuthGuard>
          }
        />
      </Routes>
    </Router>
  )
}

export default App
