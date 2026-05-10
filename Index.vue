<script setup>
import { RefreshCcw,Trash } from 'lucide-vue-next';
import { ref,h,onMounted } from "vue";
import { useMessage,NDivider,NDataTable,NTag,NButton,NSpace,NAlert,NSwitch,useModal,NIcon,NSelect } from 'naive-ui'
import { ApiEnableGetBattleReport,ApiDisableGetBattleReport } from '../api';
import { GetNetData,ClearNetData } from '../../wailsjs/go/main/App'
import { cmdID } from '../cmdID';
const isLogin = ref(true);
const nmessage = useMessage()
const EnableGetBattleReport = () =>{
    ApiEnableGetBattleReport().then(v => {
        if(v.data.code == 200){
            nmessage.success("开启成功")
        }else{
            nmessage.error(v.data.msg);
        }
    }).catch(e => {
        nmessage.error("开启获取战报详情失败:"+e);
    });
}

const DisableGetBattleReport = () =>{
    ApiDisableGetBattleReport().then(v => {
        if(v.data.code == 200){
            nmessage.success("关闭成功")
        }else{
            nmessage.error(v.data.msg);
        }
    }).catch(e => {
        nmessage.error("关闭获取战报详情失败:"+e);
    });
}


function createColumns({
  sendMail
}) {
  return [
//    {
//       type: "expand",
//     //   expandable: (rowData) => rowData.name !== "Jim Green",
//       renderExpand: (rowData) => {
//         // return `${rowData.data}`;
//         return h(
//             NAlert,
//             {
//               style: {
//                 marginRight: "6px"
//               },
//             //   type: "info",
//             "show-icon": false
//             },
//             {
//               default: () => rowData.data
//             }
//           );
//       }
//     },
    {
      title: "ID",
      key: "key",
      sorter: (row1, row2) => row1.time - row2.time,
      defaultSortOrder: 'descend',
      render: (row) => {
        return row.id;
      }
    },
    {
      title: "类型",
      key: "type",
      width: 60,
      render: (row) => {
        if(row.type == 0){
            return h(
                NTag,
                {
                    size: "small",
                    type: "info",
                },
                {
                default: () => "接收"
                }
            );
        }else if(row.type == 1){
            return h(
                NTag,
                {
                    size: "small",
                    type: "success",
                },
                {
                default: () => "发送"
                }
            );
        }
      }
    },
    {
      title: "协议ID",
      key: "cmd_id",
      render: (rowData) => {
        return h(
            NTag,
            {
                type: "info",
                size: "small"
            },
            {
              default: () => rowData.cmd_id+"("+(cmdID[rowData.cmd_id] ?? "未知")+")"
            }
        );
      },      
    },
    {
      title: "包长度",
      key: "bufsize",
      width: 150,
    },
    {
      title: "数据类型",
      key: "datatype",
      width: 80,
      render: (row) => {
        if(row.datatype == 2){
            return h(
                NTag,
                {
                    size: "small"
                },
                {
                default: () => "text(2)"
                }
            );
        }else if(row.datatype == 3){
            return h(
                NTag,
                {
                    size: "small"
                },
                {
                default: () => "zlib(3)"
                }
            );
        }else if(row.datatype == 5){
            return h(
                NTag,
                {
                    size: "small"
                },
                {
                default: () => "xor(5)"
                }
            );
        }
      }
    },
    {
      title: "来源",
      key: "src",
      // render: (row, index) => {
      //   return new Date(row.time).toLocaleString('zh-CN');
      // }
    },
    {
      title: "目标",
      key: "dst",
      // render: (row, index) => {
      //   return new Date(row.time).toLocaleString('zh-CN');
      // }
    },
    {
      title: "时间",
      key: "time",
      render: (row, index) => {
        return new Date(row.time).toLocaleString('zh-CN');
      }
    },
    // {
    //   title: "Tags",
    //   key: "tags",
    //   render(row) {
    //     const tags = row.tags.map((tagKey) => {
    //       return h(
    //         NTag,
    //         {
    //           style: {
    //             marginRight: "6px"
    //           },
    //           type: "info",
    //           bordered: false
    //         },
    //         {
    //           default: () => tagKey
    //         }
    //       );
    //     });
    //     return tags;
    //   }
    // },
    {
      title: "Action",
      key: "actions",
      width: 100,
      render(row) {
        return h(
          NButton,
          {
            size: "small",
            onClick: () => {modal.create({
                title: "ID:"+row.id + "-cmdId:"+row.cmd_id+" 数据详情",
                draggable: true,
                preset: "card",
                style:{ width: '800px' },
                content: h(
                    NAlert,
                    {
                        size: "small",
                        showIcon:false
                    },
                    {
                    default: () => row.data
                    }
                )
            });}
          },
          { default: () => "查看数据" }
        );
      }
    }
  ];
}

