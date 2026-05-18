<template>
  <n-alert :title="$t('common.all.error')" type="error">
    {{ $t("common.all.comError") }}
  </n-alert>
</template>
<script setup>
import { NAlert } from "naive-ui";
</script>
