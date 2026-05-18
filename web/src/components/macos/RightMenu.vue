<script setup>
import { NIcon } from "naive-ui";
import { SearchOutlined } from "@vicons/antd";
</script>

<template>
  <!-- macOS 风格右键菜单容器 - 纯静态示例 -->
  <div
    class="fixed z-50 w-64 bg-black/85 backdrop-blur-sm user-rounded-lg shadow-lg border border-gray-700 overflow-hidden"
  >
    <!-- 菜单项列表 -->
    <div class="py-1">
      <!-- 常规菜单项 -->
      <div
        class="flex items-center px-4 py-2 text-gray-100 cursor-pointer select-none hover:user-color-bg transition-colors duration-100"
      >
        <div class="w-5 h-5 mr-3 flex-shrink-0">⌘</div>
        <div class="flex-1">新建</div>
      </div>

      <div
        class="flex items-center px-4 py-2 text-gray-100 cursor-pointer select-none hover:user-color-bg transition-colors duration-100"
      >
        <div class="w-5 h-5 mr-3 flex-shrink-0 flex items-center justify-center">
          <n-icon :component="SearchOutlined" size="16" />
        </div>
        <div class="flex-1">查找</div>
        <div class="text-sm text-gray-400">⌘F</div>
      </div>

      <!-- 分隔线 -->
      <div class="h-px my-1 bg-gray-700 mx-2"></div>

      <!-- 带快捷键的菜单项 -->
      <div
        class="flex items-center px-4 py-2 text-gray-100 cursor-pointer select-none hover:user-color-bg transition-colors duration-100"
      >
        <div class="flex-1">复制</div>
        <div class="text-sm text-gray-400">⌘C</div>
      </div>

      <div
        class="flex items-center px-4 py-2 text-gray-100 cursor-pointer select-none hover:user-color-bg transition-colors duration-100"
      >
        <div class="flex-1">粘贴</div>
        <div class="text-sm text-gray-400">⌘V</div>
      </div>

      <div
        class="flex items-center px-4 py-2 text-gray-100 cursor-pointer select-none hover:user-color-bg transition-colors duration-100"
      >
        <div class="flex-1">删除</div>
        <div class="text-sm text-gray-400">⌫</div>
      </div>

      <!-- 分隔线 -->
      <div class="h-px my-1 bg-gray-700 mx-2"></div>

      <!-- 禁用状态菜单项 -->
      <div
        class="flex items-center px-4 py-2 text-gray-100 opacity-50 cursor-not-allowed select-none"
      >
        <div class="flex-1">分享</div>
      </div>

      <!-- 更多菜单项 -->
      <div
        class="flex items-center px-4 py-2 text-gray-100 cursor-pointer select-none hover:user-color-bg transition-colors duration-100"
      >
        <div class="flex-1">属性</div>
        <div class="text-sm text-gray-400">⌘I</div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* 确保菜单项不可选择 */
.select-none {
  user-select: none;
}
</style>
