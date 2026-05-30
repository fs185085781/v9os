<script setup>
const cm = $contextMenu;
</script>

<template>
  <div
    v-if="cm.state.show"
    class="fixed z-9999 min-w-52.5 p-1.5 backdrop-blur-30 user-color-fbg user-color-ftext border user-color-border user-rounded-2.5 shadow-[0_10px_28px_color-mix(in_srgb,var(--user-text-color)_24%,transparent)]"
    :style="{ left: `${cm.state.x}px`, top: `${cm.state.y}px` }"
    @mousedown.stop
    @contextmenu.prevent.stop
  >
    <template v-for="item in cm.state.items" :key="item.key">
      <div v-if="item.type === 'separator'" class="h-px my-1.25 mx-1.5 bg-[color-mix(in_srgb,var(--user-text-color)_14%,transparent)]"></div>
      <div v-else-if="item.type === 'header'" class="px-2.5 py-1 text-3 opacity-65 font-600 cursor-default">{{ item.label }}</div>
      <div
        v-else
        class="group/item h-7.5 px-2.5 flex items-center justify-between relative whitespace-nowrap cursor-default text-3.25 user-rounded-1.75"
        :class="{
          'opacity-45': item.disabled,
          'hover:user-color-bg': !item.disabled,
        }"
        @click.stop="cm.run(item)"
      >
        <span class="flex items-center gap-2 overflow-hidden pr-6 text-ellipsis">
          <img v-if="item.iconUrl" class="w-4 h-4 rounded object-contain shrink-0" :src="item.iconUrl" />
          <span class="overflow-hidden text-ellipsis">{{ item.label }}</span>
        </span>
        <span v-if="item.children && item.children.length > 0" class="text-4.5 leading-1">&rsaquo;</span>
        <div
          v-if="item.children && item.children.length > 0"
          class="hidden group-hover/item:block absolute min-w-52.5 p-1.5 backdrop-blur-30 user-color-fbg user-color-ftext border user-color-border user-rounded-2.5 shadow-[0_10px_28px_color-mix(in_srgb,var(--user-text-color)_24%,transparent)] max-h-[calc(100vh-16px)] overflow-y-auto"
          :style="{ top: cm.state.submenuTopMap[item.key] || '-6px' }"
          :class="cm.state.submenuDirection === 'left' ? 'right-[calc(100%-2px)]' : 'left-[calc(100%-2px)]'"
        >
          <template v-for="child in item.children" :key="child.key">
            <div v-if="child.type === 'separator'" class="h-px my-1.25 mx-1.5 bg-[color-mix(in_srgb,var(--user-text-color)_14%,transparent)]"></div>
            <div v-else-if="child.type === 'header'" class="px-2.5 py-1 text-3 opacity-65 font-600 cursor-default">{{ child.label }}</div>
            <div
              v-else
              class="h-7.5 px-2.5 flex items-center justify-between relative whitespace-nowrap cursor-default text-3.25 user-rounded-1.75"
              :class="{
                'opacity-45': child.disabled,
                'hover:user-color-bg': !child.disabled,
              }"
              @click.stop="cm.run(child)"
            >
              <span class="flex items-center gap-2 overflow-hidden pr-6 text-ellipsis">
                <img v-if="child.iconUrl" class="w-4 h-4 rounded object-contain shrink-0" :src="child.iconUrl" />
                <span class="overflow-hidden text-ellipsis">{{ child.label }}</span>
              </span>
            </div>
          </template>
        </div>
      </div>
    </template>
  </div>
</template>