function createData() {
    // const resp = await GetNetData();
    // let data = JSON.parse(resp);
    // return data.data;
    GetNetData().then(resp => {
        let data = JSON.parse(resp);
        data.data = data.data.filter(e => !multipleSelectValue.value.includes(e.cmd_id));
        dataRef.value = data.data
        console.log(dataRef.value);
        const uniqueCmdIds = [...new Set(data.data.map(item => item.cmd_id))];
        console.log(uniqueCmdIds);
        options.value = [];
        uniqueCmdIds.forEach(e => {
            options.value.push({
                label: e,
                value: e,
            });
        });
        console.log("multipleSelectValue",multipleSelectValue);
        
    });
}

const multipleSelectValue = ref([]);

// console.log(await createData())
const dataRef = ref([]);
const columns = createColumns({
  sendMail(rowData) {
    message.info(`send mail to ${rowData.name}`);
  }
});

function rowKey(row) {
  return row.id;
}

onMounted(() => {
    handleChange(autor.value);
});

const autor = ref(true);
var autorInterval = false;

function handleChange(value) {
  if(value == true){
    autorInterval = setInterval(createData,1000);
  }else{
    clearInterval(autorInterval);
  }
}

const options = ref([
  {
    label: "Everybody's Got Something to Hide Except Me and My Monkey",
    value: "song0",
    disabled: true
  },
  {
    label: "Drive My Car",
    value: "song1"
  },
  {
    label: "Norwegian Wood",
    value: "song2"
  },
  {
    label: "You Won't See",
    value: "song3",
    disabled: true
  },
  {
    label: "Nowhere Man",
    value: "song4"
  },
  {
    label: "Think For Yourself",
    value: "song5"
  },
  {
    label: "The Word",
    value: "song6"
  },
  {
    label: "Michelle",
    value: "song7",
    disabled: true
  },
  {
    label: "What goes on",
    value: "song8"
  },
  {
    label: "Girl",
    value: "song9"
  },
  {
    label: "I'm looking through you",
    value: "song10"
  },
  {
    label: "In My Life",
    value: "song11"
  },
  {
    label: "Wait",
    value: "song12"
  }
])
const modal = useModal();

const clear = () => {
  ClearNetData().then(v=>{
    createData();
  });
}
</script>

<template>
    <div class="bikamoeapp">
        <div class="bikamoeapp-content">
            <div class="bikamoeapp-title">
                <h2 style="margin-bottom: 4px;">率土之滨网络工具</h2>
                <p>stzbNetTool v0.0.1 Build.2603290005</p>
            </div>
            <div style="padding: 16px;display: flex;">
                <div>
                    <n-button secondary @click="createData">
                        <template #icon>
                            <NIcon size="16">
                                <RefreshCcw />
                            </NIcon>
                        </template>
                        刷新
                    </n-button>
                </div>
                <div style="margin-left: 8px;">
                    <n-button secondary @click="clear">
                        <template #icon>
                            <NIcon size="16">
                                <Trash />
                            </NIcon>
                        </template>
                        清空
                    </n-button>
                </div>
               <div style="margin: 4px;">
                    自动刷新 <n-switch :round="false" v-model:value="autor" @update:value="handleChange" />
               </div>
               
               <!-- <div style="margin: 4px;">
                    自动刷新 <n-switch :round="false" />
               </div> -->
            </div>
            <div style="display: flex;align-items: center;padding: 16px;" >
                    <p style="flex-shrink:0">过滤协议： </p>
                    <n-select
                        v-model:value="multipleSelectValue"
                        filterable
                        multiple
                        tag
                        :options="options"
                        clearable
                        @update:value="()=>{createData()}"
                    />
            </div>
            <n-divider />
            
            <div class="bikamoeapp-list">
                <div style="padding:0 12px;width: 100%;">
                    <n-data-table
                        striped
                        size="small"
                        :row-key="rowKey"
                        :bordered="false"
                        :single-line="false"
                        :columns="columns"
                        :data="dataRef"
                        virtual-scroll
                        :max-height="500"
                    />
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped></style>