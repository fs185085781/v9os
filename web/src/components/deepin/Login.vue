<script setup>
import { Reload, Power } from "@vicons/ionicons5";
import { StarAndCrescent } from "@vicons/fa";
import { NIcon } from "naive-ui";
import { ref,computed } from "vue";
import Login from "../common/component/user/login/loginbox.vue"
const wallPaper = computed(
  () => {
    const data = {
      DefaultWallpaper: $user.settings.DefaultWallpaper,
      DefaultWallpaperType: $user.settings.DefaultWallpaperType,
    }
    if ($user.settings.DefaultWallpaperType == 'image' && $user.settings.DefaultWallpaper == "default") {
      data.DefaultWallpaper = "/assets/"+$user.settings.Mode+"/img/wallpaper.jpg"
    }
    return data;
  }
)
</script>
<template>
  <div class="w-full h-full login text-center">
    <img
      v-if="
        wallPaper.DefaultWallpaperType == 'image' &&
        wallPaper.DefaultWallpaper
      "
      class="w-100vw h-100vh object-cover fixed z--1 pointer-events-none"
      :src="wallPaper.DefaultWallpaper"
    />
    <video
      v-if="
        wallPaper.DefaultWallpaperType == 'video' &&
        wallPaper.DefaultWallpaper
      "
      class="w-100vw h-100vh object-cover fixed z--1 pointer-events-none"
      :src="wallPaper.DefaultWallpaper"
      autoplay
      muted
      loop
    ></video>
    <div class="inline-block w-auto relative top-1/2 -mt-40 border-2 p-5 user-rounded-5 user-color-border">
      <img
        class="rounded-full w-24 h-24 my-0 mx-auto"
        :src="$user.webSettings.Logo"
        alt="img"
      />
      <div class="font-semibold mt-2 text-xl text-white">
        {{ $user.webSettings.Title }}
      </div>
      <Login />
      
    </div>
    <div
      class="text-sm fixed bottom-16 left-0 right-0 mx-auto flex flex-row space-x-4 w-max"
    >
      <span class="user-color-ftext">Copyright © 2026-{{new Date().getFullYear()}}</span>
      <a class="text-blue/900" href="https://beian.miit.gov.cn/" target="_blank">{{$user.webSettings.BeianName}}</a>
      <a class="user-color-ftext" href="https://www.v9os.com" target="_blank">Power by V9os {{$user.webSettings.Version}}</a>
    </div>
  </div>
</template>
<style scoped></style>
