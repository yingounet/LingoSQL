/**
 * 设计规范主题配置
 * 基于统一视觉风格，定义前端设计规范
 */

export const theme = {
  // 主色调
  primary: '#0A78F2',
  primaryFallback: '#409EFF',
  background: '#FFFFFF',
  backgroundSecondary: '#F5F5F5',
  backgroundTertiary: '#FAFAFA',

  // 语义色
  success: '#67C23A',
  warning: '#E6A23C',
  danger: '#F56C6C',
  info: '#909399',

  // 文本颜色
  textPrimary: '#1F1F1F',
  textSecondary: '#666666',
  textTertiary: '#999999',
  placeholder: '#CCCCCC',

  // 边框颜色
  border: '#E0E0E0',
  borderLight: '#DCDFE6',

  // 导航激活态
  navActiveBg: '#E3F2FD',

  // 字体系统
  fontFamily: {
    ui: "Inter, 'Helvetica Neue', Roboto, Arial, sans-serif",
    code: "Monaco, 'Courier New', Consolas, monospace",
  },

  // 字号层级
  fontSize: {
    h1: '24px', // 1.5rem
    h2: '18px', // 1.125rem
    h3: '16px', // 1rem
    body: '14px', // 0.875rem
    small: '12px', // 0.75rem
    tiny: '11px', // 0.6875rem
  },

  // 字重
  fontWeight: {
    bold: 700,
    semiBold: 600,
    regular: 400,
    light: 300,
  },

  // 行高
  lineHeight: {
    title: 1.4,
    body: 1.6,
    code: 1.5,
  },

  // 间距系统（基于 8px 网格）
  spacing: {
    xs: '4px', // 0.25rem
    sm: '8px', // 0.5rem
    md: '16px', // 1rem
    lg: '24px', // 1.5rem
    xl: '32px', // 2rem
    xxl: '48px', // 3rem
  },

  // 圆角
  borderRadius: {
    small: '4px', // 0.25rem
    medium: '6px', // 0.375rem
    large: '8px', // 0.5rem
    round: '50%',
  },

  // 阴影
  shadow: {
    light: '0 1px 3px rgba(0, 0, 0, 0.1)',
    medium: '0 2px 8px rgba(0, 0, 0, 0.1)',
    strong: '0 4px 16px rgba(0, 0, 0, 0.15)',
  },

  // 组件高度
  componentHeight: {
    button: '36px',
    input: '40px',
    tableRow: '40px',
    navItem: '48px',
  },

  // 响应式断点
  breakpoint: {
    mobile: '768px',
    tablet: '1024px',
  },
}

// CSS 变量导出（用于在样式中使用）
export const cssVariables = {
  '--color-primary': theme.primary,
  '--color-primary-fallback': theme.primaryFallback,
  '--color-background': theme.background,
  '--color-background-secondary': theme.backgroundSecondary,
  '--color-background-tertiary': theme.backgroundTertiary,
  '--color-success': theme.success,
  '--color-warning': theme.warning,
  '--color-danger': theme.danger,
  '--color-info': theme.info,
  '--color-text-primary': theme.textPrimary,
  '--color-text-secondary': theme.textSecondary,
  '--color-text-tertiary': theme.textTertiary,
  '--color-placeholder': theme.placeholder,
  '--color-border': theme.border,
  '--color-border-light': theme.borderLight,
  '--color-nav-active-bg': theme.navActiveBg,
  '--font-family-ui': theme.fontFamily.ui,
  '--font-family-code': theme.fontFamily.code,
  '--font-size-h1': theme.fontSize.h1,
  '--font-size-h2': theme.fontSize.h2,
  '--font-size-h3': theme.fontSize.h3,
  '--font-size-body': theme.fontSize.body,
  '--font-size-small': theme.fontSize.small,
  '--font-size-tiny': theme.fontSize.tiny,
  '--spacing-xs': theme.spacing.xs,
  '--spacing-sm': theme.spacing.sm,
  '--spacing-md': theme.spacing.md,
  '--spacing-lg': theme.spacing.lg,
  '--spacing-xl': theme.spacing.xl,
  '--spacing-xxl': theme.spacing.xxl,
  '--border-radius-small': theme.borderRadius.small,
  '--border-radius-medium': theme.borderRadius.medium,
  '--border-radius-large': theme.borderRadius.large,
  '--shadow-light': theme.shadow.light,
  '--shadow-medium': theme.shadow.medium,
  '--shadow-strong': theme.shadow.strong,
}
