import React, { useState, useEffect } from 'react'
import { Table, Card, Select, Space, Tag, Button, Modal, Input, message, Tooltip } from 'antd'
import { CheckCircleOutlined, CloseCircleOutlined } from '@ant-design/icons'
import dayjs from 'dayjs'
import { approvalApi } from '../api'

const { Option } = Select
const { TextArea } = Input

const statusLabels = {
  pending: { label: '待审批', color: 'orange' },
  approved: { label: '已通过', color: 'green' },
  rejected: { label: '已拒绝', color: 'red' },
}

function ApprovalManagement() {
  const [data, setData] = useState([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [loading, setLoading] = useState(false)
  const [statusFilter, setStatusFilter] = useState('pending')
  const [rejectModalVisible, setRejectModalVisible] = useState(false)
  const [rejectTarget, setRejectTarget] = useState(null)
  const [reviewNote, setReviewNote] = useState('')

  useEffect(() => {
    loadData()
  }, [page, pageSize, statusFilter])

  const loadData = async () => {
    setLoading(true)
    try {
      const params = { page, page_size: pageSize }
      if (statusFilter) params.status = statusFilter
      const result = await approvalApi.list(params)
      setData(result.items || [])
      setTotal(result.total || 0)
    } catch (e) {
      message.error('加载审批列表失败')
    } finally {
      setLoading(false)
    }
  }

  const handleApprove = async (id) => {
    try {
      await approvalApi.approve(id)
      message.success('已通过审批，配置已更新')
      loadData()
    } catch (e) {
      message.error(e.response?.data?.error || '审批失败')
    }
  }

  const handleRejectClick = (record) => {
    setRejectTarget(record)
    setReviewNote('')
    setRejectModalVisible(true)
  }

  const handleRejectConfirm = async () => {
    if (!rejectTarget) return
    try {
      await approvalApi.reject(rejectTarget.id, reviewNote)
      message.success('已拒绝该申请')
      setRejectModalVisible(false)
      setRejectTarget(null)
      setReviewNote('')
      loadData()
    } catch (e) {
      message.error(e.response?.data?.error || '操作失败')
    }
  }

  const truncateValue = (v) => {
    if (!v) return '-'
    return v.length > 80 ? v.substring(0, 80) + '...' : v
  }

  const columns = [
    {
      title: '申请时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 170,
      render: (v) => dayjs(v).format('YYYY-MM-DD HH:mm:ss'),
    },
    {
      title: '申请人',
      dataIndex: 'applicant',
      key: 'applicant',
      width: 100,
    },
    {
      title: '配置Key',
      dataIndex: 'config_key',
      key: 'config_key',
      width: 180,
      ellipsis: true,
    },
    {
      title: '变更内容摘要',
      key: 'change_summary',
      width: 280,
      render: (_, record) => (
        <div style={{ fontSize: 12 }}>
          <div><span style={{ color: '#999' }}>旧值: </span><span style={{ color: '#cf1322' }}>{truncateValue(record.old_value)}</span></div>
          <div><span style={{ color: '#999' }}>新值: </span><span style={{ color: '#3f8600' }}>{truncateValue(record.new_value)}</span></div>
        </div>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 90,
      render: (v) => {
        const info = statusLabels[v] || { label: v, color: 'default' }
        return <Tag color={info.color}>{info.label}</Tag>
      },
    },
    {
      title: '审批人',
      dataIndex: 'reviewer',
      key: 'reviewer',
      width: 100,
      render: (v) => v || '-',
    },
    {
      title: '操作',
      key: 'action',
      width: 160,
      fixed: 'right',
      render: (_, record) => {
        if (record.status !== 'pending') {
          return <span style={{ color: '#999' }}>已处理</span>
        }
        return (
          <Space>
            <Tooltip title="通过审批并更新配置">
              <Button
                type="primary"
                size="small"
                icon={<CheckCircleOutlined />}
                onClick={() => handleApprove(record.id)}
              >
                通过
              </Button>
            </Tooltip>
            <Button
              danger
              size="small"
              icon={<CloseCircleOutlined />}
              onClick={() => handleRejectClick(record)}
            >
              拒绝
            </Button>
          </Space>
        )
      },
    },
  ]

  return (
    <Card
      title="审批管理"
      size="small"
      style={{ height: '100%', display: 'flex', flexDirection: 'column' }}
      bodyStyle={{ flex: 1, display: 'flex', flexDirection: 'column', padding: 16 }}
      extra={
        <Space>
          <Select
            value={statusFilter}
            onChange={setStatusFilter}
            style={{ width: 140 }}
          >
            <Option value="pending">待审批</Option>
            <Option value="approved">已通过</Option>
            <Option value="rejected">已拒绝</Option>
            <Option value="">全部</Option>
          </Select>
        </Space>
      }
    >
      <Table
        size="small"
        columns={columns}
        dataSource={data}
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
        scroll={{ x: 1100, y: 'calc(100vh - 280px)' }}
      />

      <Modal
        title="拒绝审批"
        open={rejectModalVisible}
        onOk={handleRejectConfirm}
        onCancel={() => { setRejectModalVisible(false); setRejectTarget(null) }}
        okText="确认拒绝"
        cancelText="取消"
      >
        <p>确认拒绝来自 <strong>{rejectTarget?.applicant}</strong> 对配置 <strong>{rejectTarget?.config_key}</strong> 的变更申请？</p>
        <TextArea
          rows={3}
          placeholder="拒绝原因（可选）"
          value={reviewNote}
          onChange={(e) => setReviewNote(e.target.value)}
        />
      </Modal>
    </Card>
  )
}

export default ApprovalManagement
