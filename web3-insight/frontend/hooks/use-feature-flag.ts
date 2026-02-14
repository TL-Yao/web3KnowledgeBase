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
