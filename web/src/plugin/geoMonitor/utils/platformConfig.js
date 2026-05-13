// 预设平台配置：code 与 utils SDK ——对应，新增平台必须从下拉中选择
export const platformOptions = [
  {
    code: 'deepseek',
    name: 'DeepSeek',
    apiBase: 'https://api.deepseek.com',
    webBase: 'https://chat.deepseek.com',
    icon: 'simple-icons:deepseek',
    color: '#4D6BFE'
  },
  {
    code: 'qwen',
    name: '通义千问',
    apiBase: 'https://dashscope.aliyuncs.com/compatible-mode/v1',
    webBase: 'https://www.qianwen.com/',
    icon: 'simple-icons:alibabacloud',
    color: '#FF6A00'
  },
  {
    code: 'zhipu',
    name: '智谱GLM',
    apiBase: 'https://open.bigmodel.cn/api/paas/v4',
    webBase: 'https://chatglm.cn',
    icon: 'mdi:brain',
    color: '#6C5CE7'
  },
  {
    code: 'doubao',
    name: '豆包',
    apiBase: 'https://ark.cn-beijing.volces.com/api/v3',
    webBase: 'https://www.doubao.com/chat/',
    icon: 'simple-icons:bytedance',
    color: '#3256A8'
  },
  {
    code: 'kimi',
    name: 'Kimi',
    apiBase: 'https://api.moonshot.cn',
    webBase: 'https://www.kimi.com/',
    icon: 'mdi:moon-waning-crescent',
    color: '#8B5CF6'
  },
  {
    code: 'wenxin',
    name: '文心一言',
    apiBase: 'https://qianfan.baidubce.com/v2',
    webBase: 'https://yiyan.baidu.com',
    icon: 'simple-icons:baidu',
    color: '#2932E1'
  },
  {
    code: 'yuanbao',
    name: '元宝',
    apiBase: 'https://hunyuan.tencentcloudapi.com',
    webBase: 'https://yuanbao.tencent.com',
    icon: 'simple-icons:wechat',
    color: '#07C160'
  }
]

export function getPlatformByCode(code) {
  return platformOptions.find((p) => p.code === code)
}
