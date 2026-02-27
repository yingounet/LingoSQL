<template>
  <div class="page-header">
    <div class="header-content">
      <div class="title-section">
        <h1 class="page-title">{{ title }}</h1>
        <p class="page-description" v-if="description">{{ description }}</p>
      </div>
      <div class="actions-section" v-if="$slots.actions">
        <slot name="actions"></slot>
      </div>
    </div>
    <!-- 可选的额外内容插槽 -->
    <div class="header-extra" v-if="$slots.extra">
      <slot name="extra"></slot>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  title: string
  description?: string
}>()
</script>

<style scoped>
.page-header {
  margin-bottom: var(--spacing-lg);
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: var(--spacing-md);
}

.title-section {
  flex: 1;
}

.page-title {
  margin: 0;
  font-size: var(--font-size-h1);
  font-weight: 700;
  color: var(--color-text-primary);
  line-height: 1.3;
}

.page-description {
  margin: var(--spacing-xs) 0 0;
  font-size: var(--font-size-body);
  color: var(--color-text-secondary);
  line-height: 1.5;
}

.actions-section {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  flex-shrink: 0;
}

.header-extra {
  margin-top: var(--spacing-md);
}

/* 响应式 */
@media (max-width: 768px) {
  .header-content {
    flex-direction: column;
    align-items: stretch;
  }
  
  .actions-section {
    margin-top: var(--spacing-md);
  }
  
  .page-title {
    font-size: 20px;
  }
}
</style>
