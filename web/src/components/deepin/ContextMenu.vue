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
      <div
        v-else
        class="group/item h-7.5 px-2.5 flex items-center justify-between relative whitespace-nowrap cursor-default text-3.25 user-rounded-1.75"
        :class="{
          'opacity-45': item.disabled,
          'hover:user-color-bg': !item.disabled,
        }"
        @click.stop="cm.run(item)"
      >
        <span class="overflow-hidden pr-6 text-ellipsis">{{ item.label }}</span>
        <span v-if="item.children && item.children.length > 0" class="text-4.5 leading-1">&rsaquo;</span>
        <div
          v-if="item.children && item.children.length > 0"
          class="hidden group-hover/item:block absolute left-[calc(100%+4px)] top--1.5 min-w-52.5 p-1.5 backdrop-blur-30 user-color-fbg user-color-ftext border user-color-border user-rounded-2.5 shadow-[0_10px_28px_color-mix(in_srgb,var(--user-text-color)_24%,transparent)]"
        >
          <template v-for="child in item.children" :key="child.key">
            <div v-if="child.type === 'separator'" class="h-px my-1.25 mx-1.5 bg-[color-mix(in_srgb,var(--user-text-color)_14%,transparent)]"></div>
            <div
              v-else
              class="h-7.5 px-2.5 flex items-center justify-between relative whitespace-nowrap cursor-default text-3.25 user-rounded-1.75"
              :class="{
                'opacity-45': child.disabled,
                'hover:user-color-bg': !child.disabled,
              }"
              @click.stop="cm.run(child)"
            >
              <span class="overflow-hidden pr-6 text-ellipsis">{{ child.label }}</span>
            </div>
          </template>
        </div>
      </div>
    </template>
  </div>
</template>
