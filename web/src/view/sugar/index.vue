<template>
  <div ref="containerRef" class="sugar-app-container">
    <!-- 未来在这里集成布局组件，例如侧边栏、标题栏等 -->
    <main class="main-content">
      <div id="univer-sheet-container" class="univer-container"></div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { useApp } from '@/composables/useApp'

// 定义组件名称，用于keep-alive
defineOptions({
  name: 'SugarApp'
})

const containerRef = ref<HTMLDivElement | null>(null)

const updateHeight = () => {
  if (containerRef.value) {
    const top = containerRef.value.offsetTop
    containerRef.value.style.height = `calc(100vh - ${top}px - 60px)`
  }
}

onMounted(() => {
  updateHeight()
  window.addEventListener('resize', updateHeight)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', updateHeight)
})

// useApp 组合式函数封装了所有初始化和清理逻辑。
// 它会在 onMounted 时自动运行，并在 onBeforeUnmount 时自动关闭。
// 现在还支持 onActivated 和 onDeactivated 来处理keep-alive的状态切换。
const app = useApp()

// 暴露应用状态供调试使用
defineExpose({
  app
})
</script>

<style scoped>
.sugar-app-container {
  width: 100%;
  /* height is now set dynamically */
  display: flex;
  flex-direction: column;
}

.main-content {
  flex: 1;
  display: flex;
  overflow: hidden;
  padding: 8px;
}

.univer-container {
  flex: 1;
  border-radius: 8px;
  background-color: white;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}
</style>