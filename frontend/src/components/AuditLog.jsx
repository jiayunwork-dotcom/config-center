import React, { useState, useEffect } from 'react'
import { Table, Card, Select, Input, Space, Tag, DatePicker } from 'antd'
import { SearchOutlined } from '@ant-design/icons'
import dayjs from 'dayjs'
import { auditApi, userApi } from '../api'

const { Option } = Select
const { RangePicker } = DatePicker

const actionLabels = {
  create: { label: '创建', color: 'green' },
  update: { label: '更新', color: 'blue' },
  delete: { label: '删除', color: 'red' },
  rollback: { label: '回滚', color: 'orange' },
  start: { label: '启动', color: 'cyan' },
  full_push: { label: '全量推送', color: 'purple' },
  grant_role: { label: '授权角色', color: 'geekblue' },
  revoke_role: { label: '撤销角色', color: 'volcano' },
}

const resourceLabels = {
  namespace: '命名空间',
  group: '分组',
  config: '配置项',
  gray: '灰度发布',
  user_role: '用户角色',
}

function AuditLog() {
  const [data, setData] = useState([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [loading, setLoading] = useState(false)
  const [actionFilter, setActionFilter] = useState()
  const [userFilter, setUserFilter] = useState()
  const [users, setUsers] = useState([])
  const [searchText, setSearchText] = useState('')

  useEffect(() => {
    loadUsers()
  }, [])

  useEffect(() => {
    loadData()
  }, [page, pageSize, actionFilter, userFilter])

  const loadUsers = async () => {
    try {
      const list = await userApi.list()
      setUsers(list || [])
    } catch (e) {}
  }

  const loadData = async () => {
    setLoading(true)
    try {
      const params = { page, page_size: pageSize }
      if (actionFilter) params.action = actionFilter
      if (userFilter) params.user_id = userFilter
      const result = await auditApi.list(params)
      setData(result.items || [])
      setTotal(result.total || 0)
    } catch (e) {
    } finally {
      setLoading(false)
    }
  }

  const columns = [
    {
      title: '时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180,
      render: (v) => dayjs(v).format('YYYY-MM-DD HH:mm:ss'),
      sorter: (a, b) => new Date(a.created_at) - new Date(b.created_at),
    },
    {
      title: '操作人',
      dataIndex: 'username',
      key: 'username',
      width: 120,
      render: (v) => v || '-',
    },
    {
      title: '操作类型',
      dataIndex: 'action',
      key: 'action',
      width: 110,
      render: (v) => {
        const info = actionLabels[v] || { label: v, color: 'default' }
        return <Tag color={info.color}>{info.label}</Tag>
      },
      filters: Object.entries(actionLabels).map(([k, v]) => ({ text: v.label, value: k })),
      onFilter: (value, record) => record.action === value,
    },
    {
      title: '资源类型',
      dataIndex: 'resource_type',
      key: 'resource_type',
      width: 110,
      render: (v) => resourceLabels[v] || v,
    },
    {
      title: '资源名称',
      dataIndex: 'resource_name',
      key: 'resource_name',
      width: 200,
      ellipsis: true,
      render: (v) => v || '-',
    },
    {
      title: '变更前值',
      dataIndex: 'old_value',
      key: 'old_value',
      width: 200,
      ellipsis: true,
      render: (v) => v ? <span style={{ color: '#cf1322' }}>{v}</span> : '-',
    },
    {
      title: '变更后值',
      dataIndex: 'new_value',
      key: 'new_value',
      width: 200,
      ellipsis: true,
      render: (v) => v ? <span style={{ color: '#3f8600' }}>{v}</span> : '-',
    },
    {
      title: '来源IP',
      dataIndex: 'ip_address',
      key: 'ip_address',
      width: 130,
      render: (v) => v || '-',
    },
  ]

  const filteredData = searchText
    ? data.filter(item =>
        (item.username && item.username.toLowerCase().includes(searchText.toLowerCase())) ||
        (item.resource_name && item.resource_name.toLowerCase().includes(searchText.toLowerCase())) ||
        (item.action && item.action.includes(searchText))
      )
    : data

  return (
    <Card
      title="审计日志"
      size="small"
      style={{ height: '100%', display: 'flex', flexDirection: 'column' }}
      bodyStyle={{ flex: 1, display: 'flex', flexDirection: 'column', padding: 16 }}
      extra={
        <Space wrap>
          <Input
            placeholder="搜索操作人/资源名"
            prefix={<SearchOutlined />}
            allowClear
            value={searchText}
            onChange={e => setSearchText(e.target.value)}
            style={{ width: 200 }}
          />
          <Select
            placeholder="操作类型"
            allowClear
            style={{ width: 140 }}
            value={actionFilter}
            onChange={setActionFilter}
          >
            {Object.entries(actionLabels).map(([k, v]) => (
              <Option key={k} value={k}>{v.label}</Option>
            ))}
          </Select>
          <Select
            placeholder="操作人"
            allowClear
            showSearch
            optionFilterProp="children"
            style={{ width: 140 }}
            value={userFilter}
            onChange={setUserFilter}
          >
            {users.map(u => (
              <Option key={u.id} value={u.id}>{u.username}</Option>
            ))}
          </Select>
        </Space>
      }
    >
      <Table
        size="small"
        columns={columns}
        dataSource={filteredData}
        rowKey="id"
        loading={loading}
        pagination={{
          current: page,
          pageSize,
          total,
          showSizeChanger: true,
          showQuickJumper: true,
          showTotal: (t) => `共 ${t} 条`,
          onChange: (p, ps) => { setPage(p); setPageSize(ps) },
        }}
        scroll={{ x: 1200, y: 'calc(100vh - 280px)' }}
      />
    </Card>
  )
}

export default AuditLog
