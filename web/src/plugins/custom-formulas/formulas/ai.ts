import { useUserStore } from '@/pinia/modules/user'

/**
 * @description AI类公式
 * 提供AI相关的智能分析功能
 */

/**
 * AIFETCH 函数的中文本地化
 */
export const functionAiFetchZhCN = {
  formula: {
    functionList: {
      AIFETCH: {
        description: '使用AI Agent获取和分析数据，根据自然语言描述智能提取相关信息。',
        abstract: 'AI智能数据获取',
        links: [
          {
            title: '教学',
            url: 'https://univer.ai',
          },
        ],
        functionParameter: {
          agentName: {
            name: 'Agent名称',
            detail: '指定要使用的AI Agent名称',
          },
          description: {
            name: '分析需求',
            detail: '用自然语言描述您的数据分析需求',
          },
          dataRange: {
            name: '数据范围',
            detail: '可选的数据范围，如果提供则优先使用该范围的数据',
          },
        },
      },
    },
  },
}

/**
 * AIEXPLAINRANGE 函数的中文本地化
 */
export const functionAiExplainRangeZhCN = {
  formula: {
    functionList: {
      AIEXPLAINRANGE: {
        description: '使用AI分析指定数据范围，提供智能的数据洞察和解释。',
        abstract: 'AI数据范围分析',
        links: [
          {
            title: '教学',
            url: 'https://univer.ai',
          },
        ],
        functionParameter: {
          dataSource: {
            name: '数据源',
            detail: '要分析的数据范围或数组',
          },
          description: {
            name: '分析需求',
            detail: '用自然语言描述您希望AI如何分析这些数据',
          },
        },
      },
    },
  },
}

/**
 * AI公式定义
 */
