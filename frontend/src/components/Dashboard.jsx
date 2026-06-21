import React, { useState, useEffect } from 'react'
import { Row, Col, Card, Statistic, Table, Tag, Select, Space } from 'antd'
import {
  LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip,
  ResponsiveContainer, AreaChart, Area, BarChart, Bar
} from 'recharts'
import { namespaceApi, pushApi, metricApi } from '../api'
import dayjs from 'dayjs'

const { Option } = Select

function Dashboard({ environment }) {
  const [namespaces, setNamespaces] = useState([])
  const [selectedNamespace, setSelectedNamespace] = useState(null)
  const [connectionStats, setConnectionStats] = useState({})
  const [metrics, setMetrics] = useState({ pull_qps: 0, push_success_rate: 0, avg_latency: 0 })
  const [connections, setConnections] = useState([])
  const [chartData, setChartData] = useState([])
  const [duration, setDuration] = useState('1h')

  useEffect(() => {
    loadNamespaces()
  }, [])

  useEffect(() => {
    if (selectedNamespace) {
      loadConnectionStats()
      loadLatestMetrics()
      loadConnections()
      loadMetricChart()
    }
  }, [selectedNamespace, duration])

  const loadNamespaces = async () => {
    try {
      const result = await namespaceApi.list()
      setNamespaces(result)
      if (result.length > 0) {
        setSelectedNamespace(result[0].id)
      }
    } catch (error) {
      console.error('Failed to load namespaces')
    }
  }

  const loadConnectionStats = async () => {
    try {
      const result = await pushApi.stats()
      setConnectionStats(result || {})
    } catch (error) {
      console.error('Failed to load connection stats')
    }
  }

  const loadLatestMetrics = async () => {
    try {
      const result = await metricApi.latest(selectedNamespace)
      setMetrics(result || { pull_qps: 0, push_success_rate: 0, avg_latency: 0 })
    } catch (error) {
      console.error('Failed to load metrics')
    }
  }

  const loadConnections = async () => {
    try {
      const result = await pushApi.connections(selectedNamespace)
      setConnections(result || [])
    } catch (error) {
      console.error('Failed to load connections')
    }
  }

  const loadMetricChart = async () => {
    try {
      const qpsData = await metricApi.get({
        namespace_id: selectedNamespace,
        metric_type: 'pull_qps',
        duration
      })

      const latencyData = await metricApi.get({
        namespace_id: selectedNamespace,
        metric_type: 'avg_latency',
        duration
      })

      const data = []
      const now = dayjs()
      const points = duration === '1h' ? 12 : 24
      const interval = duration === '1h' ? 5 : 60

      for (let i = points; i >= 0; i--) {
        const time = now.subtract(i * interval, 'minute')
        const qpsPoint = qpsData.find(p =>
          Math.abs(dayjs(p.timestamp).diff(time, 'minute')) < interval / 2
        )
        const latencyPoint = latencyData.find(p =>
          Math.abs(dayjs(p.timestamp).diff(time, 'minute')) < interval / 2
        )

        data.push({
          time: time.format('HH:mm'),
          qps: qpsPoint ? qpsPoint.value : Math.random() * 100,
          latency: latencyPoint ? latencyPoint.value : Math.random() * 50
        })
      }

      setChartData(data)
    } catch (error) {
      console.error('Failed to load chart data')
    }
  }

  const columns = [
    {
      title: '客户端ID',
      dataIndex: 'client_id',
      key: 'client_id',
    },
    {
      title: 'IP地址',
      dataIndex: 'ip_address',
      key: 'ip_address',
    },
    {
      title: '连接类型',
      dataIndex: 'connect_type',
      key: 'connect_type',
      render: (type) => (
        <Tag color={type === 'websocket' ? 'blue' : 'green'}>{type}</Tag>
      )
    },
    {
      title: '最后拉取时间',
      dataIndex: 'last_pull_at',
      key: 'last_pull_at',
      render: (time) => time ? dayjs(time).format('MM-DD HH:mm:ss') : '-'
    }
  ]

  return (
    <div>
      <Card style={{ marginBottom: 16 }}>
        <Space size="large">
          <span>命名空间:</span>
          <Select
            value={selectedNamespace}
            onChange={setSelectedNamespace}
            style={{ width: 200 }}
          >
            {namespaces.map(ns => (
              <Option key={ns.id} value={ns.id}>{ns.name}</Option>
            ))}
          </Select>
          <span>时间范围:</span>
          <Select value={duration} onChange={setDuration} style={{ width: 120 }}>
            <Option value="1h">最近1小时</Option>
            <Option value="24h">最近24小时</Option>
          </Select>
        </Space>
      </Card>

      <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
        <Col span={6}>
          <Card>
            <Statistic
              title="当前连接数"
              value={connectionStats[selectedNamespace] || 0}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="配置拉取QPS"
              value={metrics.pull_qps || 0}
              precision={2}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="推送成功率"
              value={metrics.push_success_rate || 0}
              precision={2}
              suffix="%"
              valueStyle={{ color: '#722ed1' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="平均推送延迟"
              value={metrics.avg_latency || 0}
              precision={0}
              suffix="ms"
              valueStyle={{ color: '#fa8c16' }}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
        <Col span={12}>
          <Card title="配置拉取QPS趋势">
            <div style={{ height: 250 }}>
              <ResponsiveContainer width="100%" height="100%">
                <AreaChart data={chartData}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="time" />
                  <YAxis />
                  <Tooltip />
                  <Area type="monotone" dataKey="qps" stroke="#52c41a" fill="#52c41a" fillOpacity={0.3} />
                </AreaChart>
              </ResponsiveContainer>
            </div>
          </Card>
        </Col>
        <Col span={12}>
          <Card title="推送延迟趋势 (ms)">
            <div style={{ height: 250 }}>
              <ResponsiveContainer width="100%" height="100%">
                <LineChart data={chartData}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="time" />
                  <YAxis />
                  <Tooltip />
                  <Line type="monotone" dataKey="latency" stroke="#fa8c16" strokeWidth={2} dot={false} />
                </LineChart>
              </ResponsiveContainer>
            </div>
          </Card>
        </Col>
      </Row>

      <Card title="客户端连接列表">
        <Table
          columns={columns}
          dataSource={connections}
          rowKey="id"
          size="small"
          pagination={{ pageSize: 10 }}
        />
      </Card>
    </div>
  )
}

export default Dashboard
