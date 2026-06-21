import React, { useState } from 'react'
import { Form, Input, Button, Card, message, App } from 'antd'
import { UserOutlined, LockOutlined, AppstoreOutlined } from '@ant-design/icons'
import { authApi, setToken, setUser } from '../api'
import { useNavigate } from 'react-router-dom'

function LoginPage() {
  const [loading, setLoading] = useState(false)
  const navigate = useNavigate()
  const [form] = Form.useForm()

  const onFinish = async (values) => {
    setLoading(true)
    try {
      const result = await authApi.login(values.username, values.password)
      setToken(result.token)
      setUser(result.user)

      try {
        const me = await authApi.me()
        setUser(me)
      } catch (e) {}

      message.success('登录成功')
      navigate('/')
    } catch (error) {
      const errMsg = error.response?.data?.error || '登录失败，请检查用户名和密码'
      message.error(errMsg)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div style={{
      minHeight: '100vh',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
      padding: 24
    }}>
      <Card
        style={{
          width: '100%',
          maxWidth: 400,
          boxShadow: '0 10px 40px rgba(0,0,0,0.15)',
          borderRadius: 12
        }}
        bodyStyle={{ padding: 40 }}
      >
        <div style={{ textAlign: 'center', marginBottom: 32 }}>
          <div style={{
            width: 64,
            height: 64,
            borderRadius: 16,
            background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            margin: '0 auto 16px'
          }}>
            <AppstoreOutlined style={{ fontSize: 32, color: 'white' }} />
          </div>
          <h1 style={{ fontSize: 24, fontWeight: 600, marginBottom: 4 }}>配置中心</h1>
          <p style={{ color: '#8c8c8c', margin: 0 }}>欢迎回来，请登录您的账户</p>
        </div>

        <Form
          form={form}
          name="login"
          onFinish={onFinish}
          size="large"
          layout="vertical"
        >
          <Form.Item
            name="username"
            rules={[{ required: true, message: '请输入用户名' }]}
          >
            <Input
              prefix={<UserOutlined style={{ color: '#bfbfbf' }} />}
              placeholder="用户名"
              autoComplete="username"
            />
          </Form.Item>

          <Form.Item
            name="password"
            rules={[{ required: true, message: '请输入密码' }]}
          >
            <Input.Password
              prefix={<LockOutlined style={{ color: '#bfbfbf' }} />}
              placeholder="密码"
              autoComplete="current-password"
            />
          </Form.Item>

          <Form.Item style={{ marginBottom: 0 }}>
            <Button
              type="primary"
              htmlType="submit"
              loading={loading}
              block
              style={{
                height: 44,
                background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
                border: 'none',
                fontWeight: 500
              }}
            >
              登录
            </Button>
          </Form.Item>
        </Form>

        <div style={{
          marginTop: 24,
          padding: 12,
          background: '#f5f5f5',
          borderRadius: 8,
          fontSize: 12,
          color: '#8c8c8c',
          textAlign: 'center'
        }}>
          默认账户：admin / admin123
        </div>
      </Card>
    </div>
  )
}

export default LoginPage
