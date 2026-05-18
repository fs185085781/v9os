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
      <div
        v-else
        class="win10-context-item relative flex items-center justify-between h-8 px-3 cursor-default user-rounded-1"
        :class="{
          'opacity-45': item.disabled,
          'hover:user-color-bg': !item.disabled,
        }"
        @click.stop="cm.run(item)"
      >
        <span class="truncate pr-6">{{ item.label }}</span>
        <span v-if="item.children && item.children.length > 0" class="text-xs">&gt;</span>
        <div
          v-if="item.children && item.children.length > 0"
          class="win10-context-submenu absolute hidden left-full top-0 min-w-52 py-1 user-color-ftext user-color-fbg border user-color-border shadow-xl user-rounded-2"
        >
          <template v-for="child in item.children" :key="child.key">
            <div v-if="child.type === 'separator'" class="h-px my-1 mx-2 user-color-border"></div>
            <div
              v-else
              class="flex items-center h-8 px-3 cursor-default user-rounded-1"
              :class="{
                'opacity-45': child.disabled,
                'hover:user-color-bg': !child.disabled,
              }"
              @click.stop="cm.run(child)"
            >
              <span class="truncate">{{ child.label }}</span>
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
