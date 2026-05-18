<script setup>
import { NForm, NFormItem, NInput, NButton,NSelect,NDatePicker,NGrid,NGridItem } from 'naive-ui'
import { windowsStore } from "@/stores/windows"
import { computed } from "vue";
import { postData } from "@/util/util";
const props = defineProps({
    data: {
    },
    winId: {
        type: String,
        default: ""
    }
})
const row = props.data
const winId = props.winId
const ws = windowsStore()
import emitter from '@/util/event.js';
const save = async () => {
    const res = await postData("user","save",row)
    if(res){
        ws.closeWindow(winId)
        emitter.emit("user-edit-over", {})
    }
}
const enabledOptions = computed(() => [
    {
        label: $t("model.user.enabled_select_1"),
        value: "1"
    },{
        label: $t("model.user.enabled_select_2"),
        value: "2"
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
  ><n-grid cols="0:1 800:2 1200:3 1600:4 2000:5 2400:6" :x-gap="12"><n-grid-item><n-form-item :label="$t('model.user.username')" path="Username">
      <n-input v-model:value="row.Username" :placeholder="$t('model.user.username')" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.user.name')" path="Name">
      <n-input v-model:value="row.Name" :placeholder="$t('model.user.name')" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.user.password')" path="Password">
      <n-input v-model:value="row.Password" :placeholder="$t('model.user.password')" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.user.email')" path="Email">
      <n-input v-model:value="row.Email" :placeholder="$t('model.user.email')" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.user.phone')" path="Phone">
      <n-input v-model:value="row.Phone" :placeholder="$t('model.user.phone')" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.user.otp')" path="Otp">
      <n-input v-model:value="row.Otp" :placeholder="$t('model.user.otp')" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.user.remark')" path="Remark">
      <n-input type="textarea" v-model:value="row.Remark" :placeholder="$t('model.user.remark')" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.user.enabled')" path="Enabled">
      <n-select v-model:value="row.Enabled" :options="enabledOptions" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.user.avatar')" path="Avatar">
      <n-input v-model:value="row.Avatar" :placeholder="$t('model.user.avatar')" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.user.qq_open_id')" path="QqOpenId">
      <n-input v-model:value="row.QqOpenId" :placeholder="$t('model.user.qq_open_id')" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.user.wx_open_id')" path="WxOpenId">
      <n-input v-model:value="row.WxOpenId" :placeholder="$t('model.user.wx_open_id')" />
    </n-form-item></n-grid-item>
    
  </n-grid>
  <div class="flex justify-end">
      <n-button type="primary" @click="save()" size="small">
        {{ $t("common.all.save") }}
      </n-button>
  </div>
  </n-form>
</template>