<script setup>
const cm = $contextMenu;
</script>

<template>
  <div
    v-if="cm.state.show"
    class="win10-context-menu fixed z-9999 min-w-52 py-1 text-sm user-color-ftext user-color-fbg border user-color-border shadow-xl user-rounded-2 backdrop-blur"
    :style="{ left: `${cm.state.x}px`, top: `${cm.state.y}px` }"
    @mousedown.stop
    @contextmenu.prevent.stop
  >
    <template v-for="item in cm.state.items" :key="item.key">
      <div v-if="item.type === 'separator'" class="h-px my-1 mx-2 user-color-border"></div>
      <div v-else-if="item.type === 'header'" class="px-3 py-1 text-xs opacity-65 font-semibold cursor-default">{{ item.label }}</div>
      <div
        v-else
        class="win10-context-item relative flex items-center justify-between h-8 px-3 cursor-default user-rounded-1"
        :class="{
          'opacity-45': item.disabled,
          'hover:user-color-bg': !item.disabled,
        }"
        @click.stop="cm.run(item)"
      >
        <span class="flex items-center gap-2 truncate pr-6">
          <img v-if="item.iconUrl" class="w-4 h-4 rounded object-contain shrink-0" :src="item.iconUrl" />
          <span class="truncate">{{ item.label }}</span>
        </span>
        <span v-if="item.children && item.children.length > 0" class="text-xs">&gt;</span>
        <div
          v-if="item.children && item.children.length > 0"
          class="win10-context-submenu absolute hidden min-w-52 py-1 user-color-ftext user-color-fbg border user-color-border shadow-xl user-rounded-2 max-h-[calc(100vh-16px)] overflow-y-auto"
          :style="{ top: cm.state.submenuTopMap[item.key] || '0px' }"
          :class="cm.state.submenuDirection === 'left' ? 'right-[calc(100%-2px)]' : 'left-[calc(100%-2px)]'"
        >
          <template v-for="child in item.children" :key="child.key">
            <div v-if="child.type === 'separator'" class="h-px my-1 mx-2 user-color-border"></div>
            <div v-else-if="child.type === 'header'" class="px-3 py-1 text-xs opacity-65 font-semibold cursor-default">{{ child.label }}</div>
            <div
              v-else
              class="flex items-center h-8 px-3 cursor-default user-rounded-1"
              :class="{
                'opacity-45': child.disabled,
                'hover:user-color-bg': !child.disabled,
              }"
              @click.stop="cm.run(child)"
            >
              <span class="flex items-center gap-2 truncate">
                <img v-if="child.iconUrl" class="w-4 h-4 rounded object-contain shrink-0" :src="child.iconUrl" />
                <span class="truncate">{{ child.label }}</span>
              </span>
            </div>
          </template>
        </div>
      </div>
    </template>
  </div>
</template>

<style scoped>
.win10-context-item:hover > .win10-context-submenu {
  display: block;
}
</style>
