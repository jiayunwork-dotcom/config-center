import React, { useState, useEffect, useRef } from 'react'
import { Card, Button, Space, Select, message, Tabs, Tag, Descriptions, Modal } from 'antd'
import { SaveOutlined, HistoryOutlined, RollbackOutlined, ThunderboltOutlined, DeleteOutlined, AuditOutlined } from '@ant-design/icons'
import Editor from '@monaco-editor/react'
import { configApi, grayApi, permissionApi } from '../api'
import VersionHistory from './VersionHistory'
import DiffViewer from './DiffViewer'
import GrayPanel from './GrayPanel'

const { Option } = Select
const { TabPane } = Tabs

function ConfigEditor({ config, environment, onConfigChange, currentUser }) {
  const [value, setValue] = useState('')
  const [format, setFormat] = useState('json')
  const [currentConfig, setCurrentConfig] = useState(null)
  const [activeTab, setActiveTab] = useState('edit')
  const [diffModalVisible, setDiffModalVisible] = useState(false)
  const [diffData, setDiffData] = useState([])
  const editorRef = useRef(null)

  const canEdit = currentConfig ? permissionApi.canEdit(currentUser, currentConfig.namespace_id) : false
  const isAdmin = permissionApi.isAdmin(currentUser)
  const isProdEditor = environment === 'prod' && canEdit && !isAdmin
  const saveButtonText = isProdEditor ? '提交审批' : '保存'
  const SaveButtonIcon = isProdEditor ? <AuditOutlined /> : <SaveOutlined />

  useEffect(() => {
    if (config) {
      setCurrentConfig(config)
      setValue(config.value || '')
      setFormat(config.format || 'json')
    }
  }, [config])

  const handleFormatChange = (newFormat) => {
    setFormat(newFormat)
  }

  const handleSave = async () => {
    if (!canEdit) {
      message.warning('您没有修改该配置的权限')
      return
    }
    if (!currentConfig) return

    try {
      const validateResult = await configApi.validate({
        value,
        format,
        schema: currentConfig.schema
      })

      if (!validateResult.valid) {
        message.error(`格式校验失败: ${validateResult.message}`)
        return
      }

      const result = await configApi.update(currentConfig.id, {
        value,
        description: '手动更新'
      })

      if (result.requires_approval) {
        message.info('已提交，等待管理员审批')
        onConfigChange && onConfigChange()
        return
      }

      setCurrentConfig(result)
      message.success('保存成功')
      onConfigChange && onConfigChange()
    } catch (error) {
      if (error.response && error.response.status === 409) {
        message.warning('该配置已有待审批的变更申请')
      } else {
        message.error('保存失败')
      }
    }
  }

  const getLanguage = () => {
    switch (format) {
      case 'json': return 'json'
      case 'yaml': return 'yaml'
      case 'toml': return 'toml'
      case 'properties': return 'plaintext'
      default: return 'plaintext'
    }
  }

  const handleEditorDidMount = (editor) => {
    editorRef.current = editor
  }

  const handleRollback = async (version) => {
    if (!canEdit) {
      message.warning('您没有修改该配置的权限')
      return
    }
    Modal.confirm({
      title: '确认回滚',
      content: `确定要回滚到版本 ${version} 吗？`,
      onOk: async () => {
        try {
          const result = await configApi.rollback(currentConfig.id, version)
          setCurrentConfig(result)
          setValue(result.value)
          setFormat(result.format)
          message.success('回滚成功')
        } catch (error) {
          message.error('回滚失败')
        }
      }
    })
  }

  if (!currentConfig) {
    return (
      <div className="empty-state config-editor-container">
        请从左侧选择一个配置项进行编辑
      </div>
    )
  }

  return (
    <div className="config-editor-container">
      <Card
        size="small"
        title={
          <Space>
            <span>{currentConfig.key}</span>
            <Tag color="blue">{format.toUpperCase()}</Tag>
            <Tag color="green">v{currentConfig.current_version}</Tag>
            {!canEdit && <Tag color="default">只读</Tag>}
            {isProdEditor && <Tag color="orange">需审批</Tag>}
          </Space>
        }
        extra={
          <Space>
            <Select
              value={format}
              onChange={handleFormatChange}
              style={{ width: 120 }}
              size="small"
              disabled={!canEdit}
            >
              <Option value="json">JSON</Option>
              <Option value="yaml">YAML</Option>
              <Option value="properties">Properties</Option>
              <Option value="toml">TOML</Option>
            </Select>
            <Button
              type="primary"
              icon={SaveButtonIcon}
              onClick={handleSave}
              size="small"
              disabled={!canEdit}
            >
              {saveButtonText}
            </Button>
            <Button
              icon={<ThunderboltOutlined />}
              size="small"
              onClick={() => setActiveTab('gray')}
              disabled={!canEdit}
            >
              灰度发布
            </Button>
          </Space>
        }
        style={{ marginBottom: 16 }}
      >
        <Descriptions size="small" column={3}>
          <Descriptions.Item label="命名空间">
            {currentConfig.namespace_id}
          </Descriptions.Item>
          <Descriptions.Item label="环境">{environment}</Descriptions.Item>
          <Descriptions.Item label="层级">{currentConfig.level}</Descriptions.Item>
        </Descriptions>
      </Card>

      <Card style={{ flex: 1, display: 'flex', flexDirection: 'column' }} bodyStyle={{ flex: 1, display: 'flex', flexDirection: 'column', padding: 0 }}>
        <Tabs activeKey={activeTab} onChange={setActiveTab} style={{ flex: 1, display: 'flex', flexDirection: 'column' }}>
          <TabPane tab="编辑" key="edit" style={{ flex: 1, padding: 16 }}>
            <div className="editor-wrapper" style={{ flex: 1 }}>
              <Editor
                height="100%"
                language={getLanguage()}
                value={value}
                onChange={(val) => setValue(val || '')}
                onMount={handleEditorDidMount}
                theme="vs-dark"
                options={{
                  minimap: { enabled: false },
                  fontSize: 14,
                  lineNumbers: 'on',
                  automaticLayout: true,
                  wordWrap: 'on',
                  readOnly: !canEdit,
                }}
              />
            </div>
          </TabPane>

          <TabPane tab="版本历史" key="history" style={{ flex: 1, overflow: 'auto', padding: 16 }}>
            <VersionHistory
              configId={currentConfig.id}
              onRollback={handleRollback}
              canEdit={canEdit}
              onCompare={(v1, v2, diff) => {
                setDiffData(diff)
                setDiffModalVisible(true)
              }}
            />
          </TabPane>

          <TabPane tab="灰度发布" key="gray" style={{ flex: 1, overflow: 'auto', padding: 16 }}>
            <GrayPanel
              configId={currentConfig.id}
              currentVersion={currentConfig.current_version}
              canEdit={canEdit}
            />
          </TabPane>
        </Tabs>
      </Card>

      <Modal
        title="版本差异对比"
        open={diffModalVisible}
        onCancel={() => setDiffModalVisible(false)}
        footer={null}
        width={800}
      >
        <DiffViewer diff={diffData} />
      </Modal>
    </div>
  )
}

export default ConfigEditor
