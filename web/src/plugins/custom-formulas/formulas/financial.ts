import type { BaseValueObject, IFunctionInfo } from '@univerjs/preset-sheets-core'
import {
  ArrayValueObject,
  BaseFunction,
  FunctionType,
  NumberValueObject,
  StringValueObject,
} from '@univerjs/preset-sheets-core'

/**
 * LMDI 自定义函数实现
 * 基于对数平均迪氏指数法 (LMDI) 计算任意单个因素对总指标变化的贡献值
 */
export class LmdiFunction extends BaseFunction {
  override calculate(
    baseAggregateValue: BaseValueObject,
    compAggregateValue: BaseValueObject,
    baseFactorValue: BaseValueObject,
    compFactorValue: BaseValueObject
  ): BaseValueObject {
    // 检查是否有错误值
    if (baseAggregateValue.isError()) return baseAggregateValue
    if (compAggregateValue.isError()) return compAggregateValue
    if (baseFactorValue.isError()) return baseFactorValue
    if (compFactorValue.isError()) return compFactorValue

    // 获取数值
    const baseAgg = this.getNumericValue(baseAggregateValue)
    const compAgg = this.getNumericValue(compAggregateValue)
    const baseFactor = this.getNumericValue(baseFactorValue)
    const compFactor = this.getNumericValue(compFactorValue)

    // 验证输入值必须为正数
    if (baseAgg <= 0 || compAgg <= 0 || baseFactor <= 0 || compFactor <= 0) {
      return StringValueObject.create('#NUM!')
    }

    // 处理总指标无变化的特殊情况
    if (baseAgg === compAgg) {
      return new NumberValueObject(0)
    }

    try {
      // 执行核心LMDI计算
      const l_v = (compAgg - baseAgg) / (Math.log(compAgg) - Math.log(baseAgg))
      const result = l_v * Math.log(compFactor / baseFactor)

      // 检查结果是否有效
      if (!isFinite(result)) {
        return StringValueObject.create('#NUM!')
      }

      return new NumberValueObject(result)
    } catch (error) {
      return StringValueObject.create('#VALUE!')
    }
  }

  /**
   * 从 BaseValueObject 中提取数值
   */
  private getNumericValue(valueObject: BaseValueObject): number {
    if (valueObject.isArray()) {
      // 如果是数组，取第一个非空数值
      const array = valueObject as ArrayValueObject
      const flatArray = array.getArrayValue().flat()
      for (const item of flatArray) {
        if (typeof item === 'number' && !isNaN(item)) {
          return item
        }
      }
      return 0
    }

    const value = valueObject.getValue()
    if (typeof value === 'number') {
      return value
    }

    // 尝试转换字符串为数字
    if (typeof value === 'string') {
      const num = parseFloat(value)
      return isNaN(num) ? 0 : num
    }

    return 0
  }
}

/**
 * LMDI 函数的中文本地化
 */
export const functionLmdiZhCN = {
  formula: {
    functionList: {
      LMDI: {
        description: '基于对数平均迪氏指数法 (LMDI) 计算任意单个因素对总指标变化的贡献值。',
        abstract: '计算因素对总指标变化的贡献值',
        links: [
          {
            title: '教学',
            url: 'https://univer.ai',
          },
        ],
        functionParameter: {
          baseAggregateValue: {
            name: '基础期总指标值',
            detail: '所有因素相乘后的最终期初结果 (V_base)',
          },
          compAggregateValue: {
            name: '对比期总指标值',
            detail: '所有因素相乘后的最终期末结果 (V_comp)',
          },
          baseFactorValue: {
            name: '基础期因素值',
            detail: '您希望计算贡献的那个因素的期初值 (Xi_base)',
          },
          compFactorValue: {
            name: '对比期因素值',
            detail: '您希望计算贡献的那个因素的期末值 (Xi_comp)',
          },
        },
      },
    },
  },
}

/**
 * 财务公式定义
 */
export const financialFormulas = [
  {
    name: 'SUGAR.LMDI',
    implementation: (baseAggregateValue: any, compAggregateValue: any, baseFactorValue: any, compFactorValue: any) => {
      // 验证输入值必须为正数
      if (baseAggregateValue <= 0 || compAggregateValue <= 0 || baseFactorValue <= 0 || compFactorValue <= 0) {
        return '#NUM!'
      }

      // 处理总指标无变化的特殊情况
      if (baseAggregateValue === compAggregateValue) {
        return 0
      }

      try {
        // 执行核心LMDI计算
        const l_v = (compAggregateValue - baseAggregateValue) / (Math.log(compAggregateValue) - Math.log(baseAggregateValue))
        const result = l_v * Math.log(compFactorValue / baseFactorValue)

        // 检查结果是否有效
        if (!isFinite(result)) {
          return '#NUM!'
        }

        return result
      } catch (error) {
        return '#VALUE!'
      }
    },
    config: {
      description: {
        functionName: 'SUGAR.LMDI',
        description: 'formula.functionList.LMDI.description',
        abstract: 'formula.functionList.LMDI.abstract',
        functionParameter: [
          {
            name: 'formula.functionList.LMDI.functionParameter.baseAggregateValue.name',
            detail: 'formula.functionList.LMDI.functionParameter.baseAggregateValue.detail',
            example: '100',
            require: 1,
            repeat: 0,
          },
          {
            name: 'formula.functionList.LMDI.functionParameter.compAggregateValue.name',
            detail: 'formula.functionList.LMDI.functionParameter.compAggregateValue.detail',
            example: '120',
            require: 1,
            repeat: 0,
          },
          {
            name: 'formula.functionList.LMDI.functionParameter.baseFactorValue.name',
            detail: 'formula.functionList.LMDI.functionParameter.baseFactorValue.detail',
            example: '10',
            require: 1,
            repeat: 0,
          },
          {
            name: 'formula.functionList.LMDI.functionParameter.compFactorValue.name',
            detail: 'formula.functionList.LMDI.functionParameter.compFactorValue.detail',
            example: '12',
            require: 1,
            repeat: 0,
          },
        ],
      },
      locales: {
        zhCN: functionLmdiZhCN,
      },
    },
    locales: functionLmdiZhCN,
  },
  // 可以在这里添加更多财务公式
]