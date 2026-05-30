<script setup>
import Login from "../common/component/user/login/loginbox.vue"
import { computed, reactive, ref } from "vue"
const wallPaper = computed(
  () => {
    const data = reactive({
      DefaultWallpaper: $user.settings.DefaultWallpaper,
      DefaultWallpaperType: $user.settings.DefaultWallpaperType
    })
    if ($user.settings.DefaultWallpaperType == 'image' && $user.settings.DefaultWallpaper == "default") {
      userDefaultWallpaper(data)
    }
    return data;
  }
)
const userDefaultWallpaper = (data) => {
  const canvas = document.createElement('canvas')
  canvas.width = 1920
  canvas.height = 1080
  const ctx = canvas.getContext('2d')

  const color = getComputedStyle(document.documentElement)
    .getPropertyValue('--user-primary-color')
    .trim()

  const hexToRgb = hex => ({
    r: parseInt(hex.slice(1, 3), 16),
    g: parseInt(hex.slice(3, 5), 16),
    b: parseInt(hex.slice(5, 7), 16)
  })

  const rgb = hexToRgb(color)

  // ===== 1. 浅色底 =====
  const lighten = 0.2
  ctx.fillStyle = `rgb(
    ${Math.round(rgb.r + (255 - rgb.r) * lighten)},
    ${Math.round(rgb.g + (255 - rgb.g) * lighten)},
    ${Math.round(rgb.b + (255 - rgb.b) * lighten)}
  )`
  ctx.fillRect(0, 0, canvas.width, canvas.height)

  // ===== 2. 随机直线算法 =====
  const lineCount = 600

  for (let i = 0; i < lineCount; i++) {
    // 随机起点
    const x1 = Math.random() * canvas.width
    const y1 = Math.random() * canvas.height

    // 随机方向
    const angle = Math.random() * Math.PI * 2
    const length = 40 + Math.random() * 180

    const x2 = x1 + Math.cos(angle) * length
    const y2 = y1 + Math.sin(angle) * length

    // 随机深浅
    const isLight = Math.random() > 0.5
    const factor = isLight ? 1.7 : 0.4
    const alpha = 0.06 + Math.random() * 0.12

    ctx.beginPath()
    ctx.moveTo(x1, y1)
    ctx.lineTo(x2, y2)

    ctx.strokeStyle = `rgba(
      ${Math.round(rgb.r * factor)},
      ${Math.round(rgb.g * factor)},
      ${Math.round(rgb.b * factor)},
      ${alpha}
    )`

    // 随机线宽
    ctx.lineWidth = 0.6 + Math.random() * 1.2
    ctx.stroke()
  }

  // ===== 3. 全局柔光（压住杂乱感）=====
  const glow = ctx.createRadialGradient(
    canvas.width / 2,
    canvas.height / 2,
    0,
    canvas.width / 2,
    canvas.height / 2,
    Math.hypot(canvas.width, canvas.height)
  )
  glow.addColorStop(0, `rgba(${rgb.r}, ${rgb.g}, ${rgb.b}, 0.04)`)
  glow.addColorStop(1, 'transparent')

  ctx.fillStyle = glow
  ctx.fillRect(0, 0, canvas.width, canvas.height)

  // ===== 输出 =====
  canvas.toBlob(blob => {
    if (blob) data.DefaultWallpaper = URL.createObjectURL(blob)
  })
}
</script>
<template>
  <div class="w-full h-full relative overflow-hidden">
    <!-- 全屏背景 -->
    <img v-if="
      wallPaper.DefaultWallpaperType == 'image' &&
      wallPaper.DefaultWallpaper
    " class="absolute inset-0 w-full h-full object-cover" :src="wallPaper.DefaultWallpaper" />
    <video v-if="
      wallPaper.DefaultWallpaperType == 'video' &&
      wallPaper.DefaultWallpaper
    " class="absolute inset-0 w-full h-full object-cover" :src="wallPaper.DefaultWallpaper" autoplay muted
      loop></video>

    <!-- 登录卡片 - 居中 -->
    <div class="relative z-10 w-full h-full flex items-center justify-center">
      <div class="w-[500px] user-rounded-xl shadow-2xl border border-gray-100 p-10">
        <!-- Logo和标题 -->
        <div class="text-center mb-6">
          <img class="w-16 h-16 mx-auto mb-4" :src="$user.webSettings.Logo" alt="Logo" />
          <h1 class="text-2xl font-bold user-color-ftext">
            {{ $user.webSettings.Title }}
          </h1>
        </div>

        <!-- 账号登录标题 -->
        <div class="text-center mb-3">
          <h2 class="text-2xl font-bold user-color-ftext">
            欢迎回来,请登录您的账号
          </h2>
        </div>
        <!-- 登录组件占位 -->
        <Login />
      </div>
    </div>
    <div
      class="text-sm fixed z-99 bottom-16 left-0 right-0 mx-auto flex flex-row space-x-4 w-max"
    >
      <span class="user-color-ftext">Copyright © 2026-{{new Date().getFullYear()}}</span>
      <a class="text-blue/900" href="https://beian.miit.gov.cn/" target="_blank">{{$user.webSettings.BeianName}}</a>
      <a class="user-color-ftext" href="https://www.v9os.com" target="_blank">Power by V9os {{$user.webSettings.Version}}</a>
    </div>
  </div>
</template>
<style scoped></style>
