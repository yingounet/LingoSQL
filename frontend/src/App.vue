<template>
  <el-config-provider :locale="elementLocale">
    <router-view />
  </el-config-provider>
</template>

<script setup lang="ts">
import { elementLocale } from '@/composables/useLocale'
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

:root {
  /* 颜色系统 */
  --color-primary: #0A78F2;
  --color-primary-fallback: #409EFF;
  --color-primary-hover: #0969d9;
  --color-primary-active: #085bbf;
  --color-background: #FFFFFF;
  --color-background-secondary: #F5F5F5;
  --color-background-tertiary: #FAFAFA;
  --color-success: #67C23A;
  --color-warning: #E6A23C;
  --color-danger: #F56C6C;
  --color-info: #909399;
  --color-text-primary: #1F1F1F;
  --color-text-secondary: #666666;
  --color-text-tertiary: #999999;
  --color-placeholder: #CCCCCC;
  --color-border: #E0E0E0;
  --color-border-light: #DCDFE6;
  --color-nav-active-bg: #E3F2FD;
  --color-primary-soft: #E3F2FD;
  --color-warning-soft: #FFF3E0;
  --color-success-soft: #E8F5E9;
  --color-error: #F56C6C;

  /* 字体系统 */
  --font-family-ui: Inter, 'Helvetica Neue', Roboto, Arial, sans-serif;
  --font-family-code: Monaco, 'Courier New', Consolas, monospace;
  --font-size-h1: 24px;
  --font-size-h2: 18px;
  --font-size-h3: 16px;
  --font-size-body: 14px;
  --font-size-small: 12px;
  --font-size-tiny: 11px;

  /* 间距系统 */
  --spacing-xs: 4px;
  --spacing-sm: 8px;
  --spacing-md: 16px;
  --spacing-lg: 24px;
  --spacing-xl: 32px;
  --spacing-xxl: 48px;

  /* 圆角 */
  --border-radius-small: 4px;
  --border-radius-medium: 6px;
  --border-radius-large: 8px;

  /* 阴影 */
  --shadow-sm: 0 1px 3px rgba(0, 0, 0, 0.1);
  --shadow-light: 0 1px 3px rgba(0, 0, 0, 0.1);
  --shadow-md: 0 2px 8px rgba(0, 0, 0, 0.1);
  --shadow-medium: 0 2px 8px rgba(0, 0, 0, 0.1);
  --shadow-lg: 0 4px 16px rgba(0, 0, 0, 0.15);
  --shadow-strong: 0 4px 16px rgba(0, 0, 0, 0.15);
  color-scheme: light;
}

[data-theme='dark'] {
  --color-primary: #4C8DFF;
  --color-primary-fallback: #4C8DFF;
  --color-primary-hover: #5A9BFF;
  --color-primary-active: #3B7BE6;
  --color-background: #0F1115;
  --color-background-secondary: #151A21;
  --color-background-tertiary: #1B212B;
  --color-success: #4ADE80;
  --color-warning: #FBBF24;
  --color-danger: #F87171;
  --color-info: #94A3B8;
  --color-text-primary: #E5E7EB;
  --color-text-secondary: #CBD5E1;
  --color-text-tertiary: #94A3B8;
  --color-placeholder: #64748B;
  --color-border: #2B3440;
  --color-border-light: #334155;
  --color-nav-active-bg: #1F2937;
  --color-primary-soft: rgba(76, 141, 255, 0.16);
  --color-warning-soft: rgba(251, 191, 36, 0.16);
  --color-success-soft: rgba(74, 222, 128, 0.16);
  --color-error: #F87171;
  --shadow-sm: 0 1px 3px rgba(0, 0, 0, 0.45);
  --shadow-light: 0 1px 3px rgba(0, 0, 0, 0.45);
  --shadow-md: 0 2px 8px rgba(0, 0, 0, 0.5);
  --shadow-medium: 0 2px 8px rgba(0, 0, 0, 0.5);
  --shadow-lg: 0 4px 16px rgba(0, 0, 0, 0.6);
  --shadow-strong: 0 4px 16px rgba(0, 0, 0, 0.6);
  color-scheme: dark;
}

body {
  font-family: var(--font-family-ui);
  font-size: var(--font-size-body);
  color: var(--color-text-primary);
  background-color: var(--color-background);
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  line-height: 1.6;
}

#app {
  width: 100%;
  height: 100vh;
}

/* 标题样式 */
h1 {
  font-size: var(--font-size-h1);
  font-weight: 700;
  line-height: 1.4;
  color: var(--color-text-primary);
}

h2 {
  font-size: var(--font-size-h2);
  font-weight: 600;
  line-height: 1.4;
  color: var(--color-text-primary);
}

h3 {
  font-size: var(--font-size-h3);
  font-weight: 600;
  line-height: 1.4;
  color: var(--color-text-primary);
}
</style>
