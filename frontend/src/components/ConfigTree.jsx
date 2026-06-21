import React, { useState, useEffect } from 'react'
import { Tree, Button, Modal, Input, message, Space } from 'antd'
import { PlusOutlined, FolderOutlined, FileOutlined, AppstoreOutlined } from '@ant-design/icons'
import { namespaceApi, groupApi, configApi } from '../api'

const { DirectoryTree } = Tree

function ConfigTree({ environment, onSelect }) {
  const [treeData, setTreeData] = useState([])
  const [expandedKeys, setExpandedKeys] = useState([])
  const [addModalVisible, setAddModalVisible] = useState(false)
  const [addType, setAddType] = useState('namespace')
  const [parentId, setParentId] = useState(null)
  const [newName, setNewName] = useState('')
  const [newDesc, setNewDesc] = useState('')

  useEffect(() => {
    loadTreeData()
  }, [environment])

  const loadTreeData = async () => {
    try {
      const namespaces = await namespaceApi.list()
      const result = []

      for (const ns of namespaces) {
        const groups = await groupApi.list(ns.id)
        const groupNodes = []

        for (const group of groups) {
          const configs = await configApi.list({
            namespace_id: ns.id,
            group_id: group.id,
            environment
          })
          const keyNodes = configs.map(cfg => ({
            key: `config-${cfg.id}`,
            title: cfg.key,
            icon: <FileOutlined />,
            isLeaf: true,
            config: cfg
          }))

          groupNodes.push({
            key: `group-${group.id}`,
            title: group.name,
            icon: <FolderOutlined />,
            children: keyNodes,
            group
          })
        }

        result.push({
          key: `ns-${ns.id}`,
          title: ns.name,
          icon: <AppstoreOutlined />,
          children: groupNodes,
          namespace: ns
        })
      }

      setTreeData(result)
    } catch (error) {
      message.error('加载配置树失败')
    }
  }

  const handleSelect = (selectedKeys, info) => {
    if (info.node.config) {
      onSelect(info.node.config)
    }
  }

  const handleAddNamespace = () => {
    setAddType('namespace')
    setParentId(null)
    setNewName('')
    setNewDesc('')
    setAddModalVisible(true)
  }

  const handleAddGroup = (nsId) => {
    setAddType('group')
    setParentId(nsId)
    setNewName('')
    setNewDesc('')
    setAddModalVisible(true)
  }

  const handleAddConfig = (groupId, nsId) => {
    setAddType('config')
    setParentId({ groupId, nsId })
    setNewName('')
    setNewDesc('')
    setAddModalVisible(true)
  }

  const handleAddConfirm = async () => {
    try {
      if (addType === 'namespace') {
        await namespaceApi.create({ name: newName, description: newDesc, tenant_id: 1 })
        message.success('命名空间创建成功')
      } else if (addType === 'group') {
        await groupApi.create({
          name: newName,
          description: newDesc,
          namespace_id: parentId,
          tenant_id: 1
        })
        message.success('分组创建成功')
      } else if (addType === 'config') {
        await configApi.create({
          key: newName,
          value: '{}',
          format: 'json',
          environment,
          namespace_id: parentId.nsId,
          group_id: parentId.groupId,
          tenant_id: 1,
          level: 'group'
        })
        message.success('配置项创建成功')
      }
      setAddModalVisible(false)
      loadTreeData()
    } catch (error) {
      message.error('创建失败')
    }
  }

  const renderTitle = (nodeData) => {
    const key = nodeData.key
    const prefix = key.split('-')[0]

    return (
      <span style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', width: '100%' }}>
        <span>{nodeData.title}</span>
        <Button
          type="text"
          size="small"
          icon={<PlusOutlined />}
          onClick={(e) => {
            e.stopPropagation()
            if (prefix === 'ns') {
              handleAddGroup(nodeData.namespace.id)
            } else if (prefix === 'group') {
              handleAddConfig(nodeData.group.id, nodeData.namespace ? nodeData.namespace.id : null)
            }
          }}
        />
      </span>
    )
  }

  const renderTreeNodes = (nodes) => {
    return nodes.map(node => {
      const title = renderTitle(node)
      if (node.children) {
        return {
          ...node,
          title,
          children: renderTreeNodes(node.children)
        }
      }
      return { ...node, title }
    })
  }

  return (
    <div className="tree-container">
      <div className="tree-title">
        <span>配置树</span>
        <Button
          type="primary"
          size="small"
          icon={<PlusOutlined />}
          onClick={handleAddNamespace}
        >
          命名空间
        </Button>
      </div>
      <DirectoryTree
        multiple={false}
        defaultExpandAll={false}
        expandedKeys={expandedKeys}
        onExpand={setExpandedKeys}
        onSelect={handleSelect}
        treeData={renderTreeNodes(treeData)}
      />

      <Modal
        title={addType === 'namespace' ? '新建命名空间' : addType === 'group' ? '新建分组' : '新建配置项'}
        open={addModalVisible}
        onOk={handleAddConfirm}
        onCancel={() => setAddModalVisible(false)}
      >
        <Space direction="vertical" style={{ width: '100%' }}>
          <Input
            placeholder={addType === 'config' ? '配置键名' : '名称'}
            value={newName}
            onChange={e => setNewName(e.target.value)}
          />
          <Input.TextArea
            placeholder="描述（可选）"
            value={newDesc}
            onChange={e => setNewDesc(e.target.value)}
            rows={3}
          />
        </Space>
      </Modal>
    </div>
  )
}

export default ConfigTree
