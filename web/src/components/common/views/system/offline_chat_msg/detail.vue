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
  ><n-grid cols="0:1 800:2 1200:3 1600:4 2000:5 2400:6" :x-gap="12"><n-grid-item><n-form-item :label="$t('model.offline_chat_msg.from')" path="From">
      <n-input v-model:value="row.From" :placeholder="$t('model.offline_chat_msg.from')" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.offline_chat_msg.to')" path="To">
      <n-input v-model:value="row.To" :placeholder="$t('model.offline_chat_msg.to')" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.offline_chat_msg.msg')" path="Msg">
      <n-input v-model:value="row.Msg" :placeholder="$t('model.offline_chat_msg.msg')" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.offline_chat_msg.type')" path="Type">
      <n-input v-model:value="row.Type" :placeholder="$t('model.offline_chat_msg.type')" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.offline_chat_msg.date_time')" path="DateTime">
      <n-date-picker class="w-full" v-model:value="row.DateTime" type="datetime" clearable />
    </n-form-item></n-grid-item>
    
    <n-grid-item><n-form-item :label="$t('model.common.createdat')" path="CreatedAt">
      <n-date-picker class="w-full" v-model:value="row.CreatedAt" type="datetime" clearable />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.common.updatedat')" path="UpdatedAt">
      <n-date-picker class="w-full" v-model:value="row.UpdatedAt" type="datetime" clearable />
    </n-form-item></n-grid-item>
    </n-grid></n-form>
</template>