export const aiFormulas = [
  {
    name: 'AIFETCH',
    implementation: (agentName: any, description: any, dataRange?: any) => {
      // 参数验证
      if (!agentName || !description) {
        return '#VALUE!'
      }

      // 转换参数为字符串
      const agentNameStr = String(agentName || '')
      const descriptionStr = String(description || '')
      const dataRangeStr = dataRange ? String(dataRange) : undefined

      try {
        // 构建请求数据
        const requestData: any = {
          agentName: agentNameStr,
          description: descriptionStr,
        }

        // 如果提供了数据范围，添加到请求中
        if (dataRangeStr) {
          requestData.dataRange = dataRangeStr
        }

        // 发送同步请求到后端
        const xhr = new XMLHttpRequest()
        xhr.open('POST', '/api/sugarFormulaQuery/executeAiFetch', false) // 同步请求
        xhr.setRequestHeader('Content-Type', 'application/json')
        
        // 添加认证头
        const userStore = useUserStore()
        if (userStore.token) {
          xhr.setRequestHeader('x-token', userStore.token)
        }
        if ((userStore.userInfo as any).ID) {
          xhr.setRequestHeader('x-user-id', (userStore.userInfo as any).ID)
        }

        xhr.send(JSON.stringify(requestData))

        if (xhr.status === 200) {
          const response = JSON.parse(xhr.responseText)
          if (response.code === 0 && response.data) {
            const result = response.data
            
            // 如果有二维数组结果，返回数组数据
            if (result.result && Array.isArray(result.result) && result.result.length > 0) {
              // 如果是单行单列，返回单个值
              if (result.result.length === 1 && result.result[0].length === 1) {
                return result.result[0][0]
              }
              // 如果是单列多行，返回纵向数组
              if (result.result[0].length === 1) {
                return result.result.map((row: any[]) => [row[0]])
              }
              // 多行多列，返回二维数组
              return result.result
            }
            
            // 如果有文本结果，返回文本
            if (result.text) {
              return result.text
            }
            
            // 如果有错误信息，返回错误
            if (result.error) {
              return result.error
            }
            
            return '#N/A'
          } else {
            return response.msg || '#ERROR!'
          }
        } else {
          return '#CONNECT!'
        }
      } catch (error) {
        console.error('AIFETCH: 执行异常:', error)
        return '#ERROR!'
      }
    },
    config: {
      description: {
        functionName: 'AIFETCH',
        description: 'formula.functionList.AIFETCH.description',
        abstract: 'formula.functionList.AIFETCH.abstract',
        functionParameter: [
          {
            name: 'formula.functionList.AIFETCH.functionParameter.agentName.name',
            detail: 'formula.functionList.AIFETCH.functionParameter.agentName.detail',
            example: '"数据分析助手"',
            require: 1,
            repeat: 0,
          },
          {
            name: 'formula.functionList.AIFETCH.functionParameter.description.name',
            detail: 'formula.functionList.AIFETCH.functionParameter.description.detail',
            example: '"帮我分析销售数据的趋势"',
            require: 1,
            repeat: 0,
          },
          {
            name: 'formula.functionList.AIFETCH.functionParameter.dataRange.name',
            detail: 'formula.functionList.AIFETCH.functionParameter.dataRange.detail',
            example: '"A1:C10"',
            require: 0,
            repeat: 0,
          },
        ],
      },
      locales: {
        zhCN: functionAiFetchZhCN,
      },
    },
    locales: functionAiFetchZhCN,
  },
  {
    name: 'AIEXPLAINRANGE',
    implementation: (dataSource: any, description: any) => {
      // 参数验证
      if (!dataSource || !description) {
        return '#VALUE!'
      }

      // 转换参数
      const descriptionStr = String(description || '')
      
      // 处理数据源 - 将其转换为二维数组
      let dataArray: any[][] = []
      
      try {
        if (Array.isArray(dataSource)) {
          // 如果已经是数组，确保是二维数组格式
          if (Array.isArray(dataSource[0])) {
            dataArray = dataSource
          } else {
            // 如果是一维数组，转换为二维数组（单列）
            dataArray = dataSource.map(item => [item])
          }
        } else {
          // 如果是单个值，转换为1x1的二维数组
          dataArray = [[dataSource]]
        }

        // 构建请求数据
        const requestData = {
          dataSource: dataArray,
          description: descriptionStr,
        }

        // 发送同步请求到后端
        const xhr = new XMLHttpRequest()
        xhr.open('POST', '/api/sugarFormulaQuery/executeAiExplainRange', false) // 同步请求
        xhr.setRequestHeader('Content-Type', 'application/json')
        
        // 添加认证头
        const userStore = useUserStore()
        if (userStore.token) {
          xhr.setRequestHeader('x-token', userStore.token)
        }
        if ((userStore.userInfo as any).ID) {
          xhr.setRequestHeader('x-user-id', (userStore.userInfo as any).ID)
        }

        xhr.send(JSON.stringify(requestData))

        if (xhr.status === 200) {
          const response = JSON.parse(xhr.responseText)
          if (response.code === 0 && response.data) {
            const result = response.data
            
            // 优先返回文本分析结果
            if (result.text) {
              return result.text
            }
            
            // 如果有二维数组结果，返回数组数据
            if (result.result && Array.isArray(result.result) && result.result.length > 0) {
              // 如果是单行单列，返回单个值
              if (result.result.length === 1 && result.result[0].length === 1) {
                return result.result[0][0]
              }
              // 如果是单列多行，返回纵向数组
              if (result.result[0].length === 1) {
                return result.result.map((row: any[]) => [row[0]])
              }
              // 多行多列，返回二维数组
              return result.result
            }
            
            // 如果有错误信息，返回错误
            if (result.error) {
              return result.error
            }
            
            return '#N/A'
          } else {
            return response.msg || '#ERROR!'
          }
        } else {
          return '#CONNECT!'
        }
      } catch (error) {
        console.error('AIEXPLAINRANGE: 执行异常:', error)
        return '#ERROR!'
      }
    },
    config: {
      description: {
        functionName: 'AIEXPLAINRANGE',
        description: 'formula.functionList.AIEXPLAINRANGE.description',
        abstract: 'formula.functionList.AIEXPLAINRANGE.abstract',
        functionParameter: [
          {
            name: 'formula.functionList.AIEXPLAINRANGE.functionParameter.dataSource.name',
            detail: 'formula.functionList.AIEXPLAINRANGE.functionParameter.dataSource.detail',
            example: 'A1:C10',
            require: 1,
            repeat: 0,
          },
          {
            name: 'formula.functionList.AIEXPLAINRANGE.functionParameter.description.name',
            detail: 'formula.functionList.AIEXPLAINRANGE.functionParameter.description.detail',
            example: '"分析这些数据的趋势和异常值"',
            require: 1,
            repeat: 0,
          },
        ],
      },
      locales: {
        zhCN: functionAiExplainRangeZhCN,
      },
    },
    locales: functionAiExplainRangeZhCN,
  },
]
