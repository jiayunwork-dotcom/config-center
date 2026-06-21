import React, { useState, useEffect } from 'react'
import { List, Button, Tag, Space, message, Timeline, Select } from 'antd'
import { ClockCircleOutlined, UserOutlined, RollbackOutlined } from '@ant-design/icons'
import dayjs from 'dayjs'
import { configApi } from '../api'

function VersionHistory({ configId, onRollback, onCompare, canEdit = true }) {
  const [versions, setVersions] = useState([])
  const [loading, setLoading] = useState(false)
  const [compareVersion1, setCompareVersion1] = useState(null)
  const [compareVersion2, setCompareVersion2] = useState(null)

  useEffect(() => {
    if (configId) {
      loadVersions()
    }
  }, [configId])

  const loadVersions = async () => {
    setLoading(true)
    try {
      const result = await configApi.getVersions(configId, { page: 1, page_size: 50 })
      setVersions(result.items || [])
    } catch (error) {
      message.error('加载版本历史失败')
    }
    setLoading(false)
  }

  const getChangeTypeColor = (type) => {
    switch (type) {
      case 'create': return 'green'
      case 'update': return 'blue'
      case 'rollback': return 'orange'
      default: return 'default'
    }
  }

  const getChangeTypeText = (type) => {
    switch (type) {
      case 'create': return '创建'
      case 'update': return '更新'
      case 'rollback': return '回滚'
      default: return type
    }
  }

  const handleCompare = async () => {
    if (!compareVersion1 || !compareVersion2) {
      message.warning('请选择两个版本进行对比')
      return
    }

    try {
      const result = await configApi.compareVersions(configId, compareVersion1, compareVersion2)
      onCompare && onCompare(compareVersion1, compareVersion2, result.diff || [])
    } catch (error) {
      message.error('对比失败')
    }
  }

  return (
    <div>
      <Space style={{ marginBottom: 16 }}>
        <Select
          placeholder="选择版本1"
          style={{ width: 150 }}
          value={compareVersion1}
          onChange={setCompareVersion1}
        >
          {versions.map(v => (
            <Select.Option key={v.version} value={v.version}>
              v{v.version}
            </Select.Option>
          ))}
        </Select>
        <span>vs</span>
        <Select
          placeholder="选择版本2"
          style={{ width: 150 }}
          value={compareVersion2}
          onChange={setCompareVersion2}
        >
          {versions.map(v => (
            <Select.Option key={v.version} value={v.version}>
              v{v.version}
            </Select.Option>
          ))}
        </Select>
        <Button type="primary" size="small" onClick={handleCompare}>
          对比
        </Button>
      </Space>

      <Timeline
        mode="left"
        items={versions.map((version, index) => ({
          color: index === 0 ? 'green' : 'blue',
          children: (
            <div style={{ padding: '8px 0' }}>
              <Space direction="vertical" size="small">
                <Space>
                  <Tag color={getChangeTypeColor(version.change_type)}>
                    v{version.version} - {getChangeTypeText(version.change_type)}
                  </Tag>
                  <span><UserOutlined /> {version.operator}</span>
                </Space>
                <span style={{ color: '#999', fontSize: 12 }}>
                  <ClockCircleOutlined /> {dayjs(version.created_at).format('YYYY-MM-DD HH:mm:ss')}
                </span>
                {version.description && (
                  <p style={{ margin: 0, color: '#666' }}>{version.description}</p>
                )}
                <Button
                  type="link"
                  size="small"
                  icon={<RollbackOutlined />}
                  onClick={() => onRollback && onRollback(version.version)}
                  disabled={index === 0 || !canEdit}
                >
                  回滚到此版本
                </Button>
              </Space>
            </div>
          )
        }))}
      />
    </div>
  )
}

export default VersionHistory
