import React, { useState, useEffect } from 'react'
import { Card, Button, Space, Table, Checkbox, Modal, Select, message, Tag } from 'antd'
import { 
  DeleteOutlined, 
  CopyOutlined, 
  ExportOutlined, 
  UnorderedListOutlined,
  CloseOutlined
} from '@ant-design/icons'
import { configApi } from '../api'

const { Option } = Select

function ConfigItemList({ namespace, group, environment, onRefresh, onSelectConfig }) {
  const [configItems, setConfigItems] = useState([])
  const [loading, setLoading] = useState(false)
  const [batchMode, setBatchMode] = useState(false)
  const [selectedRowKeys, setSelectedRowKeys] = useState([])
  const [deleteModalVisible, setDeleteModalVisible] = useState(false)
  const [copyModalVisible, setCopyModalVisible] = useState(false)
  const [targetEnv, setTargetEnv] = useState('')
  const [copyResult, setCopyResult] = useState(null)

  useEffect(() => {
    if (group && group.id) {
      loadConfigItems()
    }
  }, [group, environment])

  const loadConfigItems = async () => {
    if (!group || !group.id) return
    setLoading(true)
    try {
      const items = await configApi.list({
        namespace_id: namespace?.id,
        group_id: group.id,
        environment
      })
      setConfigItems(items)
    } catch (error) {
      message.error('加载配置项失败')
    } finally {
      setLoading(false)
    }
  }

  const handleBatchModeToggle = () => {
    setBatchMode(!batchMode)
    setSelectedRowKeys([])
  }

  const handleSelectAll = (e) => {
    if (e.target.checked) {
      setSelectedRowKeys(configItems.map(item => item.id))
    } else {
      setSelectedRowKeys([])
    }
  }

  const handleSelectItem = (id, checked) => {
    if (checked) {
      setSelectedRowKeys([...selectedRowKeys, id])
    } else {
      setSelectedRowKeys(selectedRowKeys.filter(key => key !== id))
    }
  }

  const getSelectedItems = () => {
    return configItems.filter(item => selectedRowKeys.includes(item.id))
  }

  const handleDeleteClick = () => {
    if (selectedRowKeys.length === 0) {
      message.warning('请先选择要删除的配置项')
      return
    }
    setDeleteModalVisible(true)
  }

  const handleDeleteConfirm = async () => {
    try {
      await configApi.batchDelete(selectedRowKeys)
      message.success(`成功删除 ${selectedRowKeys.length} 条配置项`)
      setDeleteModalVisible(false)
      setSelectedRowKeys([])
      setBatchMode(false)
      loadConfigItems()
      onRefresh && onRefresh()
    } catch (error) {
      message.error('删除失败')
    }
  }

  const handleCopyClick = () => {
    if (selectedRowKeys.length === 0) {
      message.warning('请先选择要复制的配置项')
      return
    }
    const otherEnvs = ['dev', 'staging', 'prod'].filter(e => e !== environment)
    setTargetEnv(otherEnvs[0] || '')
    setCopyResult(null)
    setCopyModalVisible(true)
  }

  const handleCopyConfirm = async () => {
    if (!targetEnv) {
      message.warning('请选择目标环境')
      return
    }
    try {
      const result = await configApi.batchCopy(selectedRowKeys, targetEnv)
      setCopyResult(result)
      loadConfigItems()
      onRefresh && onRefresh()
    } catch (error) {
      message.error('复制失败')
    }
  }

  const handleExport = () => {
    if (selectedRowKeys.length === 0) {
      message.warning('请先选择要导出的配置项')
      return
    }

    const items = getSelectedItems()
    const exportData = items.map(item => ({
      key: item.key,
      value: item.value,
      format: item.format,
      environment: item.environment
    }))

    const jsonStr = JSON.stringify(exportData, null, 2)
    const blob = new Blob([jsonStr], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    
    const fileName = `${namespace?.name || 'unknown'}-${group?.name || 'unknown'}-${environment}-export.json`
    
    const link = document.createElement('a')
    link.href = url
    link.download = fileName
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(url)

    message.success(`已导出 ${items.length} 条配置项`)
  }

  const columns = [
    {
      title: '配置键',
      dataIndex: 'key',
      key: 'key',
      render: (text, record) => (
        <a onClick={() => onSelectConfig && onSelectConfig(record)}>{text}</a>
      )
    },
    {
      title: '格式',
      dataIndex: 'format',
      key: 'format',
      width: 100,
      render: (format) => <Tag color="blue">{format.toUpperCase()}</Tag>
    },
    {
      title: '版本',
      dataIndex: 'current_version',
      key: 'version',
      width: 80,
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180,
      render: (time) => new Date(time).toLocaleString()
    }
  ]

  if (batchMode) {
    columns.unshift({
      title: (
        <Checkbox
        checked={selectedRowKeys.length === configItems.length && configItems.length > 0}
        indeterminate={selectedRowKeys.length > 0 && selectedRowKeys.length < configItems.length}
        onChange={handleSelectAll}
      />
    ),
      dataIndex: 'select',
      key: 'select',
      width: 50,
      render: (_, record) => (
        <Checkbox
          checked={selectedRowKeys.includes(record.id)}
          onChange={(e) => handleSelectItem(record.id, e.target.checked)}
          onClick={(e) => e.stopPropagation()}
        />
      )
    })
  }

  const otherEnvs = ['dev', 'staging', 'prod'].filter(e => e !== environment)
  const envLabels = { dev: '开发环境', staging: '测试环境', prod: '生产环境' }

  return (
    <div className="config-list-container">
      <Card
        size="small"
        title={
          <Space>
            <UnorderedListOutlined />
            <span>{group?.name || '配置项列表'}</span>
            <Tag color="green">{configItems.length} 条</Tag>
          </Space>
        }
        extra={
          <Space>
            {!batchMode ? (
              <Button
                type="primary"
                size="small"
                icon={<UnorderedListOutlined />}
                onClick={handleBatchModeToggle}
                disabled={configItems.length === 0}
              >
                批量操作
              </Button>
            ) : (
              <Button
                size="small"
                icon={<CloseOutlined />}
                onClick={handleBatchModeToggle}
              >
                退出批量
              </Button>
            )}
          </Space>
        }
        style={{ marginBottom: 16 }}
      />

      <Table
        rowKey="id"
        loading={loading}
        dataSource={configItems}
        columns={columns}
        pagination={{ pageSize: 20 }}
        size="small"
        rowClassName={() => 'config-item-row'}
      />

      {batchMode && selectedRowKeys.length > 0 && (
        <div className="batch-action-bar">
          <Space>
            <span>已选择 {selectedRowKeys.length} 项</span>
            <Button
              danger
              size="small"
              icon={<DeleteOutlined />}
              onClick={handleDeleteClick}
            >
              批量删除
            </Button>
            <Button
              size="small"
              icon={<CopyOutlined />}
              onClick={handleCopyClick}
            >
              批量环境复制
            </Button>
            <Button
              size="small"
              icon={<ExportOutlined />}
              onClick={handleExport}
            >
              批量导出
            </Button>
          </Space>
        </div>
      )}

      <Modal
        title="确认批量删除"
        open={deleteModalVisible}
        onOk={handleDeleteConfirm}
        onCancel={() => setDeleteModalVisible(false)}
        okText="确认删除"
        okType="danger"
        width={500}
      >
        <p>即将删除 <strong>{selectedRowKeys.length}</strong> 条配置项：</p>
        <div style={{ maxHeight: 300, overflowY: 'auto', border: '1px solid #f0f0f0', padding: 12, borderRadius: 4 }}>
          {getSelectedItems().map(item => (
            <div key={item.id} style={{ padding: '4px 0' }}>
              • {item.key}
            </div>
          ))}
        </div>
        <p style={{ marginTop: 12, color: '#ff4d4f' }}>
          删除后无法恢复，请谨慎操作！
        </p>
      </Modal>

      <Modal
        title="批量环境复制"
        open={copyModalVisible}
        onOk={copyResult ? () => {
          setCopyModalVisible(false)
          setCopyResult(null)
          setSelectedRowKeys([])
          setBatchMode(false)
        } : handleCopyConfirm}
        onCancel={() => {
          setCopyModalVisible(false)
          setCopyResult(null)
        }}
        okText={copyResult ? '完成' : '确认复制'}
        width={500}
      >
        {!copyResult ? (
          <Space direction="vertical" style={{ width: '100%' }}>
            <p>将选中的 <strong>{selectedRowKeys.length}</strong> 条配置项复制到目标环境：</p>
            <Select
              value={targetEnv}
              onChange={setTargetEnv}
              style={{ width: '100%' }}
              placeholder="请选择目标环境"
            >
              {otherEnvs.map(env => (
                <Option key={env} value={env}>{envLabels[env]}</Option>
              ))}
            </Select>
            <p style={{ color: '#999', fontSize: 12 }}>
              如目标环境已存在相同Key的配置项，将跳过并提示。
            </p>
          </Space>
        ) : (
          <div>
            <p>
              复制完成：成功 <strong style={{ color: '#52c41a' }}>{copyResult.success_count}</strong> 条，
              跳过 <strong style={{ color: '#faad14' }}>{copyResult.skipped_count}</strong> 条，
              失败 <strong style={{ color: '#ff4d4f' }}>{copyResult.failed_count}</strong> 条
            </p>
            <div style={{ maxHeight: 300, overflowY: 'auto', border: '1px solid #f0f0f0', padding: 12, borderRadius: 4 }}>
              {copyResult.results.map((item, index) => (
                <div key={index} style={{ padding: '4px 0', display: 'flex', justifyContent: 'space-between' }}>
                  <span>• {item.key}</span>
                  <Tag color={item.status === 'success' ? 'green' : item.status === 'skipped' ? 'orange' : 'red'}>
                    {item.status === 'success' ? '成功' : item.status === 'skipped' ? '已跳过' : '失败'}
                  </Tag>
                </div>
              ))}
            </div>
          </div>
        )}
      </Modal>
    </div>
  )
}

export default ConfigItemList
