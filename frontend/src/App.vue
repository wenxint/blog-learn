<script setup>
// App.vue 只做两件事：设置组件库语言 + 渲染当前路由。
//
// 用 el-config-provider 而不是 app.use(ElementPlus, { locale })：
// 按需引入模式下没有「全局安装」这一步了，语言包通过 Provider 组件下发，
// 分页器的「共 x 条」「上一页」才会显示中文。
import zhCn from 'element-plus/es/locale/lang/zh-cn'
</script>

<template>
  <el-config-provider :locale="zhCn">
    <RouterView />
  </el-config-provider>
</template>
