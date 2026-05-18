<script setup>
/**
 * ConfigSettingsPanel.vue - 配置设置面板（管理员专用）
 * 通过 API 读写系统配置，走 ioc 保证分布式一致性
 */
import { ref, computed, onMounted, onUnmounted } from "vue";
import { postData, getApiHost } from "@/util/util";
import { NSwitch, NInput, NInputNumber, NSelect, NButton } from "naive-ui";

const loading = ref(false);
const saving = ref(false);
const config = ref(null);
let restartDelayTimer = null;
let healthTimer = null;

// Go time.Duration 是纳秒，前端用秒展示
const nsToSec = (ns) => Math.round((ns || 0) / 1e9);
const secToNs = (s) => (s || 0) * 1e9;
// expire_duration 用小时展示
const nsToHour = (ns) => Math.round((ns || 0) / 3.6e12);
const hourToNs = (h) => (h || 0) * 3.6e12;

async function loadConfig() {
  loading.value = true;
  const res = await postData("system", "configGet", {}, "err");
  loading.value = false;
  if (res) {
    config.value = res;
  }
}

async function saveConfig() {
  if (!config.value) return;
  saving.value = true;
  const res = await postData("system", "configSave", config.value, "okerr");
  saving.value = false;
  if (res && res.restart) {
    $msg.message.info($t("common.settings.restarting"));
    const host = await getApiHost();
    if (restartDelayTimer) clearTimeout(restartDelayTimer);
    restartDelayTimer = setTimeout(() => {
      const start = Date.now();
      if (healthTimer) clearInterval(healthTimer);
      healthTimer = setInterval(async () => {
        try {
          const r = await fetch(`${host}/health`, { method: "GET" });
          if (r.ok) {
            clearInterval(healthTimer);
            healthTimer = null;
            location.reload();
          }
        } catch {}
        if (Date.now() - start > 60000) {
          clearInterval(healthTimer);
          healthTimer = null;
          $msg.message.error($t("common.settings.restart_timeout"));
          location.reload();
        }
      }, 2000);
    }, 2000);
  }
}

async function confirmSave() {
  if (await $msg.util.confirm($t("common.settings.save_confirm_content"))) {
    saveConfig();
  }
}

const logLevelOptions = [
  { label: "Debug", value: "debug" },
  { label: "Info", value: "info" },
  { label: "Warn", value: "warn" },
  { label: "Error", value: "error" },
];

// 日志输出方式（多选）
const logOutputOptions = [
  { label: "File", value: "file" },
  { label: "Console", value: "console" },
  { label: "Database", value: "db" },
];

// 数据库驱动类型
const databaseDriverOptions = [
  { label: "SQLite", value: "sqlite" },
  { label: "MySQL", value: "mysql" },
  { label: "GaussDB", value: "gaussdb" },
  { label: "SQL Server", value: "sqlserver" },
  { label: "ClickHouse", value: "clickhouse" },
  { label: "PostgreSQL", value: "postgres" },
];

// 缓存驱动类型
const cacheDriverOptions = [
  { label: "File", value: "file" },
  { label: "Redis", value: "redis" },
];

// 消息队列驱动类型
const queueDriverOptions = [
  { label: "Memory", value: "mem" },
  { label: "RocketMQ", value: "rocket" },
  { label: "Redis", value: "redis" },
];

