import React, { useState } from 'react'
import { Layout, Menu, Breadcrumb, Select, Space } from 'antd'
import { SettingOutlined, AppstoreOutlined, DashboardOutlined } from '@ant-design/icons'
import ConfigTree from './components/ConfigTree'
import ConfigEditor from './components/ConfigEditor'
import ConfigItemList from './components/ConfigItemList'
import Dashboard from './components/Dashboard'
import './App.css'

const { Header, Sider, Content } = Layout
const { Option } = Select

function App() {
  const [selectedKey, setSelectedKey] = useState('configs')
  const [selectedConfig, setSelectedConfig] = useState(null)
  const [selectedGroup, setSelectedGroup] = useState(null)
  const [selectedNamespace, setSelectedNamespace] = useState(null)
  const [viewMode, setViewMode] = useState('editor')
  const [environment, setEnvironment] = useState('dev')
  const [treeRefreshKey, setTreeRefreshKey] = useState(0)

  const menuItems = [
    { key: 'configs', icon: <SettingOutlined />, label: '配置管理' },
    { key: 'dashboard', icon: <DashboardOutlined />, label: '监控看板' }
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
          onClick={({ key }) => setSelectedKey(key)}
          style={{ flex: 1, minWidth: 0 }}
        />
        <Space className="env-switch">
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
            />
          </Sider>
        )}
        <Content className="main-content">
          <Breadcrumb style={{ marginBottom: 16 }}>
            <Breadcrumb.Item>首页</Breadcrumb.Item>
            <Breadcrumb.Item>
              {selectedKey === 'configs' ? '配置管理' : '监控看板'}
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
              />
            ) : (
              <ConfigEditor
                config={selectedConfig}
                environment={environment}
                onConfigChange={handleRefreshTree}
              />
            )
          ) : (
            <Dashboard environment={environment} />
          )}
        </Content>
      </Layout>
    </Layout>
  )
}

export default App
