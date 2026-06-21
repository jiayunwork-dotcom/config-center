import React, { useState, useEffect } from 'react'
import { Card, Button, Space, List, Tag, Progress, Modal, Form, Select, InputNumber, Input, message, Statistic, Row, Col } from 'antd'
import { PlayCircleOutlined, CheckCircleOutlined, RollbackOutlined } from '@ant-design/icons'
import dayjs from 'dayjs'
import { grayApi } from '../api'

function GrayPanel({ configId, currentVersion, canEdit = true }) {
  const [grayList, setGrayList] = useState([])
  const [createModalVisible, setCreateModalVisible] = useState(false)
  const [form] = Form.useForm()

  useEffect(() => {
    if (configId) {
      loadGrayList()
    }
  }, [configId])

  const loadGrayList = async () => {
    try {
      const result = await grayApi.list(configId)
      setGrayList(result || [])
    } catch (error) {
      message.error('加载灰度列表失败')
    }
  }

  const handleCreate = async (values) => {
    try {
      const data = {
        config_item_id: configId,
        target_version: values.target_version,
        strategy: values.strategy,
        tenant_id: 1
      }

      if (values.strategy === 'ip_list') {
        data.ip_list = values.ip_list ? values.ip_list.split(',').map(s => s.trim()) : []
      } else {
        data.percentage = values.percentage
      }

      await grayApi.create(data)
      message.success('灰度发布创建成功')
      setCreateModalVisible(false)
      form.resetFields()
      loadGrayList()
    } catch (error) {
      message.error('创建失败')
    }
  }

  const handleStart = async (id) => {
    try {
      await grayApi.start(id)
      message.success('灰度发布已启动')
      loadGrayList()
    } catch (error) {
      message.error('启动失败')
    }
  }

  const handleFullPush = async (id) => {
    Modal.confirm({
      title: '全量推送',
      content: '确认将此灰度配置全量推送到所有实例？',
      onOk: async () => {
        try {
          await grayApi.fullPush(id)
          message.success('全量推送成功')
          loadGrayList()
        } catch (error) {
          message.error('全量推送失败')
        }
      }
    })
  }

  const handleRollback = async (id) => {
    Modal.confirm({
      title: '灰度回滚',
      content: '确认回滚此灰度发布？',
      onOk: async () => {
        try {
          await grayApi.rollback(id)
          message.success('灰度回滚成功')
          loadGrayList()
        } catch (error) {
          message.error('回滚失败')
        }
      }
    })
  }

  const getStatusColor = (status) => {
    switch (status) {
      case 'pending': return 'default'
      case 'running': return 'processing'
      case 'completed': return 'success'
      case 'rolled_back': return 'error'
      default: return 'default'
    }
  }

  const getStatusText = (status) => {
    switch (status) {
      case 'pending': return '待启动'
      case 'running': return '进行中'
      case 'completed': return '已完成'
      case 'rolled_back': return '已回滚'
      default: return status
    }
  }

  const getProgress = (item) => {
    if (item.total_count === 0) return 0
    return Math.round((item.pushed_count / item.total_count) * 100)
  }

  return (
    <div>
      <div style={{ marginBottom: 16 }}>
        <Button
          type="primary"
          onClick={() => setCreateModalVisible(true)}
          disabled={!canEdit}
        >
          新建灰度发布
        </Button>
      </div>

      {grayList.length > 0 && (
        <Row gutter={[16, 16]}>
          {grayList.map(item => (
            <Col span={24} key={item.id}>
              <Card size="small">
                <Space direction="vertical" style={{ width: '100%' }} size="middle">
                  <Space>
                    <Tag color="blue">目标版本 v{item.target_version}</Tag>
                    <Tag color={getStatusColor(item.status)}>{getStatusText(item.status)}</Tag>
                    <Tag color="purple">{item.strategy === 'ip_list' ? 'IP列表' : '百分比'}</Tag>
                  </Space>

                  {item.status === 'running' && (
                    <Progress
                      percent={getProgress(item)}
                      status="active"
                      format={(percent) => `${item.pushed_count}/${item.total_count} 实例`}
                    />
                  )}

                  {item.started_at && (
                    <div style={{ color: '#999', fontSize: 12 }}>
                      开始时间: {dayjs(item.started_at).format('YYYY-MM-DD HH:mm:ss')}
                      {item.status === 'running' && (
                        <span style={{ marginLeft: 16 }}>
                          持续时长: {dayjs().diff(dayjs(item.started_at), 'minute')} 分钟
                        </span>
                      )}
                    </div>
                  )}

                  <Space>
                    {item.status === 'pending' && (
                      <Button
                        type="primary"
                        size="small"
                        icon={<PlayCircleOutlined />}
                        onClick={() => handleStart(item.id)}
                        disabled={!canEdit}
                      >
                        启动灰度
                      </Button>
                    )}
                    {item.status === 'running' && (
                      <>
                        <Button
                          type="primary"
                          size="small"
                          icon={<CheckCircleOutlined />}
                          onClick={() => handleFullPush(item.id)}
                          disabled={!canEdit}
                        >
                          全量推送
                        </Button>
                        <Button
                          danger
                          size="small"
                          icon={<RollbackOutlined />}
                          onClick={() => handleRollback(item.id)}
                          disabled={!canEdit}
                        >
                          灰度回滚
                        </Button>
                      </>
                    )}
                  </Space>
                </Space>
              </Card>
            </Col>
          ))}
        </Row>
      )}

      {grayList.length === 0 && (
        <div style={{ textAlign: 'center', color: '#999', padding: 40 }}>
          暂无灰度发布记录
        </div>
      )}

      <Modal
        title="新建灰度发布"
        open={createModalVisible}
        onCancel={() => setCreateModalVisible(false)}
        footer={null}
      >
        <Form form={form} layout="vertical" onFinish={handleCreate}>
          <Form.Item
            name="target_version"
            label="目标版本"
            rules={[{ required: true, message: '请输入目标版本号' }]}
          >
            <InputNumber min={1} style={{ width: '100%' }} defaultValue={currentVersion} />
          </Form.Item>

          <Form.Item
            name="strategy"
            label="灰度策略"
            rules={[{ required: true, message: '请选择灰度策略' }]}
          >
            <Select>
              <Select.Option value="ip_list">IP列表</Select.Option>
              <Select.Option value="percentage">百分比</Select.Option>
            </Select>
          </Form.Item>

          <Form.Item noStyle shouldUpdate={(prev, curr) => prev.strategy !== curr.strategy}>
            {({ getFieldValue }) => {
              const strategy = getFieldValue('strategy')
              if (strategy === 'ip_list') {
                return (
                  <Form.Item name="ip_list" label="IP列表（逗号分隔）">
                    <Input.TextArea rows={3} placeholder="192.168.1.1, 192.168.1.2" />
                  </Form.Item>
                )
              }
              if (strategy === 'percentage') {
                return (
                  <Form.Item name="percentage" label="推送百分比">
                    <InputNumber min={1} max={99} style={{ width: '100%' }} defaultValue={10} />
                  </Form.Item>
                )
              }
              return null
            }}
          </Form.Item>

          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit">创建</Button>
              <Button onClick={() => setCreateModalVisible(false)}>取消</Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default GrayPanel
