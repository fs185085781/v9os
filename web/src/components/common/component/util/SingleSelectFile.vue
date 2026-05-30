<template>
    <div>
        <n-input-group>
            <n-input :size="props.size" v-model:value="props.val" v-if="!onlybtn"
                :style="{ width: 'calc(100% - 60px)' }" disabled />
            <n-button :size="props.size" type="primary" ghost @click="toSelectFile()" class="w-60px">
                选择
            </n-button>
        </n-input-group>
        <div class="hidden">
            <input type="file" :accept="props.accept" ref="fileInput" @change="fileChange">
        </div>
    </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { NInput, NButton, NInputGroup } from 'naive-ui'
import { postData, absoluteUrl } from '@/util/util'
import { useEventBus } from "@/util/event.js";
const props = defineProps({
    scene: {
        type: String,
        default: ''
    },
    onlybtn: {
        type: Boolean,
        default: false
    },
    val: {
        type: String,
        default: ''
    },
    accept: {
        type: String,
        default: ''
    },
    size: {
        type: String,
        default: ''
    }
})
const fileInput = ref(null)
const emit = defineEmits(['change'])
const slbOpen = ref(false)
const hasFileSystem = ref(false)
const initData = async () => {
    const res = await postData('system', 'appconfig', {}, "")
    slbOpen.value = res.distributed_enabled
    const apps = await $user.getMyApps()
    apps.forEach(app => {
        if (app.code === "file_system") {
            hasFileSystem.value = true
        }
    })
}
onMounted(() => {
    initData()
})
const toSelectFile = () => {
    if (hasFileSystem.value) {
        useEventBus("expand-change", (msg) => {
            if (msg.winId) {
                $wins.closeWindow(msg.winId)
            }
            if (msg.type != "fileSelected") {
                return
            }
            if (!msg.confirm) {
                return
            }
            emit('change', msg.data.url)
        });
        const ext = props.accept.replaceAll(".", "");
        $wins.addWindow({
            title: '文件管理',
            width: 800,
            height: 600,
            iframeUrl: absoluteUrl('/page/file_system/?page=explorer&expand=selectFile&relative=true&save=&ext=' + encodeURIComponent(ext))
        })
    } else {
        if (slbOpen.value) {
            if($user.user.IsAdmin == 1){
                $msg.message.error("请先安装文件管理系统插件")
            }else{
                $msg.message.error("请先向管理员申请文件管理系统权限")
            }
        } else {
            fileInput.value.value = ""
            fileInput.value.click()
        }
    }
}
const fileChange = async () => {
    const file = fileInput.value.files[0]
    const data = new FormData();
    data.append('file', file);
    data.append('scene', props.scene);
    const res = await postData('file', 'upload', data, "okerr")
    if (res) {
        emit('change', res)
    }
}
</script>