// DSN 用逗号分隔的字符串编辑
const dsnText = computed({
  get: () => (config.value?.database?.dsn || []).join(","),
  set: (v) => {
    if (config.value?.database)
      config.value.database.dsn = v ? v.split(",") : [];
  },
});
const logOutputText = computed({
  get: () => (config.value?.log?.output || []).join(","),
  set: (v) => {
    if (config.value?.log) config.value.log.output = v ? v.split(",") : [];
  },
});
// 日志输出方式（多选）
const logOutputValues = computed({
  get: () => config.value?.log?.output || [],
  set: (v) => {
    if (config.value?.log) config.value.log.output = v || [];
  },
});
const cacheRedisAddrs = computed({
  get: () => (config.value?.cachebase?.redis?.addrs || []).join(","),
  set: (v) => {
    if (config.value?.cachebase?.redis)
      config.value.cachebase.redis.addrs = v ? v.split(",") : [];
  },
});
const rocketAddrs = computed({
  get: () => (config.value?.mqbase?.rocket?.addrs || []).join(","),
  set: (v) => {
    if (config.value?.mqbase?.rocket)
      config.value.mqbase.rocket.addrs = v ? v.split(",") : [];
  },
});
// Duration 用秒编辑
const serverReadTimeout = computed({
  get: () => nsToSec(config.value?.server?.read_timeout),
  set: (v) => {
    if (config.value?.server) config.value.server.read_timeout = secToNs(v);
  },
});
const serverWriteTimeout = computed({
  get: () => nsToSec(config.value?.server?.write_timeout),
  set: (v) => {
    if (config.value?.server) config.value.server.write_timeout = secToNs(v);
  },
});
const authExpireHours = computed({
  get: () => nsToHour(config.value?.auth?.expire_duration),
  set: (v) => {
    if (config.value?.auth) config.value.auth.expire_duration = hourToNs(v);
  },
});
const authRefreshExpireHours = computed({
  get: () => nsToHour(config.value?.auth?.refresh_expire_duration),
  set: (v) => {
    if (config.value?.auth)
      config.value.auth.refresh_expire_duration = hourToNs(v);
  },
});
// Redis Duration 用秒编辑
const cacheRedisReadTimeout = computed({
  get: () => nsToSec(config.value?.cachebase?.redis?.read_timeout),
  set: (v) => {
    if (config.value?.cachebase?.redis)
      config.value.cachebase.redis.read_timeout = secToNs(v);
  },
});
const cacheRedisWriteTimeout = computed({
  get: () => nsToSec(config.value?.cachebase?.redis?.write_timeout),
  set: (v) => {
    if (config.value?.cachebase?.redis)
      config.value.cachebase.redis.write_timeout = secToNs(v);
  },
});
const cacheRedisDialTimeout = computed({
  get: () => nsToSec(config.value?.cachebase?.redis?.dial_timeout),
  set: (v) => {
    if (config.value?.cachebase?.redis)
      config.value.cachebase.redis.dial_timeout = secToNs(v);
  },
});

onMounted(() => {
  loadConfig();
});
onUnmounted(() => {
  if (restartDelayTimer) clearTimeout(restartDelayTimer);
  if (healthTimer) clearInterval(healthTimer);
});
</script>

