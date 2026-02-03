'use client'

import { featureFlags, FeatureFlag, DEV_SHOW_ALL_FEATURES } from '@/config/features'

/**
 * Hook 用于检查功能是否启用
 *
 * @example
 * const { isEnabled } = useFeatureFlag('instantResearch')
 * if (!isEnabled) return <DisabledFeatureBanner />
 */
export function useFeatureFlag(feature: FeatureFlag) {
  const isEnabled = DEV_SHOW_ALL_FEATURES || featureFlags[feature]

  return {
    isEnabled,
    isDisabled: !isEnabled,
  }
}

/**
 * Hook 用于获取多个功能状态
 */
export function useFeatureFlags<T extends FeatureFlag[]>(features: T) {
  return features.reduce((acc, feature) => {
    (acc as Record<FeatureFlag, boolean>)[feature] = DEV_SHOW_ALL_FEATURES || featureFlags[feature]
    return acc
  }, {} as Record<T[number], boolean>)
}
