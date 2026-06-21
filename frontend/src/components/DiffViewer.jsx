import React from 'react'

function DiffViewer({ diff }) {
  if (!diff || diff.length === 0) {
    return <div style={{ textAlign: 'center', color: '#999', padding: 20 }}>无差异</div>
  }

  return (
    <div style={{ maxHeight: 500, overflow: 'auto', fontFamily: 'monospace', fontSize: 13 }}>
      {diff.map((line, index) => (
        <div
          key={index}
          className={`diff-line diff-${line.type}`}
          style={{
            padding: '2px 12px',
            whiteSpace: 'pre-wrap',
            wordBreak: 'break-all',
            backgroundColor: line.type === 'insert' ? '#e6ffed' : line.type === 'delete' ? '#ffeef0' : '#fff',
            color: line.type === 'insert' ? '#22863a' : line.type === 'delete' ? '#b31d28' : '#24292e'
          }}
        >
          <span style={{ display: 'inline-block', width: 20, color: '#999' }}>
            {line.type === 'insert' ? '+' : line.type === 'delete' ? '-' : ' '}
          </span>
          {line.content}
        </div>
      ))}
    </div>
  )
}

export default DiffViewer
