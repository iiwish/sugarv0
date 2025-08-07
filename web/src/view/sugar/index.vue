<template>
  <div class="sugar-app-container">
    <!-- 未来在这里集成布局组件，例如侧边栏、标题栏等 -->
    <main class="main-content">
      <div id="univer-sheet-container" class="univer-container"></div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { useApp } from '@/composables/useApp'

// useApp 组合式函数封装了所有初始化和清理逻辑。
// 它会在 onMounted 时自动运行，并在 onBeforeUnmount 时自动关闭。
// 这使得视图组件保持了极度的简洁，只负责渲染 UI。
useApp()
</script>

<style scoped>
.sugar-app-container {
  width: 100vw;
  height: 100vh;
  display: flex;
  flex-direction: column;
}

.main-content {
  flex: 1;
  display: flex;
  overflow: hidden;
}

.univer-container {
  flex: 1;
  margin: 8px;
  border-radius: 8px;
  background-color: white;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}
</style>