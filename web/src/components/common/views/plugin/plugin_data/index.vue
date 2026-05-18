<script setup>
import { NCard, NDataTable, NButton, NInput, NFlex, NSelect } from "naive-ui";
import { ref, h, reactive, computed, onMounted } from "vue";
import { postData, postBlob } from "@/util/util";
import { windowsStore } from "@/stores/windows";
import EditCom from "./edit.vue";
import DetailCom from "./detail.vue";
import ImportCom from "./import.vue";
import { cloneDeep } from "lodash";
import { checkAuth } from "@/directives/auth.js";
const props = defineProps({
  winId: {
    type: String,
    default: "",
  },
});
const winId = props.winId;
const ws = windowsStore();
const loading = ref(false);

const columnFn = () => [
  {
    type: "selection",
  },
  {
    title: $t("model.plugin_data.plugin_name"),
    key: "PluginName",
    sorter: true,
    resizable: true,
    ellipsis: true,
    minWidth: 120,
  },
  {
    title: $t("model.plugin_data.plugin_table"),
    key: "PluginTable",
    sorter: true,
    resizable: true,
    ellipsis: true,
    minWidth: 120,
  },
  {
    title: $t("model.plugin_data.data_id"),
    key: "DataId",
    sorter: true,
    resizable: true,
    ellipsis: true,
    minWidth: 120,
  },
  {
    title: $t("model.plugin_data.user_id"),
    key: "UserId",
    sorter: true,
    resizable: true,
    ellipsis: true,
    minWidth: 120,
  },
  {
    title: $t("model.plugin_data.dept_id"),
    key: "DeptId",
    sorter: true,
    resizable: true,
    ellipsis: true,
    minWidth: 120,
  },
  {
    title: $t("model.plugin_data.field1"),
    key: "Field1",
    sorter: true,
    resizable: true,
    ellipsis: true,
    minWidth: 120,
  },
  {
    title: $t("model.plugin_data.field2"),
    key: "Field2",
    sorter: true,
    resizable: true,
    ellipsis: true,
    minWidth: 120,
  },
  {
    title: $t("model.plugin_data.field3"),
    key: "Field3",
    sorter: true,
    resizable: true,
    ellipsis: true,
    minWidth: 120,
  },
  {
    title: $t("model.plugin_data.field4"),
    key: "Field4",
    sorter: true,
    resizable: true,
    ellipsis: true,
    minWidth: 120,
  },
  {
    title: $t("model.plugin_data.field5"),
    key: "Field5",
    sorter: true,
    resizable: true,
    ellipsis: true,
    minWidth: 120,
  },
  {
    title: $t("model.plugin_data.field6"),
    key: "Field6",
    sorter: true,
    resizable: true,
    ellipsis: true,
    minWidth: 120,
  },
  {
    title: $t("model.plugin_data.field7"),
    key: "Field7",
    sorter: true,
    resizable: true,
    ellipsis: true,
    minWidth: 120,
  },
  {
    title: $t("model.plugin_data.field8"),
    key: "Field8",
    sorter: true,
    resizable: true,
    ellipsis: true,
    minWidth: 120,
  },
  {
    title: $t("model.plugin_data.field9"),
    key: "Field9",
    sorter: true,
    resizable: true,
    ellipsis: true,
    minWidth: 120,
  },
  {
    title: $t("model.plugin_data.field10"),
    key: "Field10",
    sorter: true,
    resizable: true,
    ellipsis: true,
    minWidth: 120,
  },
  {
    title: $t("model.plugin_data.field11"),
    key: "Field11",
    sorter: true,
    resizable: true,
    ellipsis: true,
    minWidth: 120,
  },
  {
    title: $t("model.plugin_data.field12"),
    key: "Field12",
    sorter: true,
    resizable: true,
    ellipsis: true,
    minWidth: 120,
  },
  {
    title: $t("model.plugin_data.field13"),
    key: "Field13",
    sorter: true,
    resizable: true,
    ellipsis: true,
    minWidth: 120,
  },
  {
    title: $t("model.plugin_data.field14"),
    key: "Field14",
    sorter: true,
    resizable: true,
    ellipsis: true,
    minWidth: 120,
  },
  {
    title: $t("model.plugin_data.field15"),
    key: "Field15",
    sorter: true,
    resizable: true,
    ellipsis: true,
    minWidth: 120,
  },
  {
    title: $t("model.plugin_data.field16"),
    key: "Field16",
    sorter: true,
    resizable: true,
    ellipsis: true,
    minWidth: 120,
  },
  {
    title: $t("model.plugin_data.field17"),
    key: "Field17",
    sorter: true,
    resizable: true,
    ellipsis: true,
    minWidth: 120,
  },
  {
    title: $t("model.plugin_data.field18"),
    key: "Field18",
    sorter: true,
    resizable: true,
    ellipsis: true,
    minWidth: 120,
  },
  {
    title: $t("model.plugin_data.field19"),
    key: "Field19",
    sorter: true,
    resizable: true,
    ellipsis: true,
    minWidth: 120,
  },
  {
    title: $t("model.plugin_data.field20"),
    key: "Field20",
    sorter: true,
    resizable: true,
    ellipsis: true,
    minWidth: 120,
  },
  {
    title: $t("model.plugin_data.text_field1"),
    key: "TextField1",
    sorter: true,
    resizable: true,
    ellipsis: true,
    minWidth: 120,
  },
  {
    title: $t("model.plugin_data.text_field2"),
    key: "TextField2",
    sorter: true,
    resizable: true,
    ellipsis: true,
    minWidth: 120,
  },
  {
    title: $t("model.plugin_data.text_field3"),
    key: "TextField3",
    sorter: true,
    resizable: true,
    ellipsis: true,
    minWidth: 120,
  },
  {
    title: $t("model.plugin_data.text_field4"),
    key: "TextField4",
    sorter: true,
    resizable: true,
    ellipsis: true,
    minWidth: 120,
  },
  {
    title: $t("model.plugin_data.text_field5"),
    key: "TextField5",
    sorter: true,
    resizable: true,
    ellipsis: true,
    minWidth: 120,
  },
  {
    title: $t("model.plugin_data.index_field1"),
    key: "IndexField1",
    sorter: true,
    resizable: true,
    ellipsis: true,
    minWidth: 120,
  },
  {
    title: $t("model.plugin_data.index_field2"),
    key: "IndexField2",
    sorter: true,
    resizable: true,
    ellipsis: true,
    minWidth: 120,
  },
  {
    title: $t("model.plugin_data.index_field3"),
    key: "IndexField3",
    sorter: true,
    resizable: true,
    ellipsis: true,
    minWidth: 120,
  },
  {
    title: $t("model.plugin_data.index_field4"),
    key: "IndexField4",
    sorter: true,
    resizable: true,
    ellipsis: true,
    minWidth: 120,
  },
  {
    title: $t("model.plugin_data.index_field5"),
    key: "IndexField5",
    sorter: true,
    resizable: true,
    ellipsis: true,
    minWidth: 120,
  },
  {
    title: $t("model.common.createdat"),
    key: "CreatedAt",
    sorter: true,
    resizable: true,
    ellipsis: true,
    minWidth: 120,
    width: 150,
    render(row) {
      return new Date(row.CreatedAt).format();
    },
  },
  {
    title: $t("model.common.updatedat"),
    key: "UpdatedAt",
    sorter: true,
    resizable: true,
    ellipsis: true,
    minWidth: 120,
    width: 150,
    render(row) {
      return new Date(row.UpdatedAt).format();
    },
  },
  {
    title: $t("common.all.action"),
    key: "actions",
    fixed: "right",
    width: 160,
    render(row) {
      const buttons = [];
      if (checkAuth("/api/plugin_data/save")) {
        buttons.push(
          h(
            NButton,
            {
              size: "tiny",
              type: "primary",
              onClick: () => toEdit(row),
            },
            { default: () => $t("common.all.edit") },
          ),
        );
      }
      if (checkAuth("/api/plugin_data/detail")) {
        buttons.push(
          h(
            NButton,
            {
              class: "m-l-1",
              size: "tiny",
              type: "default",
              onClick: () => toDetail(row),
            },
            { default: () => $t("common.all.detail") },
          ),
        );
      }
      if (checkAuth("/api/plugin_data/delates")) {
        buttons.push(
          h(
            NButton,
            {
              class: "m-l-1",
              size: "tiny",
              type: "error",
              onClick: () => toDelate(row),
            },
            { default: () => $t("common.all.delete") },
          ),
        );
      }
      return buttons;
    },
  },
];
const columns = ref(columnFn());
const pagination = reactive({
  page: 1,
  pageSize: 10,
  showSizePicker: true,
  pageSizes: [10, 50, 100, 500, 1000],
  prefix({ itemCount }) {
    return $t("common.all.itemsCount", { count: itemCount });
  },
});
const selectedKeys = ref([]);
const rowKey = (row) => {
  return row.ID;
};
const handleCheckRow = function (keys, rows) {
  const listKeys = [];
  rows.forEach((item) => {
    if (item && item.ID) {
      listKeys.push(item.ID);
    }
  });
  selectedKeys.value = listKeys;
};
const param = reactive({
  page: 1,
  pageSize: 10,
  keyword: "",
  table: "plugin_data_0",
});
const data = ref([]);
const handleSorterChange = (sorter) => {
  if (!param.sorter) {
    param.sorter = [];
  }
  let has = param.sorter.filter((x) => sorter.columnKey == x.columnKey).shift();
  if (has) {
    has.order = sorter.order;
    has.sorter = sorter.sorter;
  } else {
    param.sorter.push(sorter);
  }
  param.page = pagination.page;
  param.pageSize = pagination.pageSize;
  query();
};
const handleFiltersChange = (filters) => {
  param.filters = !filters ? null : filters;
  param.page = pagination.page;
  param.pageSize = pagination.pageSize;
  query();
};
const handlePageChange = (currentPage) => {
  param.page = currentPage;
  query();
};
const handlePageSizeChange = (pageSize) => {
  param.pageSize = pageSize;
  handlePageChange(1);
};
const search = () => {
  param.page = 1;
  query();
};
const toEdit = async (tmp) => {
  let row = null;
  if (tmp.ID) {
    row = await postData("plugin_data", "detail?table=" + param.table, {
      ID: tmp.ID,
    });
  } else {
    row = cloneDeep(tmp);
  }
  row.Table = param.table;
  ws.addWindow(
    {
      width: 500,
      height: 350,
      title: computed(() => {
        return row.ID
          ? $t("common.all.edit2", { entity: $t("model.plugin_data.model") })
          : $t("common.all.add2", { entity: $t("model.plugin_data.model") });
      }),
      component: EditCom,
      data: row,
    },
    winId,
  );
};
const toDetail = async (row) => {
  const res = await postData("plugin_data", "detail?table=" + param.table, {
    ID: row.ID,
  });
  if (res) {
    ws.addWindow(
      {
        width: 500,
        height: 350,
        title: computed(() => {
          return $t("common.all.detail2", {
            entity: $t("model.plugin_data.model"),
          });
        }),
        component: DetailCom,
        data: res,
      },
      winId,
    );
  }
};
const toDelate = async (row) => {
  const flag = await $msg.util.confirm($t("common.all.deleteOneConfirm"));
  if (!flag) {
    return;
  }
  const res = await postData("plugin_data", "delates?table=" + param.table, [
    row.ID,
  ]);
  if (res) {
    query();
  }
};
const toDelates = async () => {
  const flag = await $msg.util.confirm(
    $t("common.all.deleteManyConfirm", { count: selectedKeys.value.length }),
  );
  if (!flag) {
    return;
  }
  const res = await postData(
    "plugin_data",
    "delates?table=" + param.table,
    selectedKeys.value,
  );
  if (res) {
    query();
  }
};
const importXlsx = () => {
  ws.addWindow(
    {
      width: 400,
      height: 300,
      title: computed(() => {
        return $t("common.all.import");
      }),
      component: ImportCom,
      data: { Table: param.table },
    },
    winId,
  );
};
const exportXlsx = async () => {
  const data = await postBlob("plugin_data", "export", param, "okerr");
  if (data == null) {
    return;
  }
  const a = document.createElement("a");
  a.href = URL.createObjectURL(data.blob);
  a.download = data.name;
  a.click();
};
const query = async () => {
  selectedKeys.value = [];
  loading.value = true;
  pagination.page = param.page;
  pagination.pageSize = param.pageSize;
  const res = await postData("plugin_data", "page", param);
  if (res) {
    data.value = res.data;
    pagination.itemCount = res.total;
  } else {
    data.value = [];
    pagination.itemCount = 0;
  }
  if (param.sorter) {
    param.sorter.forEach((item) => {
      const column = columns.value
        .filter((x) => x.key == item.columnKey)
        .shift();
      if (column) {
        column.sortOrder = item.order;
        column.sorter = true;
      }
    });
  }
  loading.value = false;
};
query();
import { useEventBus } from "@/util/event.js";
useEventBus("plugin_data-edit-over", (msg) => {
  query();
});
useEventBus("lang-change", (msg) => {
  columns.value = columnFn();
});
const tableOptions = reactive([]);
onMounted(async () => {
  const res = await postData("plugin_data", "tables");
  if (res) {
    for (let i = 0; i < res; i++) {
      tableOptions.push({
        label: "Table " + i,
        value: "plugin_data_" + i,
      });
    }
  }
});
</script>
<template>
  <div
    v-if="!checkAuth('/api/plugin_data/page')"
    class="flex items-center justify-center h-full"
  >
    <div>{{ $t("common.all.pagenoauth") }}</div>
  </div>
  <div class="h-full" v-auth="'/api/plugin_data/page'">
    <n-card class="h-full rounded-0" :bordered="false">
      <div class="h-full">
        <n-flex inline>
          <div class="inline-flex">
            <div class="w-16 line-height-7">Table</div>
            <n-select
              v-model:value="param.table"
              class="w-30"
              :options="tableOptions"
              size="small"
            />
          </div>
          <div class="inline-flex">
            <div class="w-32 line-height-7">
              {{ $t("model.plugin_data.plugin_name") }}
            </div>
            <n-input
              v-model:value="param.keyword"
              :placeholder="
                $t('common.all.searchKeyword', {
                  keyword: $t('model.plugin_data.plugin_name'),
                })
              "
              size="small"
            />
          </div>
          <n-button type="primary" size="small" @click="search()">
            {{ $t("common.all.search") }}
          </n-button>
          <n-button
            v-auth="'/api/plugin_data/save'"
            type="primary"
            size="small"
            @click="toEdit({})"
          >
            {{ $t("common.all.add") }}
          </n-button>
          <n-button
            v-auth="'/api/plugin_data/delates'"
            type="error"
            size="small"
            v-if="selectedKeys.length > 0"
            @click="toDelates()"
          >
            {{ $t("common.all.delete") }}({{ selectedKeys.length }})
          </n-button>
          <n-button
            v-auth="'/api/plugin_data/import'"
            type="default"
            size="small"
            @click="importXlsx()"
          >
            {{ $t("common.all.import") }}
          </n-button>
          <n-button
            v-auth="'/api/plugin_data/export'"
            type="primary"
            size="small"
            @click="exportXlsx()"
          >
            {{ $t("common.all.export") }}
          </n-button>
        </n-flex>
        <n-data-table
          :scroll-x="columns.length * 120"
          remote
          :columns="columns"
          :data="data"
          :loading="loading"
          :pagination="pagination"
          :row-key="rowKey"
          @update:sorter="handleSorterChange"
          @update:filters="handleFiltersChange"
          @update:page="handlePageChange"
          @update:page-size="handlePageSizeChange"
          :single-line="false"
          @update:checked-row-keys="handleCheckRow"
          flex-height
          class="h-[calc(100%-30px)] m-t-2"
          striped
          virtual-scroll
          size="small"
        />
      </div>
    </n-card>
  </div>
</template>