<template>
  <div class="max-w-150 mx-auto py-6 px-8" v-if="loading && !config">
    <div class="flex justify-center py-15">
      <n-spin size="medium" />
    </div>
  </div>
  <div class="max-w-150 mx-auto py-6 px-8" v-else-if="config">
    <!-- Server -->
    <div class="mb-6">
      <div
        class="text-14px font-600 user-color-ftext mb-3 pb-1.5 border-b user-color-line"
      >
        {{ $t("common.settings.server") }}
      </div>
      <div class="flex items-center justify-between py-2 min-h-9">
        <span
          class="text-13px user-color-ftext shrink-0 mr-4"
          >{{ $t("common.settings.request_log") }}</span
        >
        <n-switch v-model:value="config.server.request_log" />
      </div>
      <div class="flex items-center justify-between py-2 min-h-9">
        <span
          class="text-13px user-color-ftext shrink-0 mr-4"
          >{{ $t("common.settings.store_host") }}</span
        >
        <n-input
          v-model:value="config.server.store_host"
          class="w-65"
          size="small"
        />
      </div>
      <div class="flex items-center justify-between py-2 min-h-9">
        <span
          class="text-13px user-color-ftext shrink-0 mr-4"
          >{{ $t("common.settings.store_type") }}</span
        >
        <n-input
          v-model:value="config.server.store_type"
          class="w-45"
          size="small"
        />
      </div>
      <div class="flex items-center justify-between py-2 min-h-9">
        <span
          class="text-13px user-color-ftext shrink-0 mr-4"
          >{{ $t("common.settings.read_timeout") }}</span
        >
        <n-input-number
          v-model:value="serverReadTimeout"
          :min="1"
          :max="600"
          class="w-35"
          size="small"
        />
      </div>
      <div class="flex items-center justify-between py-2 min-h-9">
        <span
          class="text-13px user-color-ftext shrink-0 mr-4"
          >{{ $t("common.settings.write_timeout") }}</span
        >
        <n-input-number
          v-model:value="serverWriteTimeout"
          :min="1"
          :max="600"
          class="w-35"
          size="small"
        />
      </div>
      <div class="flex items-center justify-between py-2 min-h-9">
        <span
          class="text-13px user-color-ftext shrink-0 mr-4"
          >{{ $t("common.settings.system_id") }}</span
        >
        <span class="text-13px user-color-ftext">{{
          config.server?.system_id || "-"
        }}</span>
      </div>
      <div class="flex items-center justify-between py-2 min-h-9">
        <span
          class="text-13px user-color-ftext shrink-0 mr-4"
          >{{ $t("common.settings.proxy_host") }}</span
        >
        <n-input
          v-model:value="config.server.proxy_host"
          class="w-45"
          size="small"
        />
      </div>
      <div class="flex items-center justify-between py-2 min-h-9">
        <span
          class="text-13px user-color-ftext shrink-0 mr-4"
          >{{ $t("common.settings.proxy_token") }}</span
        >
        <n-input
          v-model:value="config.server.proxy_token"
          class="w-45"
          size="small"
        />
      </div>
      <div class="flex items-center justify-between py-2 min-h-9">
        <span
          class="text-13px user-color-ftext shrink-0 mr-4"
          >{{ $t("common.settings.upgrade_channel") }}</span
        >
        <n-select
          v-model:value="config.server.upgrade_channel"
          :options="[
            { label: 'Stable', value: 'stable' },
            { label: 'Beta', value: 'beta' },
          ]"
          class="w-35"
          size="small"
        />
      </div>
    </div>

    <!-- Log -->
    <div class="mb-6">
      <div
        class="text-14px font-600 user-color-ftext mb-3 pb-1.5 border-b user-color-line"
      >
        {{ $t("common.settings.log") }}
      </div>
      <div class="flex items-center justify-between py-2 min-h-9">
        <span
          class="text-13px user-color-ftext shrink-0 mr-4"
          >{{ $t("common.settings.level") }}</span
        >
        <n-select
          v-model:value="config.log.level"
          :options="logLevelOptions"
          class="w-35"
          size="small"
        />
      </div>
      <div class="flex items-center justify-between py-2 min-h-9">
        <span
          class="text-13px user-color-ftext shrink-0 mr-4"
          >{{ $t("common.settings.output") }}</span
        >
        <n-select
          v-model:value="logOutputValues"
          :options="logOutputOptions"
          multiple
          clearable
          class="w-50"
          size="small"
        />
      </div>
      <div class="flex items-center justify-between py-2 min-h-9">
        <span
          class="text-13px user-color-ftext shrink-0 mr-4"
          >{{ $t("common.settings.dir") }}</span
        >
        <n-input v-model:value="config.log.dir" class="w-50" size="small" />
      </div>
      <div class="flex items-center justify-between py-2 min-h-9">
        <span
          class="text-13px user-color-ftext shrink-0 mr-4"
          >{{ $t("common.settings.max_size") }}</span
        >
        <n-input-number
          v-model:value="config.log.max_size"
          :min="1"
          :max="1024"
          class="w-35"
          size="small"
        />
      </div>
      <div class="flex items-center justify-between py-2 min-h-9">
        <span
          class="text-13px user-color-ftext shrink-0 mr-4"
          >{{ $t("common.settings.max_backups") }}</span
        >
        <n-input-number
          v-model:value="config.log.max_backups"
          :min="1"
          :max="100"
          class="w-35"
          size="small"
        />
      </div>
      <div class="flex items-center justify-between py-2 min-h-9">
        <span
          class="text-13px user-color-ftext shrink-0 mr-4"
          >{{ $t("common.settings.max_age") }}</span
        >
        <n-input-number
          v-model:value="config.log.max_age"
          :min="1"
          :max="365"
          class="w-35"
          size="small"
        />
      </div>
    </div>

    <!-- CORS -->
    <div class="mb-6">
      <div
        class="text-14px font-600 user-color-ftext mb-3 pb-1.5 border-b user-color-line"
      >
        {{ $t("common.settings.cors") }}
      </div>
      <div class="flex items-center justify-between py-2 min-h-9">
        <span
          class="text-13px user-color-ftext shrink-0 mr-4"
          >{{ $t("common.settings.enabled") }}</span
        >
        <n-switch v-model:value="config.cors.enabled" />
      </div>
    </div>

    <!-- Rate Limit -->
    <div class="mb-6">
      <div
        class="text-14px font-600 user-color-ftext mb-3 pb-1.5 border-b user-color-line"
      >
        {{ $t("common.settings.rate_limit") }}
      </div>
      <div class="flex items-center justify-between py-2 min-h-9">
        <span
          class="text-13px user-color-ftext shrink-0 mr-4"
          >{{ $t("common.settings.enabled") }}</span
        >
        <n-switch v-model:value="config.rate_limit.enabled" />
      </div>
      <div class="flex items-center justify-between py-2 min-h-9">
        <span
          class="text-13px user-color-ftext shrink-0 mr-4"
          >{{ $t("common.settings.rps") }}</span
        >
        <n-input-number
          v-model:value="config.rate_limit.rps"
          :min="1"
          :max="100000"
          class="w-35"
          size="small"
        />
      </div>
      <div class="flex items-center justify-between py-2 min-h-9">
        <span
          class="text-13px user-color-ftext shrink-0 mr-4"
          >{{ $t("common.settings.burst") }}</span
        >
        <n-input-number
          v-model:value="config.rate_limit.burst"
          :min="1"
          :max="100000"
          class="w-35"
          size="small"
        />
      </div>
    </div>

    <!-- Distributed -->
    <div class="mb-6">
      <div
        class="text-14px font-600 user-color-ftext mb-3 pb-1.5 border-b user-color-line"
      >
        {{ $t("common.settings.distributed") }}
      </div>
      <div class="flex items-center justify-between py-2 min-h-9">
        <span
          class="text-13px user-color-ftext shrink-0 mr-4"
          >{{ $t("common.settings.enabled") }}</span
        >
        <n-switch v-model:value="config.distributed.enabled" />
      </div>
    </div>

    <!-- Auth -->
    <div class="mb-6">
      <div
        class="text-14px font-600 user-color-ftext mb-3 pb-1.5 border-b user-color-line"
      >
        {{ $t("common.settings.auth") }}
      </div>
      <div class="flex items-center justify-between py-2 min-h-9">
        <span
          class="text-13px user-color-ftext shrink-0 mr-4"
          >{{ $t("common.settings.token_expire") }}</span
        >
        <n-input-number
          v-model:value="authExpireHours"
          :min="1"
          :max="8760"
          class="w-35"
          size="small"
        />
      </div>
      <div class="flex items-center justify-between py-2 min-h-9">
        <span
          class="text-13px user-color-ftext shrink-0 mr-4"
          >{{ $t("common.settings.refresh_token_expire") }}</span
        >
        <n-input-number
          v-model:value="authRefreshExpireHours"
          :min="1"
          :max="8760"
          class="w-35"
          size="small"
        />
      </div>
    </div>

    <!-- Database -->
    <div class="mb-6">
      <div
        class="text-14px font-600 user-color-ftext mb-3 pb-1.5 border-b user-color-line"
      >
        {{ $t("common.settings.database") }}
      </div>
      <div class="flex items-center justify-between py-2 min-h-9">
        <span
          class="text-13px user-color-ftext shrink-0 mr-4"
          >{{ $t("common.settings.driver") }}</span
        >
        <n-select
          v-model:value="config.database.driver"
          :options="databaseDriverOptions"
          class="w-45"
          size="small"
        />
      </div>
      <div class="flex items-center justify-between py-2 min-h-9">
        <span
          class="text-13px user-color-ftext shrink-0 mr-4"
          >{{ $t("common.settings.dsn") }}</span
        >
        <n-input v-model:value="dsnText" class="w-65" size="small" />
      </div>
      <div class="flex items-center justify-between py-2 min-h-9">
        <span
          class="text-13px user-color-ftext shrink-0 mr-4"
          >{{ $t("common.settings.max_idle_conns") }}</span
        >
        <n-input-number
          v-model:value="config.database.max_idle_conns"
          :min="1"
          :max="1000"
          class="w-35"
          size="small"
        />
      </div>
      <div class="flex items-center justify-between py-2 min-h-9">
        <span
          class="text-13px user-color-ftext shrink-0 mr-4"
          >{{ $t("common.settings.max_open_conns") }}</span
        >
        <n-input-number
          v-model:value="config.database.max_open_conns"
          :min="1"
          :max="10000"
          class="w-35"
          size="small"
        />
      </div>
      <div class="flex items-center justify-between py-2 min-h-9">
        <span
          class="text-13px user-color-ftext shrink-0 mr-4"
          >{{ $t("common.settings.cache") }}</span
        >
        <n-switch v-model:value="config.database.cache" />
      </div>
      <div class="flex items-center justify-between py-2 min-h-9">
        <span
          class="text-13px user-color-ftext shrink-0 mr-4"
          >{{ $t("common.settings.soft_delete") }}</span
        >
        <n-switch v-model:value="config.database.soft_delete" />
      </div>
    </div>

    <!-- Cachebase -->
    <div class="mb-6">
      <div
        class="text-14px font-600 user-color-ftext mb-3 pb-1.5 border-b user-color-line"
      >
        {{ $t("common.settings.cache") }}
      </div>
      <div class="flex items-center justify-between py-2 min-h-9">
        <span
          class="text-13px user-color-ftext shrink-0 mr-4"
          >{{ $t("common.settings.driver") }}</span
        >
        <n-select
          v-model:value="config.cachebase.driver"
          :options="cacheDriverOptions"
          class="w-35"
          size="small"
        />
      </div>
      <template v-if="config.cachebase?.driver === 'file'">
        <div class="flex items-center justify-between py-2 min-h-9">
          <span
            class="text-13px user-color-ftext shrink-0 mr-4"
            >{{ $t("common.settings.file_dir") }}</span
          >
          <n-input
            v-model:value="config.cachebase.file.dir"
            class="w-50"
            size="small"
          />
        </div>
      </template>
      <template v-if="config.cachebase?.driver === 'redis'">
        <div class="flex items-center justify-between py-2 min-h-9">
          <span
            class="text-13px user-color-ftext shrink-0 mr-4"
            >{{ $t("common.settings.redis_mode") }}</span
          >
          <n-input
            v-model:value="config.cachebase.redis.mode"
            placeholder="standalone"
            class="w-40"
            size="small"
          />
        </div>
        <div class="flex items-center justify-between py-2 min-h-9" v-if="config.cachebase.redis.mode === 'sentinel'">
          <span
            class="text-13px user-color-ftext shrink-0 mr-4"
            >{{ $t("common.settings.master_name") }}</span
          >
          <n-input
            v-model:value="config.cachebase.redis.master_name"
            class="w-45"
            size="small"
          />
        </div>
        <div class="flex items-center justify-between py-2 min-h-9">
          <span
            class="text-13px user-color-ftext shrink-0 mr-4"
            >{{ $t("common.settings.redis_addrs") }}</span
          >
          <n-input v-model:value="cacheRedisAddrs" class="w-55" size="small" />
        </div>
        <div class="flex items-center justify-between py-2 min-h-9">
          <span
            class="text-13px user-color-ftext shrink-0 mr-4"
            >{{ $t("common.settings.redis_password") }}</span
          >
          <n-input
            v-model:value="config.cachebase.redis.password"
            type="password"
            show-password-on="click"
            class="w-45"
            size="small"
          />
        </div>
        <div class="flex items-center justify-between py-2 min-h-9">
          <span
            class="text-13px user-color-ftext shrink-0 mr-4"
            >{{ $t("common.settings.redis_db") }}</span
          >
          <n-input-number
            v-model:value="config.cachebase.redis.db"
            :min="0"
            :max="15"
            class="w-35"
            size="small"
          />
        </div>
        <div class="flex items-center justify-between py-2 min-h-9">
          <span
            class="text-13px user-color-ftext shrink-0 mr-4"
            >{{ $t("common.settings.pool_size") }}</span
          >
          <n-input-number
            v-model:value="config.cachebase.redis.pool_size"
            :min="1"
            :max="1000"
            class="w-35"
            size="small"
          />
        </div>
        <div class="flex items-center justify-between py-2 min-h-9">
          <span
            class="text-13px user-color-ftext shrink-0 mr-4"
            >{{ $t("common.settings.min_idle_conns") }}</span
          >
          <n-input-number
            v-model:value="config.cachebase.redis.min_idle_conns"
            :min="0"
            :max="1000"
            class="w-35"
            size="small"
          />
        </div>
        <div class="flex items-center justify-between py-2 min-h-9">
          <span
            class="text-13px user-color-ftext shrink-0 mr-4"
            >{{ $t("common.settings.read_timeout") }}</span
          >
          <n-input-number
            v-model:value="cacheRedisReadTimeout"
            :min="1"
            :max="600"
            class="w-35"
            size="small"
          />
        </div>
        <div class="flex items-center justify-between py-2 min-h-9">
          <span
            class="text-13px user-color-ftext shrink-0 mr-4"
            >{{ $t("common.settings.write_timeout") }}</span
          >
          <n-input-number
            v-model:value="cacheRedisWriteTimeout"
            :min="1"
            :max="600"
            class="w-35"
            size="small"
          />
        </div>
        <div class="flex items-center justify-between py-2 min-h-9">
          <span
            class="text-13px user-color-ftext shrink-0 mr-4"
            >{{ $t("common.settings.dial_timeout") }}</span
          >
          <n-input-number
            v-model:value="cacheRedisDialTimeout"
            :min="1"
            :max="600"
            class="w-35"
            size="small"
          />
        </div>
      </template>
    </div>

    <!-- Queuebase -->
    <div class="mb-6">
      <div
        class="text-14px font-600 user-color-ftext mb-3 pb-1.5 border-b user-color-line"
      >
        {{ $t("common.settings.queue") }}
      </div>
      <div class="flex items-center justify-between py-2 min-h-9">
        <span
          class="text-13px user-color-ftext shrink-0 mr-4"
          >{{ $t("common.settings.driver") }}</span
        >
        <n-select
          v-model:value="config.mqbase.driver"
          :options="queueDriverOptions"
          class="w-40"
          size="small"
        />
      </div>
      <template v-if="config.mqbase?.driver === 'mem'">
        <div class="flex items-center justify-between py-2 min-h-9">
          <span
            class="text-13px user-color-ftext shrink-0 mr-4"
            >{{ $t("common.settings.mem_capacity") }}</span
          >
          <n-input-number
            v-model:value="config.mqbase.mem.capacity"
            :min="1000"
            :max="10000000"
            class="w-40"
            size="small"
          />
        </div>
      </template>
      <template v-if="config.mqbase?.driver === 'rocket'">
        <div class="flex items-center justify-between py-2 min-h-9">
          <span
            class="text-13px user-color-ftext shrink-0 mr-4"
            >{{ $t("common.settings.rocket_addrs") }}</span
          >
          <n-input v-model:value="rocketAddrs" class="w-55" size="small" />
        </div>
        <div class="flex items-center justify-between py-2 min-h-9">
          <span
            class="text-13px user-color-ftext shrink-0 mr-4"
            >{{ $t("common.settings.rocket_access_key") }}</span
          >
          <n-input
            v-model:value="config.mqbase.rocket.access_key"
            class="w-45"
            size="small"
          />
        </div>
        <div class="flex items-center justify-between py-2 min-h-9">
          <span
            class="text-13px user-color-ftext shrink-0 mr-4"
            >{{ $t("common.settings.rocket_secret") }}</span
          >
          <n-input
            v-model:value="config.mqbase.rocket.secret"
            type="password"
            show-password-on="click"
            class="w-45"
            size="small"
          />
        </div>
        <div class="flex items-center justify-between py-2 min-h-9">
          <span
            class="text-13px user-color-ftext shrink-0 mr-4"
            >{{ $t("common.settings.rocket_topic") }}</span
          >
          <n-input
            v-model:value="config.mqbase.rocket.topic"
            class="w-40"
            size="small"
          />
        </div>
      </template>
      <template v-if="config.mqbase?.driver === 'redis'">
        <div class="flex items-center justify-between py-2 min-h-9">
          <span
            class="text-13px user-color-ftext shrink-0 mr-4"
            >{{ $t("common.settings.redis_addr") }}</span
          >
          <n-input
            v-model:value="config.mqbase.redis.addr"
            class="w-55"
            size="small"
          />
        </div>
        <div class="flex items-center justify-between py-2 min-h-9">
          <span
            class="text-13px user-color-ftext shrink-0 mr-4"
            >{{ $t("common.settings.redis_password") }}</span
          >
          <n-input
            v-model:value="config.mqbase.redis.password"
            type="password"
            show-password-on="click"
            class="w-45"
            size="small"
          />
        </div>
        <div class="flex items-center justify-between py-2 min-h-9">
          <span
            class="text-13px user-color-ftext shrink-0 mr-4"
            >{{ $t("common.settings.redis_db") }}</span
          >
          <n-input-number
            v-model:value="config.mqbase.redis.db"
            :min="0"
            :max="15"
            class="w-35"
            size="small"
          />
        </div>
        <div class="flex items-center justify-between py-2 min-h-9">
          <span
            class="text-13px user-color-ftext shrink-0 mr-4"
            >{{ $t("common.settings.redis_pool_size") }}</span
          >
          <n-input-number
            v-model:value="config.mqbase.redis.pool_size"
            :min="1"
            :max="1000"
            class="w-35"
            size="small"
          />
        </div>
        <div class="flex items-center justify-between py-2 min-h-9">
          <span
            class="text-13px user-color-ftext shrink-0 mr-4"
            >{{ $t("common.settings.redis_topic") }}</span
          >
          <n-input
            v-model:value="config.mqbase.redis.topic"
            class="w-40"
            size="small"
          />
        </div>
      </template>
    </div>

    <!-- 保存按钮 -->
    <div
      class="pt-3 border-t user-color-line flex justify-end"
    >
      <n-button
        type="primary"
        :loading="saving"
        @click="confirmSave"
        size="medium"
      >
        {{ $t("common.all.save") }}
      </n-button>
    </div>
  </div>
</template>

