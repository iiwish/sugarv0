// API响应的通用类型定义
export interface ApiResponse<T = any> {
  code: number
  data: T
  msg: string
}

// 工作空间树节点类型
export interface WorkspaceTreeNode {
  id: string
  name: string
  type: 'folder' | 'file'
  parentId?: string
  teamId: string
  children?: WorkspaceTreeNode[]
  content?: any
  createdAt?: string
  updatedAt?: string
  createdBy?: string
  updatedBy?: string
}

// 团队信息类型
export interface TeamInfo {
  id: string
  teamName: string
}

// 工作空间信息类型
export interface WorkspaceInfo {
  id: string
  name: string
  teamId: string
  description?: string
  createdAt?: string
  updatedAt?: string
}

// 文件夹操作相关类型
export interface CreateFolderData {
  name: string
  type: 'folder' | 'file'
  parentId?: string
  teamId: string
  content?: any
}

export interface RenameItemData {
  id: string
  name: string
}

export interface MoveItemData {
  id: string
  parentId?: string
}