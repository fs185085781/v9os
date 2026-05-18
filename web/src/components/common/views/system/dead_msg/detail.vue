<script setup>
import { NForm, NFormItem, NInput,NSelect,NDatePicker,NGrid,NGridItem } from 'naive-ui'
import { computed } from "vue";
const props = defineProps({
    data: {
    },
    winId: {
        type: String,
        default: ""
    }
})
const row = props.data
const stypeOptions = computed(() => [
    {
        label: $t("model.dead_msg.stype_select_1"),
        value: "1"
    },{
        label: $t("model.dead_msg.stype_select_2"),
        value: "2"
    },{
        label: $t("model.dead_msg.stype_select_3"),
        value: "3"
    },{
        label: $t("model.dead_msg.stype_select_4"),
        value: "4"
    },
])
</script>
<template>
  <n-form
    :model="row"
    label-placement="left"
    label-width="auto"
    require-mark-placement="right-hanging"
    size="small"
    class="w-full p-5"
    :disabled="true"
  ><n-grid cols="0:1 800:2 1200:3 1600:4 2000:5 2400:6" :x-gap="12"><n-grid-item><n-form-item :label="$t('model.dead_msg.plugin')" path="Plugin">
      <n-input v-model:value="row.Plugin" :placeholder="$t('model.dead_msg.plugin')" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.dead_msg.stype')" path="Stype">
      <n-select v-model:value="row.Stype" :options="stypeOptions" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.dead_msg.url')" path="Url">
      <n-input v-model:value="row.Url" :placeholder="$t('model.dead_msg.url')" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.dead_msg.data')" path="Data">
      <n-input v-model:value="row.Data" :placeholder="$t('model.dead_msg.data')" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.dead_msg.msg_id')" path="MsgId">
      <n-input v-model:value="row.MsgId" :placeholder="$t('model.dead_msg.msg_id')" />
    </n-form-item></n-grid-item>
    
    <n-grid-item><n-form-item :label="$t('model.common.createdat')" path="CreatedAt">
      <n-date-picker class="w-full" v-model:value="row.CreatedAt" type="datetime" clearable />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.common.updatedat')" path="UpdatedAt">
      <n-date-picker class="w-full" v-model:value="row.UpdatedAt" type="datetime" clearable />
    </n-form-item></n-grid-item>
    </n-grid></n-form>
</template>