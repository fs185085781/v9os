import { NIcon } from "naive-ui";
import { h} from "vue";

const bgColors = [
  '#52c41a', // 绿
  '#faad14', // 金
  '#1890ff', // 蓝
  '#f5222d', // 红
  '#722ed1', // 紫
  '#13c2c2', // 青
  '#eb2f96', // 粉
  '#fa541c', // 橙红
  '#2f54eb', // 深蓝
  '#a0d911', // 亮绿
  '#fa8c16', // 橙
  '#bfbfbf'  // 灰
]

const recentIndexes = []
function pickColor() {
  let index
  let retry = 0

  do {
    index = Math.floor(Math.random() * bgColors.length)
    retry++
  } while (
    recentIndexes.includes(index) &&
    retry < 50
  )

  // 更新最近记录
  recentIndexes.push(index)
  if (recentIndexes.length > 3) {
    recentIndexes.shift()
  }

  return bgColors[index]
}
export function renderIcon(icon, size = 28) {
  const backgroundColor = pickColor();
  return (props = {}) => {
    const iconSize = Number(props.size || size || 28);
    const borderSize = iconSize / 4;
    const borderRadius = window.$user && window.$user.settings.Round == "true" ? (borderSize+'px') : '0px'
    return h(
      'div',
      {
        style: {
          width: iconSize+'px',
          height: iconSize+'px',
          borderRadius,
          backgroundColor,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center'
        }
      },
      [
        h(NIcon, {
          size: Math.max(14, Math.round(iconSize * 0.64)),
          color: '#fff'
        }, {
          default: () => h(icon)
        })
      ]
    )
  }
}
