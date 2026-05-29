/**
 * 策略组协议类型（与后端 coredef/constants.go 保持一致）
 * 用于前端判断节点是否为策略组，避免硬编码散落在各组件中。
 */
export const STRATEGY_PROTOCOLS = ['selector', 'urltest', 'fallback', 'loadbalance'] as const;

/**
 * 判断协议类型是否为策略组
 */
export const isStrategyGroup = (protocol?: string): boolean => {
  return protocol ? STRATEGY_PROTOCOLS.includes(protocol as typeof STRATEGY_PROTOCOLS[number]) : false;
};