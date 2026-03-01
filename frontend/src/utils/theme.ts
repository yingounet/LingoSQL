/**
 * 设计规范主题配置
 * 基于统一视觉风格，定义前端设计规范
 */

export const theme = {
  // 主色调
  primary: '#0A78F2',
  primaryFallback: '#409EFF',
  primaryHover: '#0969d9',
  primaryActive: '#085bbf',
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
  primarySoft: '#E3F2FD',
  warningSoft: '#FFF3E0',
  successSoft: '#E8F5E9',

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

export const darkTheme = {
  primary: '#4C8DFF',
  primaryFallback: '#4C8DFF',
  primaryHover: '#5A9BFF',
  primaryActive: '#3B7BE6',
  background: '#0F1115',
  backgroundSecondary: '#151A21',
  backgroundTertiary: '#1B212B',
  success: '#4ADE80',
  warning: '#FBBF24',
  danger: '#F87171',
  info: '#94A3B8',
  textPrimary: '#E5E7EB',
  textSecondary: '#CBD5E1',
  textTertiary: '#94A3B8',
  placeholder: '#64748B',
  border: '#2B3440',
  borderLight: '#334155',
  navActiveBg: '#1F2937',
  primarySoft: 'rgba(76, 141, 255, 0.16)',
  warningSoft: 'rgba(251, 191, 36, 0.16)',
  successSoft: 'rgba(74, 222, 128, 0.16)',
  fontFamily: theme.fontFamily,
  fontSize: theme.fontSize,
  fontWeight: theme.fontWeight,
  lineHeight: theme.lineHeight,
  spacing: theme.spacing,
  borderRadius: theme.borderRadius,
  shadow: {
    light: '0 1px 3px rgba(0, 0, 0, 0.45)',
    medium: '0 2px 8px rgba(0, 0, 0, 0.5)',
    strong: '0 4px 16px rgba(0, 0, 0, 0.6)',
  },
  componentHeight: theme.componentHeight,
  breakpoint: theme.breakpoint,
}

// CSS 变量导出（用于在样式中使用）
export const cssVariables = {
  '--color-primary': theme.primary,
  '--color-primary-fallback': theme.primaryFallback,
  '--color-primary-hover': theme.primaryHover,
  '--color-primary-active': theme.primaryActive,
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
  '--color-primary-soft': theme.primarySoft,
  '--color-warning-soft': theme.warningSoft,
  '--color-success-soft': theme.successSoft,
  '--color-error': theme.danger,
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

export const darkCssVariables = {
  '--color-primary': darkTheme.primary,
  '--color-primary-fallback': darkTheme.primaryFallback,
  '--color-primary-hover': darkTheme.primaryHover,
  '--color-primary-active': darkTheme.primaryActive,
  '--color-background': darkTheme.background,
  '--color-background-secondary': darkTheme.backgroundSecondary,
  '--color-background-tertiary': darkTheme.backgroundTertiary,
  '--color-success': darkTheme.success,
  '--color-warning': darkTheme.warning,
  '--color-danger': darkTheme.danger,
  '--color-info': darkTheme.info,
  '--color-text-primary': darkTheme.textPrimary,
  '--color-text-secondary': darkTheme.textSecondary,
  '--color-text-tertiary': darkTheme.textTertiary,
  '--color-placeholder': darkTheme.placeholder,
  '--color-border': darkTheme.border,
  '--color-border-light': darkTheme.borderLight,
  '--color-nav-active-bg': darkTheme.navActiveBg,
  '--color-primary-soft': darkTheme.primarySoft,
  '--color-warning-soft': darkTheme.warningSoft,
  '--color-success-soft': darkTheme.successSoft,
  '--color-error': darkTheme.danger,
  '--font-family-ui': darkTheme.fontFamily.ui,
  '--font-family-code': darkTheme.fontFamily.code,
  '--font-size-h1': darkTheme.fontSize.h1,
  '--font-size-h2': darkTheme.fontSize.h2,
  '--font-size-h3': darkTheme.fontSize.h3,
  '--font-size-body': darkTheme.fontSize.body,
  '--font-size-small': darkTheme.fontSize.small,
  '--font-size-tiny': darkTheme.fontSize.tiny,
  '--spacing-xs': darkTheme.spacing.xs,
  '--spacing-sm': darkTheme.spacing.sm,
  '--spacing-md': darkTheme.spacing.md,
  '--spacing-lg': darkTheme.spacing.lg,
  '--spacing-xl': darkTheme.spacing.xl,
  '--spacing-xxl': darkTheme.spacing.xxl,
  '--border-radius-small': darkTheme.borderRadius.small,
  '--border-radius-medium': darkTheme.borderRadius.medium,
  '--border-radius-large': darkTheme.borderRadius.large,
  '--shadow-light': darkTheme.shadow.light,
  '--shadow-medium': darkTheme.shadow.medium,
  '--shadow-strong': darkTheme.shadow.strong,
}
